url-tool(1)

# NAME

url-tool - Provides a simple way to generate signed URLs from the command line

# SYNOPSIS

*url-tool* [<_OPTIONS_...> [_OPTION-ARGUMENT_]]...

# DESCRIPTION

_url-tool_(1) provides a simple way to generate signed URLs from the command line,
compatible with _go-camo_(1).

# OPTIONS

*-k*, *--key*=<_HMAC_KEY_>
	The HMAC key to use.

*-h*, *--help*
	Show help output and exit.

# COMMANDS

_url-tool_(1) has two subcommands.

*encode* <_URL_>
	Encode a URL.

	Available encode options:

	*-b*, *--base*=<_BASE_>
		The base encoding to use. Can be one of hex or base64.

	*--prefix*=<_PREFIX_>
		Optional url prefix used by encode output.

*decode* <_URL_>
	Decode a URL.

# EXAMPLES

Encode a URL as hex
```
$ ./url-tool encode \\
    -k "test" \\
    encode -p "https://img.example.org" \\
    "http://golang.org/doc/gopher/frontpage.png"
https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67
```

Encode a URL as base64
```
$ ./url-tool encode \\
    -k "test" \\
    -b base64 \\
    -p "https://img.example.org" \\
    "http://golang.org/doc/gopher/frontpage.png"
https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n
```

Decode a hex encoded URL
```
$ ./url-tool decode \\
    -k "test" \\
    "https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67"
http://golang.org/doc/gopher/frontpage.png
```

Decode a base64 encoded URL
```
$ ./url-tool decode \\
    -k "test" \\
    "https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n"
http://golang.org/doc/gopher/frontpage.png
```

# WEBSITE

https://github.com/cactus/go-camo
