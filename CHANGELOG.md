# HEAD

- Support for ECS quota detection with automaxprocs  
  ref: [\#83](https://github.com/cactus/go-camo/pull/83)

- Update depencencies

# v2.6.3 2025-04-01

- Build with go-1.24.2

- Update depencencies

# v2.6.2 2025-02-22

- Switch man-page generate to scdoc.

- Update dependencies

- Build with go-1.24

# v2.6.1 2024-12-19

- Update dependencies

# v2.6.0 2024-08-31

- Start building with Go-1.23.0

- Add CLI flag to expose ReadTimeout for tuning (frontend).  
  Default is 30s (same as previosly).

- Add CLI flag for configuring IdleTimeout, the maximum amount of time
  to wait for the next request when keep-alive is enabled (frontend).  
  Default is 30 seconds.

  > [!Note]
  > Previously this value was not set, so it defaulted to the zero value,
  > which resulted in the ReadTimeout value (default 30s) being used
  > internally by the go http server. The new default mirrors this
  > previous old default.

# v2.5.1 2024-07-02

- Rebuild release with Go-1.22.5

# v2.5.0 2024-06-27

- Implement optional redirect when max-size is exceeded  
  ref: [\#80](https://github.com/cactus/go-camo/pull/80)

# v2.4.13 2024-04-22

- Release tagged for arm64 docker building only.

# v2.4.12 2024-04-20

- Update library dependencies.

- Fix docker and github packages publishing issue.

# v2.4.11 2024-04-03

- Update library dependencies.

- Build with Go-1.22.2

# v2.4.10 2024-03-17

- Update library dependencies.

# v2.4.9 2024-02-16

- Minimum Go version now 1.21 due to quic-go dependency, due to better
  cryto/tls support for QUIC in Go-1.21.

- Update library dependencies.

# v2.4.8 2023-12-19

- Add `--automaxprocs` flag to set GOMAXPROCS automatically to match
  Linux container CPU quota/limits.

- Update library dependencies.

# v2.4.7 - 2023-11-13

- Add http3/quic server support. New flag `--quic`. Requires
  `--ssl-listen`.

# v2.4.6 - 2023-10-25

- Add `--no-debug-vars` flag to disable /debug/vars when `--metrics` is
  enabled. (#66, \#67)

# v2.4.5 - 2023-10-23

- fix htrie matching of non punycode (eg. unicode) idna hostnames

- slightly faster logging (update to mlog dependency)

- address a logging issue with missing url path output in
  `"built outgoing request"` debug log

- moderate improve performance of hostname rule processing (approx
  12-30% in microbenchmarks)

- slight improvement in request path url processing (approx 2-4% in
  microbenchmarks)

- fix /debug/vars being enabled by default (#65) due to expvars import
  side effect

# v2.4.4 - 2023-07-25

- update dependencies

- bump version in go.mod (and fix all internal module references)  
  ref: discussion
  [\#62](https://github.com/cactus/go-camo/discussions/62)

# v2.4.3 - 2023-02-18

- update library dependency golang.org/x/net.  
  refs:
  [dependabot-3](https://github.com/cactus/go-camo/security/dependabot/3),
  [dependabot-4](https://github.com/cactus/go-camo/security/dependabot/4)

# v2.4.2 - 2023-02-16

- update library dependency prometheus, covering CVE-2022-21698.  
  Note that for go-camo, the issue in the prometheus library was
  exploitable only when the metrics option/flag (--metrics) is enabled.

- build with go1.19.5

# v2.4.1 - 2022-09-28

- Rebuild release with go-1.19.1

# v2.4.0 - 2022-01-30

- Add support for internal address proxies (HTTP(S)\_PROXY).  
  issue \#55

- Add some additional documentation/caveats regarding HTTP(S)\_PROXY
  usage.

# v2.3.0 - 2021-10-20

- Add support for listening on a unix socket.

- A more graceful shutdown process, under SIGINT or SIGTERM.

# v2.2.2 - 2021-08-21

- Change test only helper library

- structure logging (json) support for http requests (debug logging).  
  PR \#52

# v2.2.1 - 2021-03-25

- Update some dependencies.

# v2.2.0 - 2021-01-10

- Move ip filtering to Dialer.Control, to further improve SSRF
  protections.  
  Note: Added a few additional debug log messages emitted when ip
  filtering

# v2.1.5 - 2020-08-11

- Rebuild release with go-1.15

# v2.1.4 - 2020-02-18

- Rebuild release with go-1.13.8

- Experimental windows build

# v2.1.3 - 2019-12-18

- Rebuild release with go-1.13.5

# v2.1.2 - 2019-11-17

- Fix for enabling metrics collection in proxy

# v2.1.1 - 2019-11-12

- Security fixes / content-type validation

- Add `ProxyFromEnvironment` support. This uses HTTP proxies directed by
  the `HTTP_PROXY` and `NO_PROXY` (or `http_proxy` and `no_proxy`)
  environment variables. See
  <https://golang.org/pkg/net/http/#ProxyFromEnvironment> for more info.

# v2.1.0 - 2019-10-02

- Support `audio/*` with `--allow-content-audio` flag (similar to how
  video is handled)

- Additional metrics datapoints when using `--metrics`

- Support only go 1.13, due to use of new error wrapping semantics

- Improve client connection early abort handling

- Improve max response side handling — only read MaxSize KB from any
  upstream server. Note: This may result in partial responses to clients
  for chunked encoding requests that are longer than MaxSize, as there
  is no way to signal the client other than closing the connection.

- Change default of `--max-size` to 0, as previously chunked encoding
  responses bypassed size restrictions (only content-length was
  previously enforced). To avoid unexpected failures (preserve backwards
  compatibility in this regard), set max-size to 0 by default moving
  forward. Previous default was 5mb (use `--max-size=5120` to set to
  previous default).

# v2.0.1 - 2019-09-12

- Slightly optimize some structure layouts to reduce memory overhead.

- Switch htrie node map from uint8 to uint32, due to go map
  optimizations. See commit bbf7b9ffee83 for more info.

- Update man page generation (makefile) to use asciidoctor. Not only is
  this easier to maintain, but it has the nice property of being
  rendered on github.

# v2.0.0 - 2019-09-08

- Remove `--allow-list` flag, and replace with a unified filtering flag
  `filter-ruleset`. See
  [go-camo-filtering(5)](man/go-camo-filtering.5.adoc) for more
  information on the accepted syntax.

- Update man pages.

- Refactor some internals (remove some regex in favor of a trie like
  data structure for some comparisons)

# v1.1.7 - 2019-08-14

- Remove old stats flag, endpoint, and feature, in favor of the new
  Prometheus endpoint. Good amount of code removal as well.

- Use a sync.Pool \[\]byte buffer for io.CopyBuffer (instead of
  io.Copy). It should reduce some small amount of GC pressure (a bit
  less garbage).

# v1.1.6 - 2019-07-26

- Support range requests to get safari video support working (#36)

# v1.1.5 - 2019-07-23

- Security fixes / SSRF

  - Fix: Ensure non-GET/HEAD request does not send outbound request
    (#35)

  - Fix: Validate redirect urls the same as initial urls (#35)

- Split out exception for missing content types (#32)

- Prometheus compatible metrics endpoint added (#34)

- Disabled credential/userinfo (`user:pass@` style) type urls by
  default. Added cli flag (`--allow-credential-urls`) to retain prior
  behavior (which allows them).

# v1.1.4 - 2019-02-26

- disable passing/generating x-forwarded-for header by default

- add new `--enable-xfwd4` flag to enable x-forwarded-for header
  passing/generation

- add optional json output for stats

- remove gomaxprocs code, as it is no longer necessary

- documentation fixes (man page update, spelling, etc)

- build release with go-1.12

# v1.1.3 - 2018-09-15

- switch to go-1.11 w/GO111MODULE support.  
  this makes building outside GOPATH easy.  
  Looks like heroku supports it now too? (heroku-buildpack-go issue
  \#249)

- build release with go-1.11

- fix ipv6 length comparison

# v1.1.2 - 2018-07-30

- fix SSRF leak, where certain requests would not match defined and
  custom ip deny-lists as expected

# v1.1.1 - 2018-07-18

- change `/healthcheck` response to 200 instead of 204.  
  solves configuration issue with some loadbalancers.

# v1.1.0 - 2018-07-16

- add flag to allow `video/*` as content type (disabled by default)

- allow setting custom server name

- add flag to expose the current version version in http response header
  (similar to how it is done for `-V` cli output)

- change root route to return 404

- add `/healthcheck` route that returns 204 status (no body content)
  useful for load balancers to check that service is running

# v1.0.18 - 2018-05-15

- change repo layout and build pipeline to dep/gox/GOPATH style

- lint fixes and minor struct alignment changes (minor optimization)

- update mlog dependency

- build with go-1.10.2

# v1.0.17 - 2018-01-25

- update dependency versions to current

- include deps in tree (ease build for heroku)

- minor makefile cleanup

- rebuild with go-1.9.3

# v1.0.16 - 2017-08-29

- rebuild with go-1.9

# v1.0.15 - 2017-02-18

- rebuild with go-1.8

- strip binaries as part of default build

# v1.0.14 - 2017-02-15

- Pass through ETag header from server. The previous omission was
  inconsistent with passing the if-none-match client request header.

# v1.0.13 - 2017-01-22

- resolve potential resource leak with redirection failures and http
  response body closing

# v1.0.12 - 2017-01-16

- better address rejection logic

# v1.0.11 - 2017-01-16

- resolve hostname and check against rfc1918 (best effort blocking of
  dns rebind attacks)

- fix regex match bug with 172.16.0.0/12 addresses (over eager match)

# v1.0.10 - 2017-01-03

- apply a more friendly default content-security-policy

# v1.0.9 - 2016-11-27

- just a rebuild of 1.0.8 with go 1.7.3

# v1.0.8 - 2016-08-20

- update go version support

- build release with go1.7

# v1.0.7 - 2016-04-18

- conver to different logging mechanism (mlog)

- fix a go16 logging issue

- add --no-log-ts command line option

# v1.0.6 - 2016-04-07

- use sync/atomic for internal stats counters

- reorganize some struct memory layout a little

- add -VV license info output

- move simple-server to its own repo

- more performant stats (replaced mutex with sync/atomic)

- fewer spawned goroutines when using stats

# v1.0.5 - 2016-02-18

- Build release with go1.6

- Switch to building with gb

# v1.0.4 - 2015-08-28

- Minor change for go1.5 with proxy timeout 504

# v1.0.3 - 2015-04-25

- revert to stdlib http client

# v1.0.2 - 2015-03-08

- fix issue with http date header generation

# v1.0.1 - 2014-12-16

- optimization for allow-list checks

- keepalive options fix

# v1.0.0 - 2014-06-22

- minor code organization changes

- fix for heroku build issue with example code

# v0.6.0 - 2014-06-13

- use simple router instead of gorilla/mux to reduce overhead and
  increase performance.

- move some code from camo proxy into the simple router

# v0.5.0 - 2014-06-02

- some minor changes to Makefile/building

- add support for HTTP HEAD requests

- add support for adding custom default response headers

- return custom headers on 404s as well.

- enable http keepalives on upstream/backends

- add support for disable http keepalives on frontend/backend separately

- upgrade library deps

- various bug fixes

# v0.4.0 - 2014-05-23

- remove config support (use env or cli flags)

- turn allowlist into a cli flag to parse a plain text file vs json
  config

- clean ups/general code hygiene

# v0.3.0 - 2014-05-13

- Transparent base64 url support

# v0.2.0 - 2014-04-17

- Remove NoFollowRedirects and add MaxRedirects

- Use <https://github.com/mreiferson/go-httpclient> to handle timeouts
  more cleanly

# v0.1.3 - 2013-06-24

- fix bug in loop prevention

- bump max idle conn count

- keep idle conn trimmer running

# v0.1.2 - 2013-03-30

- Add ReadTimeout to http.Server, to close excessive keepalive
  goroutines

# v0.1.1 - 2013-02-27

- optimize date header generation to use a ticker

- performance improvements

- fix a few subtle race conditions with stats

# v0.1.0 - 2013-01-19

- Refactor logging a bit

- Move encoding functionality into a submodule to reduce import size
  (and thus resultant binary size) for url-tool

- Prevent request loop

- Remove custom Denylist support. Filtering should be done on signed url
  generation. rfc1918 filtering retained and internalized so as do
  reduce internal network exposue surface and avoid non-routable effort.

- Inverted redirect boolean. Redirects are now followed by default, and
  the flag `no-follow` was learned.

- Use new flag parsing library for nicer help and cleaner usage.

- Specify a fallback accept header if none is provided by client.

# v0.0.4 - 2012-09-02

- Refactor Stats code out of camoproxy

- Make stats an optional flag in go-camo

- Minor documentation cleanup

- Clean up excessive logging on client initiated broken pipe

# v0.0.3 - 2012-08-05

- organize and clean up code

- make header filters exported

- start filtering response headers

- add default Server name

- fix bug dealing with header filtering logic

- add cli utility to encode/decode urls for testing, etc.

- change repo layout to be friendlier for Go development/building

- timeout flag is now a duration (15s, 10m, 1h, etc)

- X-Forwarded-For support

- Added more info to readme

# v0.0.2 - 2012-07-12

- documentation cleanup

- code reorganization

- some cleanup of command flag text

- logging code simplification

# v0.0.1 - 2012-07-07

- initial release
