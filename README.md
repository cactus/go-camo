go-camo
=======

[![Build Status](https://travis-ci.org/cactus/go-camo.png?branch=master)](https://travis-ci.org/cactus/go-camo)

## About

Go version of [Camo][1] server.

[Camo][1] is a special type of image proxy that proxies non-secure images over
SSL/TLS. This prevents mixed content warnings on secure pages.

It works in conjunction with back-end code to rewrite image URLs and sign them
with an [HMAC][4].

## How it works

First you parse the original URL, generate an HMAC signature of it, then encode
it, and then place the pieces into the expected format replacing the original
image URL.

The client requests the URL to Go-Camo. Go-Camo validates the HMAC, decodes the
URL, requests the content and streams it to the client.

Go-Camo supports both hex and base64 encoded urls at the same time.

| encoding | tradeoffs                                               |
| -------- | ------------------------------------------------------- |
| hex      | longer, case insensitive, slightly faster encode/decode |
| base64   | shorter, case sensitive, slightly slower encode/decode  |

For examples of url generation, see the [examples](examples/) directory.

While Go-Camo will support proxying HTTPS images as well, for performance
reasons you may choose to filter HTTPS requests out from proxying, and let the
client simply fetch those as they are. The code example above does this.

Note that it is recommended to front Go-Camo with a CDN when possible.

## Differences from Camo

*   Go-Camo supports 'Path Format' url format only. Camo's "Query
    String Format" is not supported.
*   Go-Camo supports "allow regex host filters".
*   Go-Camo supports client http keep-alives.
*   Go-Camo provides native SSL support.
*   Go-Camo supports using more than one os thread (via GOMAXPROCS) without the
    need of multiple instances or additional proxying.
*   Go-Camo builds to a static binary. This makes deploying to large numbers
    of servers a snap.
*   Go-Camo supports both Hex and Base64 urls. Base64 urls are smaller, but
    case sensitive.
*   Go-Camo supports HTTP HEAD requests.
*   Go-Camo allows custom default headers to be added -- useful for things
    like adding [HSTS][10] headers.

## Building

Building requires `git` and `make`. Optional requirements are `pod2man` (to
build man pages), and fpm (to build rpms).  A functional [Go][3] installation
is also required.

    # show make targets
    $ make
    Available targets:
      help                this help
      clean               clean up
      all                 build binaries and man pages
      build               build all
      build-go-camo       build go-camo
      build-url-tool      build url tool
      build-simple-server build simple server
      test                run tests
      cover               run tests with cover output
      man                 build all man pages
      man-go-camo         build go-camo man pages
      man-url-tool        build url-tool man pages
      man-simple-server   build simple-server man pages
      rpm                 build rpm

    # build all binaries and man pages. results will be in build/ dir
    $ make all

    # as an alternative to the previous command, build and strip debug symbols.
    # this is useful for production, and reduces the resulting file size.
    $ make all GOBUILD_LDFLAGS="-s"


## Running

    $ $GOPATH/bin/go-camo -c config.json

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

## Configuring

### Environment Vars

*   `GOCAMO_HMAC` - HMAC key to use.

### Command line flags

    $ $GOPATH/bin/go-camo -h
    Usage:
      go-camo [OPTIONS]

    Application Options:
      -k, --key=           HMAC key
      -H, --header=        Extra header to return for each response. This option
                           can be used multiple times to add multiple headers
          --stats          Enable Stats
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


If an allow-list file is defined, that file is read and each line converted
into a hostname regex. If a request does not match one of the listed host
regex, then the request is denied.

If stats flag is provided, then the service will track bytes and clients
served, and offer them up at an http endpoint `/status` via HTTP GET request.

If the HMAC key is provided on the command line, it will override (if present),
an HMAC key set in the environment var.

Additional default headers (headers sent on every reply) can also be set. The
`-H, --header` argument may be specified many times.

The list of default headers sent are:

    X-Content-Type-Options: nosniff
    X-XSS-Protection: 1; mode=block
    Content-Security-Policy: default-src 'none'`

As an example, if you wanted to return a `Strict-Transport-Security` header
by default, you could add this to the command line:

    -H "Strict-Transport-Security:  max-age=16070400"

## Additional tools

Go-Camo includes a couple of additional tools.

### url-tool

The `url-tool` utility provides a simple way to generate signed URLs from the command line.

    $ $GOPATH/bin/url-tool -h
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

Example usage:

    # hex
    $ $GOPATH/bin/url-tool -k "test" encode -p "https://img.example.org" "http://golang.org/doc/gopher/frontpage.png"
    https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67

    $ $GOPATH/bin/url-tool -k "test" decode "https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67"
    http://golang.org/doc/gopher/frontpage.png

    # base64
    $ $GOPATH/bin/url-tool -k "test" encode -b base64 -p "https://img.example.org" "http://golang.org/doc/gopher/frontpage.png"
    https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n

    $ $GOPATH/bin/url-tool -k "test" decode "https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n"
    http://golang.org/doc/gopher/frontpage.png

### simple-server

The `simple-server` utility is useful for testing. It serves the contents of a
given directory over http. Nothing more.

    $ $GOPATH/bin/simple-server -h
    Usage:
      simple-server [OPTIONS] DIR

    Application Options:
      -l, --listen= Address:Port to bind to for HTTP (0.0.0.0:8000)

    Help Options:
      -h, --help    Show this help message

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
