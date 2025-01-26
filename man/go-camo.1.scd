go-camo(1)

# NAME

go-camo - Go version of Camo server

# SYNOPSIS

*go-camo* [<_OPTION_> [<_OPTION-ARGUMENT_>]]...

# DESCRIPTION

_go-camo_(1) is an implementation of Camo in Go.

Camo is a special type of image proxy that proxies non-secure images over SSL/TLS.++
This prevents mixed content warnings on secure pages.

It works in conjunction with back-end code to rewrite image URLs and sign them
with an HMAC.

# ENVIRONMENT VARS

*GOCAMO_HMAC*
	The HMAC key to use.

*HTTPS_PROXY*
	Configure an outbound proxy that will be used as the proxy URL for HTTPS
	requests.

	Either a complete URL, or a host[:port] (in which case the _http_ scheme is assumed).

*HTTP_PROXY*
	Configure an outbound proxy that will be used as the proxy URL for HTTP
	requests.

	Either a complete URL, or a host[:port] (in which case the _http_ scheme is assumed).

Regarding *Environmen Vars* vs *OPTIONS* precedence: _OPTIONS_, if provided,
override those defined in environment variables.++
For exmaple, if the HMAC key is provided on the command line, it will override
(if present), an HMAC key set in the environment var.

# OPTIONS

*-k*, *--key*=<_HMAC_KEY_>
	The HMAC key to use.

*-H*, *--header*=<_HEADER_>
	Add additional header to each response.

	This option can be used multiple times to add multiple headers.++
	When specified, these headers are set unconditionally on all responses.

	See _ADD_HEADERS_ for more info.

*--listen*=<_ADDRESS:PORT_>
	Address and port to listen to, as a string of _ADDRESS:PORT_.++
	Default: 0.0.0.0:8080

*--socket-listen*=<_SOCKET_PATH_>
	Path for unix domain socket to bind to for HTTP.

*--ssl-listen*=<_ADDRESS:PORT_>
	Address and port to listen via SSL to, as a string of _ADDRESS:PORT_.

*--ssl-key*=<_SSL-KEY-FILE_>
	Path to ssl private key.++
	Default: key.pem

*--ssl-cert*=<_SSL-CERT-FILE_>
	Path to ssl certificate.++
	Default: cert.pem

*--quic*
	Enable http3/quic. Binds to the same port number as _--ssl-listen_ but udp+quic.

*--automaxprocs*
	Set GOMAXPROCS automatically to match Linux container CPU quota/limits.

*--max-size*=<_SIZE_>
	Max response size allowed in KB. Set to 0 to disable size restriction.++
	Default: 0

*--max-size-redirect*=<_URL_>
	URL to redirect to when max-size is exceeded.

*--timeout*=<_TIME_>
	Timeout value for upstream response. Format is "4s" where s means seconds (backend).++
	Default: 4s

*--idletimeout*=<_TIME_>
	Maximum amount of time to wait for the next request when keep-alive is enabled (frontend).++
	Default: 60s

*--readtimeout*=<_TIME_>
	Maximum duration for reading the entire request, including the body (frontend).++
	Default: 30s

*--max-redirects*
	Maximum number of redirects to follow.++
	Default: 3

*--metrics*
	Enable Prometheus compatible metrics endpoint.

	If the _metrics_ flag is provided, then the service will expose a Prometheus
	/metrics endpoint and a /debug/vars endpoint from the go _expvar_ package.

	See _METRICS_ for more info.

*--no-debug-vars*
	If the _no-debug-vars_ flag is provided along with the _metrics_ flag, the
	/debug/vars endpoint is removed.

	See _METRICS_ for more info.

*--no-log-ts*
	Do not add a timestamp to logging output.

*--no-fk*
	Disable frontend http keep-alive support (frontend).

*--no-bk*
	Disable backend http keep-alive support (backend).

*--log-json*
	Log in JSON format.

*--allow-content-video*
	Additionally allow video/\* content type.

*--allow-content-audio*
	Additionally allow audio/\* content type.

*--allow-credential-urls*
	Allow urls to contain user/pass credentials.

*--filter-ruleset*=<_FILE_>
	Path to a text file that contains a list (one per line) filter rules.

	If an filter-ruleset file is defined, that file is read and each line
	converted into a filter rule.++
	If the request fails the rule-set, the request is denied.

	See _go-camo-filtering_(5) for more information.

*--server-name*=<_SERVER-NAME_>
	Value to use for the HTTP server field.++
	Default: go-camo

*--expose-server-version*
	Include the server version in the HTTP server response header.

*--enable-xfwd4*
	Enable x-forwarded-for passthrough/generation.

*-v*, *--verbose*
	Show verbose (debug) level log output

*-V*, *--version*
	Print version and exit; specify twice to show license information.

*-h*, *--help*
	Show help output and exit.


# ADD_HEADERS

Additional default headers (headers sent on every reply) can be set with the
*-H, --header* flag. This option can be used multiple times.

The list of default headers sent are:

```
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'none'
```

Additional headers are added to the above set.

As an example, if you wanted to return an Strict-Transport-Security and an
X-Frame-Options header by default, you could add this to the command line:

```
go-camo -k BEEFBEEFBEEF \\
    -H "Strict-Transport-Security: max-age=16070400" \\
    -H "X-Frame-Options: deny"
```

# METRICS

When the *--metrics* flag is used, the service will expose a
Prometheus-compatible /metrics endpoint. This can be used by monitoring
systems to gather data.

The endpoint includes all of the default go_ and process_. In addition, a
number of custom metrics.

A /debug/vars endpoint is also included with *--metrics* by default.
This endpoint returns memstats and some additional data. This endpoint can be
disabled by additionally supplying the *--no-debug-vars* flag.

[- *Metric*
:- *Type*
:- *Description*
|[ camo_response_duration_seconds
:- Histogram
:< A histogram of latencies for proxy responses.
|  camo_response_size_bytes
:  Histogram 
:  A histogram of sizes for proxy responses.
|  camo_proxy_content_length_exceeded_total
:  Counter
:  The number of requests where the content length was exceeded.
|  camo_proxy_reponses_failed_total
:  Counter
:  The number of responses that failed to send to the client.
|  camo_proxy_reponses_truncated_total
:  Counter
:  The number of responess that were too large to send.
|  camo_responses_total
:  Counter
:  Total HTTP requests processed by the go-camo, excluding scrapes.


It also includes a _camo_build_info_ metric that exposes the version. In
addition, you can expose some extra data to metrics via env vars, if desired:

- Revision via _APP_INFO_REVISION_
- Branch via _APP_INFO_BRANCH_
- BuildDate via _APP_INFO_BUILD_DATE_
- You can also override the version by setting _APP_INFO_VERSION_

# EXAMPLES

Listen on loopback port 8080 with a upstream timeout of 6 seconds:

```
go-camo -k BEEFBEEFBEEF \\
    --listen=127.0.0.1:8080 \\
    --timeout=6s
```

Set HMAC key via env var, and an HSTS header:

```
export GOCAMO_HMAC=BEEFBEEFBEEF
go-camo \\
    --listen=127.0.0.1:8080 \\
    --timeout=6s \\
    -H "Strict-Transport-Security: max-age=16070400"
```

# WEBSITE

https://github.com/cactus/go-camo

# SEE_ALSO

https://github.com/atmos/camo
