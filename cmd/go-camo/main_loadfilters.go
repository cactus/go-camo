// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cactus/go-camo/v2/pkg/camo"
	"github.com/cactus/go-camo/v2/pkg/htrie"
	"github.com/cactus/mlog"
)

func loadFilterList(fname string) ([]camo.FilterFunc, error) {
	// #nosec
	file, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("could not open filter-ruleset file: %s", err)
	}
	// #nosec
	defer file.Close()

	allowFilter := htrie.NewURLMatcher()
	denyFilter := htrie.NewURLMatcher()
	hasAllow := false
	hasDeny := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "allow|") {
			line = strings.TrimPrefix(line, "allow")
			err = allowFilter.AddRule(line)
			if err != nil {
				break
			}
			hasAllow = true
		} else if strings.HasPrefix(line, "deny|") {
			line = strings.TrimPrefix(line, "deny")
			err = denyFilter.AddRule(line)
			if err != nil {
				break
			}
			hasDeny = true
		} else {
			fmt.Println("ignoring line: ", line)
		}

		err = scanner.Err()
		if err != nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error building filter ruleset: %s", err)
	}

	// append in order. allow first, then deny filters.
	// first false value aborts the request.
	filterFuncs := make([]camo.FilterFunc, 0)

	if hasAllow {
		filterFuncs = append(filterFuncs, allowFilter.CheckURL)
	}

	// denyFilter returns true on a match. we want to invert this for a deny rule, so
	// any deny rule match should return true, and anything _not_ matching should return false
	// so just wrap and invert the bool.
	if hasDeny {
		denyF := func(u *url.URL) (bool, error) {
			chk, err := denyFilter.CheckURL(u)
			return !chk, err
		}
		filterFuncs = append(filterFuncs, denyF)
	}

	if hasAllow && hasDeny {
		mlog.Print(
			strings.Join(
				[]string{
					"Warning! Allow and Deny rules both supplied.",
					"Having Allow rules means anything not matching an allow rule is denied.",
					"THEN deny rules are evaluated. Be sure this is what you want!",
				},
				" ",
			),
		)
	}

	return filterFuncs, nil
}
