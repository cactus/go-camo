# Match Rulesets

## general deny config format:

Here is the general deny format:

```
<rule-type>|<domain-component|<url-component>
```

The components are as follows:

*   `<rule-type>`: required

    See [rule-type](#rule-type) for more information.

*   `<domain-component>`: required

    See [domain-component](#domain-component) for more information.

*   `<url-component>`: optional

    See [url-component](#url-component) for more information.


### rule-type

The `<rule-type>` component defines the filter type. The supported types are:

*   `allow`: This defines an allow filter. If _any_ allow filters are defined,
    they take precedence, and any request _NOT_ matching this rule will
    be rejected.

*   `deny`: This is a deny filter. Any request matching this this rule will be
    rejected.


### domain-component

The domain component has the following format:

```
<domain-flags>|<domain-match-rule>
```

The components are as follows:

*   `<domain-flags>`: optional

    *   `s`: This means "include subdomains". This is similar to a `*.example.com`
        match except it also matches the base domain.
        
*   `<domain-match-rule>`: required

    The domain match rule is the glob match string for domain matching. A single glob
    character may be used, to match _to the left_. Think of this as matching
    any subdomains.
    
    Partial component glob matches are not currently supported for domain match rules. 
    See [Invalid Examples](#invalid-examples).

Some examples:

This would match `example.com` as well as any subdomain, as the `s` domain-flag
is specified:
```
s|example.com
```

While this would _only_ match sudomains of `example.com`, and would _not_ match the 
base domain itself:
```
|*.example.com
```


### url-component

The url component has the following format:
```
<url-flag>|<url-match-rule>
```

The components are as follows:

*   `<url-flag>`: optional

    *   `i`: This means the url component is to be compared case
        insensitively.

        Do note that case insensitive comparisons are a bit slower. Benchmark
        for large url lists.

*   `<url-match-rule>`: required

    The url match rule is the glob match string for matching against the url
    path. Glob characters match _to the right_.

Some examples:

This would match `/foo/file.png` as well as `/fOo/FiLe.PnG`:
```
i|/foo/file.png
```

This would match any path ending in `/file.png`:
```
|*/file.png
```


## Full examples

Here are some examples of full rulesets.

Match `example.com` and any subdomains, case insensitive url prefix match.  
Note: domains matches are always case insensitive!
```
deny|s|example.com|i|/some/subdir/*
```

Match any domain, with a case sensitive url suffix match:
```
deny||*||*/somebadfile.png
```

Reject everything from `bad.example.net`, including subdomains:
```
deny|s|bad.example.net||
```

## Special notes for IDNA

Any idna domains are internally converted to ascii/punycode and matched in
that format. This _should_ make it safe to include unicode domains and have it
match either incoming format.

Thus the following _should_ match both `bücher.example.com`, as well as 
`xn--bcher-kva.example.com`.
```
deny||bücher.example.com||*
```

## Invalid Examples

These are NOT valid, as globs for domains need to break on subdomain
boundaries:
```
deny||ex*ample.com||*
deny||*example.com||*
deny||example*.com||*
```

## Other notes

*   Case insensitive components are stored twice in the tree, one for each
    character case. This can make for large trees.
*   Domains are always compared case insensitively (by lowercasing on input)
