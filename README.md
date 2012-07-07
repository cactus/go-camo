go-camo
=======

## About

Go version of [Camo][1] server.

## Differences from Camo

### Supported Features

*   Support for 'Path Format' only (does not support 'Query String Format').
*   Supports both allow and deny regex host filters.
*   Supports client http keepalives.
*   Native SSL Support
*   Supports using more than one os thread (via GOMAXPROCS) without the need of
    multiple instances or additional proxying.

## Building

    $ git clone https://github.com/cactus/go-camo.git
    $ cd go-camo
    $ export GOPATH=$(pwd)
    $ go get github.com/cactus/gologit
    $ go get code.google.com/p/gorilla/mux
    $ go install go-camo

## Running in production mode

    $ $GOPATH/bin/go-camo -config-file=config.json -follow-redirects

## Running under devweb

[Devweb][2] is useful for developing. To run under devweb:

    $ go get code.google.com/p/rsc/devweb
    $ PATH=.:$PATH $GOPATH/bin/devweb -addr=127.0.0.1:8080 go-camo/go-camo-devweb
    $ rm -f ./prox.exe  #devweb drops this file. clean it up

## Configuring

    $ $GOPATH/bin/go-camo -h
    Usage of go-camo:
      -bind-address="0.0.0.0:8080": Address:Port to bind to for HTTP
      -bind-address-ssl="": Address:Port to bind to for HTTPS/SSL/TLS
      -config-file="": JSON Config File
      -debug=false: Enable Debug Logging
      -follow-redirects=false: Enable following upstream redirects
      -hmac-key="": HMAC Key
      -max-size=5120: Max response image size (KB)
      -ssl-cert="": Path to ssl cert (cert.pem). Required if bind-address-ssl is specified.
      -ssl-key="": Path to ssl private key (key.pem). Required if bind-address-ssl is specified.
      -timeout=4: Upstream request timeout in seconds

    
    $ cat config.json
    {
        "HmacKey": "Some long string here...",
        "AllowList": [],
        "DenyList": [
            "^10\\.",
            "^169\\.254",
            "^192\\.168",
            "^172\\.(?:(?:1[6-9])|(?:2[0-9])|(?:3[0-1]))",
            "^(?:.*\\.)?example\\.(?:com|org|net)$"
        ]
    }

*   `HmacKey` is a secret key seed to the HMAC used for signing and
    validation.
*   `Allowlist` is a list of host matches to always allow.
*   `Denylist` is a list of host matches to reject.

If an AllowList is defined, and a request does not match the host regex,
then the request is denied. Default is all requests pass the Allowlist if
none is specified.

DenyList entries are matched after Allowlist, so they take precedence.
Even if a request would be allowed by an Allowlist, a Denylist match would
deny it.

Option flags, if provided, override those in the config file.

## Changelog

See `CHANGELOG.md`

## License

Released under the [MIT
license](http://www.opensource.org/licenses/mit-license.php). See `LICENSE.md`
file for details.

[1]: https://github.com/atmos/camo
[2]: http://code.google.com/p/rsc/source/browse/devweb
