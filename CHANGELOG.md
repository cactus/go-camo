Changelog
=========

## HEAD

## v1.1.0 2018-07-16
*   add flag to allow `video/*` as content type (disabled by default)
*   allow setting custom server name
*   add flag to expose the current version version in http response header
    (similar to how it is done for `-V` cli output)
*   change root route to return 404
*   add `/healthcheck` route that returns 204 status (no body content)
    useful for load balancers to check that service is running

## 1.0.18 2018-05-15
*   change repo layout and build pipeline to dep/gox/GOPATH style
*   lint fixes and minor struct alignment changes (minor optimization)
*   update mlog dependency
*   build with go-1.10.2

## 1.0.17 2018-01-25
*   update dependency versions to current
*   include deps in tree (ease build for heroku)
*   minor makefile cleanup
*   rebuild with go-1.9.3

## 1.0.16 2017-08-29
*   rebuild with go-1.9

## 1.0.15 2017-02-18
*   rebuild with go-1.8
*   strip binaries as part of default build

## 1.0.14 2017-02-15
*   Pass through ETag header from server. The previous omission was
    inconsistent with passing the if-none-match client request header.

## 1.0.13 2017-01-22
*   resolve potential resource leak with redirection failures and http response
    body closing

## 1.0.12 2017-01-16
*   better address rejection logic

## 1.0.11 2017-01-16
*   resolve hostname and check against rfc1918 (best effort blocking of dns rebind attacks)
*   fix regex match bug with 172.16.0.0/12 addresses (over eager match)

## 1.0.10 2017-01-03
*   apply a more friendly default content-security-policy

## 1.0.9 2016-11-27
*   just a rebuild of 1.0.8 with go 1.7.3

## 1.0.8 2016-08-20
*   update go version support
*   build release with go1.7

## 1.0.7 2016-04-18
*   conver to different logging mechanism (mlog)
*   fix a go16 logging issue
*   add --no-log-ts command line option

## 1.0.6 2016-04-07
*   use sync/atomic for internal stats counters
*   reorganize some struct memory layout a little
*   add -VV license info output
*   move simple-server to its own repo
*   more performant stats (replaced mutex with sync/atomic)
*   fewer spawned goroutines when using stats

## 1.0.5 2016-02-18
*   Build release with go1.6
*   Switch to building with gb

## 1.0.4 2015-08-28
*   Minor change for go1.5 with proxy timeout 504

## 1.0.3 2015-04-25
*   revert to stdlib http client

## 1.0.2 2015-03-08
*   fix issue with http date header generation

## 1.0.1 2014-12-16
*   optimization for allow-list checks
*   keepalive options fix

## 1.0.0 2014-06-22

*   minor code organization changes
*   fix for heroku build issue with example code

## 0.6.0 2014-06-13

*   use simple router instead of gorilla/mux to reduce overhead
    and increase performance.
*   move some code from camo proxy into the simple router

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
