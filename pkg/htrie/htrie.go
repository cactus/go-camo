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

// URLMatcher is a
type URLMatcher struct {
	subtrees     map[string]*URLMatcher
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
		matches := make([]*URLMatcher, 0, 8)
		return &matches
	},
}

func getURLMatcherSlice() *[]*URLMatcher {
	return matchesPool.Get().(*[]*URLMatcher)
}

func putURLMatcherSlice(s *[]*URLMatcher) {
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

func uniformLower(s, cutset string) string {
	s = strings.TrimSpace(s)
	if len(cutset) > 0 {
		s = strings.Trim(s, cutset)
	}
	s = strings.ToLower(s)
	return s
}

func (dt *URLMatcher) getOrNewSubTree(s string) *URLMatcher {
	subdt, ok := dt.subtrees[s]
	if !ok {
		subdt = &URLMatcher{
			subtrees: make(map[string]*URLMatcher),
			pathPart: s,
		}
		dt.subtrees[s] = subdt

	}
	return subdt
}

// addRulePath adds a url path rule to the matcher node
func (dt *URLMatcher) addPathRule(urlparts string) error {
	if dt.pathChecker == nil {
		dt.pathChecker = NewGlobPathChecker()
	}
	return dt.pathChecker.AddRule(urlparts)
}

func (dt *URLMatcher) parseRule(rule string) ([]string, error) {
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
			index++
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

// AddRule adds a match rule to the URLMatcher node.
func (dt *URLMatcher) AddRule(rule string) error {
	// expected format: |s|example.com|i|/some/subdir/*
	if dt == nil {
		return fmt.Errorf("node is nil")
	}

	if dt.subtrees == nil {
		dt.subtrees = make(map[string]*URLMatcher)
	}

	ruleParts, err := dt.parseRule(rule)
	if err != nil {
		return err
	}

	var (
		hostRuleFlags = ruleParts[0]
		hostRuleMatch = ruleParts[1]
		urlRuleFlags  = ruleParts[2]
		urlRuleMatch  = ruleParts[3]
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

	hostRuleMatch, err = idna.ToASCII(uniformLower(hostRuleMatch, "."))
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
		label = uniformLower(label, "")
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
				err := curdt.addPathRule(pathRule)
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

func (dt *URLMatcher) walkFind(s string) []*URLMatcher {
	// hostname should already be lowercase. avoid work by not doing it.
	// hostname := strings.ToLower(s)
	matches := *getURLMatcherSlice()
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

// CheckURL checks a *url.URL against the URLMatcher.
// If the url matches (a "hit"), it returns true.
// If the url does not match (a "miss"), it return false.
func (dt *URLMatcher) CheckURL(u *url.URL) bool {
	hostname := u.Hostname()
	matches := dt.walkFind(hostname)
	defer putURLMatcherSlice(&matches)

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

// CheckHostname checks a *url.URL's hostname component against
// the URLMatcher. The path component is ignored.
// If the hostname matches, it returns true.
// If the hostname does not match, it returns false.
func (dt *URLMatcher) CheckHostname(u *url.URL) bool {
	matches := dt.walkFind(u.Hostname())
	defer putURLMatcherSlice(&matches)
	return len(matches) > 0
}

// CheckHostnameString checks the supplied hostname (as a string).
// Note: CheckHostnameString requires that the hostname is already escaped,
// sanitized, space trimmed, and lowercased...
// Basically sanitized in a way similar to `(*url.URL).Hostname()`
func (dt *URLMatcher) CheckHostnameString(hostname string) bool {
	matches := dt.walkFind(hostname)
	defer putURLMatcherSlice(&matches)
	return len(matches) > 0
}

// NewURLMatcher returns a new URLMatcher
func NewURLMatcher() *URLMatcher {
	return &URLMatcher{
		subtrees: make(map[string]*URLMatcher),
	}
}

// NewURLMatcherWithRules returns a new URLMatcher initialized with rules.
func NewURLMatcherWithRules(rules []string) (*URLMatcher, error) {
	dt := &URLMatcher{
		subtrees: make(map[string]*URLMatcher),
	}
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			return nil, err
		}
	}
	return dt, nil
}

// MustNewURLMatcherWithRules is like NewURLMatcherWithRules but panics if one
// of the rules is invalid or cannot be parsed.
// It simplifies safe initialization of global variables.
func MustNewURLMatcherWithRules(rules []string) *URLMatcher {
	dt := &URLMatcher{
		subtrees: make(map[string]*URLMatcher),
	}
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			panic(`regexp: URLMatcher.AddRule(` + rule + `): ` + err.Error())
		}
	}
	return dt
}
