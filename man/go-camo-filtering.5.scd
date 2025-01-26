go-camo-filtering(5)

# NAME

go-camo-filtering - Go-camo filtering information

# DESCRIPTION

_go-camo_(1), an implementation of Camo in Go, accepts a filter for filtering as
part of the *--filter-ruleset* argument. ++
This document describes that syntax.

# MATCH_RULESETS

Each line of the ruleset file must adhere to the following format:

```
<RULE_TYPE>|<DOMAIN_COMPONENT|<URL_COMPONENT>
```

The components are as follows:

|[ *RULE_TYPE*
:[ _required_
:< See _RULE_TYPE_ for more information.
|  *DOMAIN_COMPONENT*
:  _required_
:  See _DOMAIN_COMPONENT_ for more information.
|  *URL_COMPONENT*
:  _optional_
:  See _URL_COMPONENT_ for more information.

# RULE_TYPE

The *RULE_TYPE* component defines the filter type.

The supported types are:

*allow*
	This defines an allow filter. If any allow filters are defined, they take
	precedence, and any request *NOT* matching this rule will be rejected.

*deny*
	This is a deny filter. Any request matching this this rule will be rejected.

# DOMAIN_COMPONENT

The domain component has the following format:

```
<DOMAIN_FLAGS>|<DOMAIN_MATCH_RULE>
```

The components are as follows:

*DOMAIN_FLAGS*: _optional_

	Valid Flags

	*s*
		This means "include subdomains". This is similar to a \*.example.com
		match except it also matches the base domain.

*DOMAIN_MATCH_RULE*: _required_

	The domain match rule is the glob match string for domain matching. A single
	glob character may be used, to match to the left. Think of this as matching
	any subdomains. Partial component glob matches are not currently supported
	for domain match rules. Domain matches are *always* case insensitive.

	As an example, the following would match example.com as well as any
	subdomain, as the *s* domain-flag is specified:

	```
	s|example.com
	```

	While the following here would only match sudomains of example.com, and
	would not match the base domain itself:

	```
	|*.example.com
	```

# URL_COMPONENT

The url component has the following format:

```
<URL_FLAG>|<URL_MATCH_RULE>
```

The components are as follows:

*URL_FLAG*: _optional_

	Valid Flags
	
	*i*
		This means the url component is to be compared case insensitively.
		Please note that case insensitive comparisons are a bit slower.
		Benchmark for large url lists.

*URL_MATCH_RULE*: _required_

	The url match rule is the glob match string for matching against the url
	path. Glob characters match to the right.

	As an example, this would match /foo/file.png as well as /fOo/FiLe.PnG:

	```
	i|/foo/file.png
	```

	This would match any path ending in /file.png:

	```
	|*/file.png
	```

# FULL_EXAMPLES

Here are some examples of full rulesets. ++
Match example.com and any subdomains, case insensitive url prefix match.

```
deny|s|example.com|i|/some/subdir/*
```

Match any domain, with a case sensitive url suffix match:

```
deny||*||*/somebadfile.png
```

Reject everything from bad.example.net, including subdomains (such as
foo.bad.example.net and bar.bad.example.net):

```
deny|s|bad.example.net||
```

# INVALID_EXAMPLES

These are NOT valid, as globs for domains need to break on subdomain
boundaries:

```
deny||ex*ample.com||*
deny||*example.com||*
deny||example*.com||*
```

# IDNA_NOTES

Any idna domains are internally converted to ascii/punycode and matched in that
format. This should make it safe to include unicode domains and have it match
either incoming format.

Thus the following should match both bücher.example.com, as well as
xn--bcher-kva.example.com.

```
deny||bücher.example.com||*
```

# OTHER_NOTES

- Case insensitive components are stored twice in the tree, one for each
  character case. This can make for large trees.
- Domains are always compared case insensitively (by lowercasing on input)

# WEBSITE

https://github.com/cactus/go-camo