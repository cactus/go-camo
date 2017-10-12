go-camo
=======

[![Build Status](https://travis-ci.org/cactus/go-camo.png?branch=master)](https://travis-ci.org/cactus/go-camo)

## Contents

*   [About](#about)
*   [How it works](#how-it-works)
*   [Differences from Camo](#differences-from-camo)
*   [Installing pre-built binaries](#installing-pre-built-binaries)
*   [Building](#building)
*   [Running](#running)
*   [Running on Heroku](#running-on-heroku)
*   [Securing an installation](#securing-an-installation)
*   [Configuring](#configuring)
*   [Additional tools](#additional-tools)
*   [Changelog](#changelog)
*   [License](#license)

## About

This is [Arachnys'][17] version of [`go-camo`][18].

In addition to proxying non-secure images over SSL/TLS, it also proxies stylesheets, and URLs in stylesheets.

---

Go version of [Camo][1] server.

[Camo][1] is a special type of image proxy that proxies non-secure images over
SSL/TLS. This prevents mixed content warnings on secure pages.

It works in conjunction with back-end code that rewrites image URLs and signs
them with an [HMAC][4].

## How it works

The general steps are as follows:

1.  A client requests a page from the web app.
2.  The original URL in the content is parsed.
3.  An HMAC signature of the url is generated.
4.  The url and hmac are encoded.
5.  The encoded url and hmac are placed into the expected format, creating
    the signed url.
6.  The signed url replaces the original image URL.
7.  The web app returns the content to the client.
8.  The client requets the signed url from Go-Camo.
9.  Go-Camo validates the HMAC, decodes the URL, then requests the content
    from the origin server and streams it to the client.

```text
   +----------+           request            +-------------+
   |          |----------------------------->|             |
   |          |                              |             |
   |          |                              |   web-app   |
   |          | img src=https://go-camo/url  |             |
   |          |<-----------------------------|             |
   |          |                              +-------------+
   |  client  |
   |          |     https://go-camo/url      +-------------+ http://some/img
   |          |----------------------------->|             |--------------->
   |          |                              |             |
   |          |                              |   go-camo   |
   |          |           img data           |             |    img data
   |          |<-----------------------------|             |<---------------
   |          |                              +-------------+
   +----------+
```

Go-Camo supports both hex and base64 encoded urls at the same time.

| encoding | tradeoffs                                 |
| -------- | ----------------------------------------- |
| hex      | longer, case insensitive, slightly faster |
| base64   | shorter, case sensitive, slightly slower  |

Benchmark results with go1.8:

```text
BenchmarkHexEncoder-2                 500000          2505 ns/op
BenchmarkB64Encoder-2                 500000          2576 ns/op
BenchmarkHexDecoder-2                 500000          2542 ns/op
BenchmarkB64Decoder-2                 500000          2687 ns/op
```

For examples of url generation, see the [examples](examples/) directory.

While Go-Camo will support proxying HTTPS images as well, for performance
reasons you may choose to filter HTTPS requests out from proxying, and let the
client simply fetch those as they are. The linked code examples do this.

Note that it is recommended to front Go-Camo with a CDN when possible.

## Differences from Camo

*   Go-Camo supports 'Path Format' url format only. Camo's "Query
    String Format" is not supported.
*   Go-Camo supports "allow regex host filters".
*   Go-Camo supports client http keep-alives.
*   Go-Camo provides native SSL support.
*   Go-Camo provides native HTTP/2 support (if built using >=go1.6).
*   Go-Camo supports using more than one os thread (via GOMAXPROCS) without the
    need of multiple instances or additional proxying.
*   Go-Camo builds to a static binary. This makes deploying to large numbers
    of servers a snap.
*   Go-Camo supports both Hex and Base64 urls. Base64 urls are smaller, but
    case sensitive.
*   Go-Camo supports HTTP HEAD requests.
*   Go-Camo allows custom default headers to be added -- useful for things
    like adding [HSTS][10] headers.

## Installing pre-built binaries

Download the tarball appropriate for your OS/ARCH from [releases][13].
Extract, and copy files to desired locations.

## Building

Building requires:

*   git
*   make
*   go (version 1.8 recommended)

Building:

```text
# show make targets
$ make
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  test                run tests
  cover               run tests with cover output
  build-setup         fetch dependencies
  build               build all
  man                 build all man pages
  tar                 build release tarball
  cross-tar           cross compile and build release tarballs

# fetch vendor dependencies
$ make build-setup

# build all binaries and man pages
# strips debug symbols by default
$ make all

# do not strip debug symbols
$ make all GOBUILD_LDFLAGS=""
```

By default, Go-Camo builds with `-tags netgo`. However, for Go versions
older than 1.5, this may not result in Go-Camo using the netgo resolver unless
your Go stdlib is also compiled with `-tags netgo`. For this reason, it is
required to build with at least go-1.5. Building with the latest Go version is
recommended.

## Running

```text
$ go-camo -k "somekey"
```

Go-Camo does not daemonize on its own. For production usage, it is recommended
to launch in a process supervisor, and drop privileges as appropriate.

Examples of supervisors include: [daemontools][5], [runit][6], [upstart][7],
[launchd][8], and many more.

For the reasoning behind lack of daemonization, see [daemontools/why][9]. In
addition, the code is much simpler because of it.

## Running on Heroku

In order to use this on Heroku with the provided Procfile, you need to:

1.  Create an app specifying the https://github.com/kr/heroku-buildpack-go
    buildpack
2.  Set `HMAC_KEY` to the key you are using

## Securing an installation

go-camo will generally do what you tell it to with regard to fetching signed
urls. There is some limited support for trying to preventing [dns
rebinding][15] attacks.

go-camo will attempt to reject any address matching an rfc1918 network block,
or a private scope ipv6 address, be it in the url or via resulting hostname
resolution. Do note, however, that this does not provide protecton for a
network that uses public address space (ipv4 or ipv6), or some of the
[more exotic][16] ipv6 addresses.

The list of networks rejected include...

| Network           | Description                   |
| ----------------- | ----------------------------- |
| `127.0.0.0/8`     | loopback                      |
| `169.254.0.0/16`  | ipv4 link local               |
| `10.0.0.0/8`      | rfc1918                       |
| `172.16.0.0/12`   | rfc1918                       |
| `192.168.0.0/16`  | rfc1918                       |
| `::1/128`         | ipv6 loopback                 |
| `fe80::/10`       | ipv6 link local               |
| `fec0::/10`       | deprecated ipv6 site-local    |
| `fc00::/7`        | ipv6 ULA                      |
| `::ffff:0:0/96`   | IPv4-mapped IPv6 address      |

More generally, it is recommended to either:

*   Run go-camo on an isolated instance (physical, vlans, firewall rules, etc).
*   Run a local resolver for go-camo that returns NXDOMAIN responses for
    addresses in blacklisted ranges (for example unbound's `private-address`
    functionality). This is also useful to help prevent dns rebinding in
    general.

## Configuring

### Environment Vars

*   `GOCAMO_HMAC` - HMAC key to use.

### Command line flags

```text
$ go-camo -h
Usage:
  go-camo [OPTIONS]

Application Options:
  -k, --key=           HMAC key
  -H, --header=        Extra header to return for each response. This option
                       can be used multiple times to add multiple headers
      --stats          Enable Stats
      --no-log-ts      Do not add a timestamp to logging
      --allow-list=    Text file of hostname allow regexes (one per line)
      --max-size=      Max response image size (KB) (5120)
      --timeout=       Upstream request timeout (4s)
      --max-redirects= Maximum number of redirects to follow (3)
      --no-fk          Disable frontend http keep-alive support
      --no-bk          Disable backend http keep-alive support
      --listen=        Address:Port to bind to for HTTP (0.0.0.0:8080)
      --ssl-listen=    Address:Port to bind to for HTTPS/SSL/TLS
      --ssl-key=       ssl private key (key.pem) path
      --ssl-cert=      ssl cert (cert.pem) path
  -v, --verbose        Show verbose (debug) log level output
  -V, --version        print version and exit

Help Options:
  -h, --help          Show this help message
```

If an allow-list file is defined, that file is read and each line converted
into a hostname regex. If a request does not match one of the listed host
regex, then the request is denied.

If stats flag is provided, then the service will track bytes and clients
served, and offer them up at an http endpoint `/status` via HTTP GET request.

If the HMAC key is provided on the command line, it will override (if present),
an HMAC key set in the environment var.

Additional default headers (sent on every response) can also be set. The
`-H, --header` argument may be specified many times.

The list of default headers sent are:

```text
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'none'; img-src data:; style-src 'unsafe-inline'
```

As an example, if you wanted to return a `Strict-Transport-Security` header
by default, you could add this to the command line:

```text
-H "Strict-Transport-Security:  max-age=16070400"
```

## Additional tools

Go-Camo includes a couple of additional tools.

### url-tool

The `url-tool` utility provides a simple way to generate signed URLs from the command line.

```text
$ url-tool -h
Usage:
  url-tool [OPTIONS] <decode | encode>

Application Options:
  -k, --key=    HMAC key
  -p, --prefix= Optional url prefix used by encode output

Help Options:
  -h, --help    Show this help message

Available commands:
  decode  Decode a url and print result
  encode  Encode a url and print result
```

Example usage:

```text
# hex
$ url-tool -k "test" encode -p "https://img.example.org" "http://golang.org/doc/gopher/frontpage.png"
https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67

$ url-tool -k "test" decode "https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67"
http://golang.org/doc/gopher/frontpage.png

# base64
$ url-tool -k "test" encode -b base64 -p "https://img.example.org" "http://golang.org/doc/gopher/frontpage.png"
https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n

$ url-tool -k "test" decode "https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n"
http://golang.org/doc/gopher/frontpage.png
```

### simple-server

The `simple-server` utility has moved to its [own repo][14].

## Changelog

See `CHANGELOG.md`

## License

Released under the [MIT
license](http://www.opensource.org/licenses/mit-license.php). See `LICENSE.md`
file for details.

[1]: https://github.com/atmos/camo
[3]: http://golang.org/doc/install
[4]: http://en.wikipedia.org/wiki/HMAC
[5]: http://cr.yp.to/daemontools.html
[6]: http://smarden.org/runit/
[7]: http://upstart.ubuntu.com/
[8]: http://launchd.macosforge.org/
[9]: http://cr.yp.to/daemontools/faq/create.html#why
[10]: https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security
[11]: https://github.com/cactus/go-camo/issues/6
[12]: https://codereview.appspot.com/151730045#msg10
[13]: https://github.com/cactus/go-camo/releases
[14]: https://github.com/cactus/static-server
[15]: https://en.wikipedia.org/wiki/DNS_rebinding
[16]: https://en.wikipedia.org/wiki/IPv6_address#Special_addresses
[17]: https://www.arachnys.com/
[18]: https://github.com/cactus/go-camo/
