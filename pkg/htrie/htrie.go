// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/idna"
)

// PathChecker is an interface that specifes a CheckPath method, which returns
// true when a path component is a "hit" and false for a "miss".
type PathChecker interface {
	CheckPath(*url.URL) bool
	AddRule(s string) error
}

type DTree struct {
	subtrees     map[string]*DTree
	pathChecker  PathChecker
	isWild       bool
	hasWildChild bool
	canMatch     bool
	hasRules     bool
	pathPart     string // mostly for debugging
}

var matchesPool = sync.Pool{
	New: func() interface{} {
		// starting backing array size of 8
		// that /seems/ like a pretty good initial value, without
		// being too crazy, and has the nice property of being a powler of 2. ;)
		matches := make([]*DTree, 0, 8)
		return &matches
	},
}

func getDTreeSlice() *[]*DTree {
	return matchesPool.Get().(*[]*DTree)
}

func putDTreeSlice(s *[]*DTree) {
	*s = (*s)[0:0]
	matchesPool.Put(s)
}

func reverse(s []string) []string {
	c := len(s) / 2
	for i := 0; i < c; i++ {
		j := len(s) - i - 1
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func uniformify(s string, cutsetLeft string, cutsetRight string, lower bool) string {
	s = strings.TrimSpace(s)
	if len(cutsetLeft) > 0 {
		s = strings.TrimLeft(s, cutsetLeft)
	}
	if len(cutsetRight) > 0 {
		s = strings.TrimRight(s, cutsetRight)
	}
	if lower {
		s = strings.ToLower(s)
	}
	return s
}

func (dt *DTree) getOrNewSubTree(s string) *DTree {
	subdt, ok := dt.subtrees[s]
	if !ok {
		subdt = &DTree{
			subtrees: make(map[string]*DTree),
			pathPart: s,
		}
		dt.subtrees[s] = subdt

	}
	return subdt
}

func (dt *DTree) AddPathRule(urlparts string) error {
	if dt.pathChecker == nil {
		dt.pathChecker = NewGlobPathChecker()
	}
	return dt.pathChecker.AddRule(urlparts)
}

func (dt *DTree) parseRule(rule string) ([]string, error) {
	if strings.Count(rule, "|") > 4 {
		rule = strings.TrimRight(rule, "|")
	}
	if strings.Count(rule, "|") != 4 {
		return nil, fmt.Errorf("Bad rule format: %s", rule)
	}

	ruleset := make([]strings.Builder, 4)
	index := 0
	// start after first `|`
	for _, r := range rule[1:] {
		if r == '|' {
			index += 1
			continue
		}
		ruleset[index].WriteRune(r)
	}
	parts := make([]string, 4)
	for i, sb := range ruleset {
		parts[i] = strings.TrimSpace(sb.String())
	}
	return parts, nil
}

func (dt *DTree) AddRule(rule string) error {
	// expected format: |s|example.com|i|/some/subdir/*
	if dt == nil {
		return fmt.Errorf("node is nil")
	}

	if dt.subtrees == nil {
		dt.subtrees = make(map[string]*DTree)
	}

	ruleParts, err := dt.parseRule(rule)
	if err != nil {
		return err
	}

	var (
		hostRuleFlags string = ruleParts[0]
		hostRuleMatch string = ruleParts[1]
		urlRuleFlags  string = ruleParts[2]
		urlRuleMatch  string = ruleParts[3]
		pathRule      string
		hasRules      bool
	)

	// check for a bare domain match rule. if the rule is a bare domain match rule,
	// then we can avoid any path processing.
	// as an optimization, a rulePart with only a `*` is effectively the same thing,
	// so avoid the path match overhead and compare as if it was a bare domain match.
	if urlRuleMatch == "" || urlRuleMatch == "*" {
		hasRules = false
	} else {
		hasRules = true
		pathRule = "|" + urlRuleFlags + "|" + urlRuleMatch
	}

	prefix := ""
	if strings.HasPrefix(hostRuleMatch, "*.") {
		prefix = "*."
		hostRuleMatch = hostRuleMatch[2:]
	}

	hostRuleMatch, err = idna.ToASCII(uniformify(hostRuleMatch, ".", ".", true))
	if err != nil {
		return err
	}

	hostRuleMatch = prefix + hostRuleMatch

	diswild := false
	if strings.Contains(hostRuleFlags, "s") {
		diswild = true
	}

	domainLabels := strings.Split(hostRuleMatch, ".")
	if len(domainLabels) == 1 && len(domainLabels[0]) == 0 {
		return fmt.Errorf("bad domain format: no domain specified")
	}

	max := len(domainLabels)
	revDomainLabels := reverse(domainLabels)
	curdt := dt
	for i, label := range revDomainLabels {
		label = uniformify(label, "", "", true)
		if len(label) == 0 {
			return fmt.Errorf("bad domain format: empty component")
		}

		if strings.Contains(label, "*") && len(label) > 1 {
			return fmt.Errorf("bad domain format: * cannot be mix matched in domain")
		}

		if label == "*" {
			if i != max-1 {
				return fmt.Errorf("bad domain format: wildcard only allowed at end")
			}

			// small optimization so we know curnode has a wildcard child
			curdt.hasWildChild = true
		}

		curdt = curdt.getOrNewSubTree(label)

		if i == max-1 {
			// hit the end of label
			curdt.canMatch = true
			if hasRules {
				curdt.hasRules = true
				err := curdt.AddPathRule(pathRule)
				if err != nil {
					return err
				}
			} else {
				curdt.hasRules = false
			}
			if diswild || label == "*" {
				curdt.isWild = true
			}
			return nil
		}
	}
	return nil
}

func (dt *DTree) walkFind(s string) []*DTree {
	// hostname should already be lowercase. avoid work by not doing it.
	// hostname := strings.ToLower(s)
	matches := *getDTreeSlice()
	labels := reverse(strings.Split(s, "."))
	plen := len(labels)
	curnode := dt
	// kind of weird ordering, because the root node isn't part of the search
	// space.
	for i, label := range labels {
		if curnode.subtrees == nil || len(curnode.subtrees) == 0 {
			break
		}

		// now check children for continuation
		v, ok := curnode.subtrees[label]
		if !ok {
			// no match, we are done
			break
		}

		curnode = v

		// got a match, and it is a wild type, so add to match list
		if curnode.isWild {
			matches = append(matches, curnode)
		}

		// not at a domain terminus, and there is a wildcard label,
		// so add child to match (if exists)
		if i < plen-1 && curnode.hasWildChild {
			if x, ok := curnode.subtrees["*"]; ok {
				matches = append(matches, x)
			}
		}
		// hit the end, and we can match at this level
		if i == plen-1 && curnode.canMatch {
			matches = append(matches, curnode)
		}
	}
	return matches
}

func (dt *DTree) CheckURL(u *url.URL) bool {
	hostname := u.Hostname()
	matches := dt.walkFind(hostname)
	defer putDTreeSlice(&matches)

	// check for base domain matches first, to avoid path checking if possible
	for _, match := range matches {
		if !match.hasRules {
			return true
		}
	}

	// no luck, so try path rules this time
	for _, match := range matches {
		// anything match.hasRules _shouldn't_ be nil, so this check is
		// likely superfluous...
		if match.pathChecker == nil {
			continue
		}
		if match.pathChecker.CheckPath(u) {
			return true
		}
	}
	return false
}

func (dt *DTree) CheckHostname(u *url.URL) bool {
	matches := dt.walkFind(u.Hostname())
	defer putDTreeSlice(&matches)
	return len(matches) > 0
}

// CheckHostnameString checks the supplied hostname (as a string).
// Note: CheckHostnameString requires that the hostname is already escaped,
// sanitized, space trimmed, and lowercased...
// Basically sanitized in a way similar to `(*url.URL).Hostname()`
func (dt *DTree) CheckHostnameString(hostname string) bool {
	matches := dt.walkFind(hostname)
	defer putDTreeSlice(&matches)
	return len(matches) > 0
}

func NewDTree() *DTree {
	return &DTree{
		subtrees: make(map[string]*DTree),
	}
}

func NewDTreeWithRules(rules []string) (*DTree, error) {
	dt := &DTree{
		subtrees: make(map[string]*DTree),
	}
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			return nil, err
		}
	}
	return dt, nil
}

// MustNewDTreeWithRules is like NewDTreeWithRules but panics if one of the rules
// is invalid or cannot be parsed.
// It simplifies safe initialization of global variables.
func MustNewDTreeWithRules(rules []string) *DTree {
	dt := &DTree{
		subtrees: make(map[string]*DTree),
	}
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			panic(`regexp: DTree.AddRule(` + rule + `): ` + err.Error())
		}
	}
	return dt
}
