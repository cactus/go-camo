Changelog
=========

## Next Release

## 0.5.0 2014-06-02

*   some minor changes to Makefile/building
*   add support for HTTP HEAD requests
*   add support for adding custom default response headers
*   return custom headers on 404s as well.
*   enable http keepalives on upstream/backends
*   add support for disable http keepalives on frontend/backend separately
*   upgrade library deps
*   various bug fixes

## 0.4.0 2014-05-23

*   remove config support (use env or cli flags)
*   turn allowlist into a cli flag to parse a plain text file vs json config
*   clean ups/general code hygiene

## 0.3.0 2014-05-13

*   Transparent base64 url support

## 0.2.0 2014-04-17

*   Remove NoFollowRedirects and add MaxRedirects
*   Use https://github.com/mreiferson/go-httpclient to handle timeouts more
    cleanly

## 0.1.3 2013-06-24

*   fix bug in loop prevention
*   bump max idle conn count
*   keep idle conn trimmer running

## 0.1.2 2013-03-30

*   Add ReadTimeout to http.Server, to close excessive keepalive goroutines

## 0.1.1 2013-02-27

*   optimize date header generation to use a ticker
*   performance improvements
*   fix a few subtle race conditions with stats

## 0.1.0 2013-01-19

*   Refactor logging a bit
*   Move encoding functionality into a submodule to reduce import size (and
    thus resultant binary size) for url-tool
*   Prevent request loop
*   Remove custom Denylist support. Filtering should be done on signed url
    generation. rfc1918 filtering retained and internalized so as do reduce
    internal network exposue surface and avoid non-routable effort.
*   Inverted redirect boolean. Redirects are now followed by default, and 
    the flag `no-follow` was learned.
*   Use new flag parsing library for nicer help and cleaner usage.
*   Specify a fallback accept header if none is provided by client.

## 0.0.4 2012-09-02

*   Refactor Stats code out of camoproxy
*   Make stats an optional flag in go-camo
*   Minor documentation cleanup
*   Clean up excessive logging on client initiated broken pipe

## 0.0.3 2012-08-05

*   organize and clean up code
*   make header filters exported 
*   start filtering response headers
*   add default Server name
*   fix bug dealing with header filtering logic
*   add cli utility to encode/decode urls for testing, etc.
*   change repo layout to be friendlier for Go development/building
*   timeout flag is now a duration (15s, 10m, 1h, etc)
*   X-Forwarded-For support
*   Added more info to readme


## 0.0.2 2012-07-12

*   documentation cleanup
*   code reorganization
*   some cleanup of command flag text
*   logging code simplification


## 0.0.1 2012-07-07

*   initial release
