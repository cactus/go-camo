chrono
======

[![GoDoc](https://godoc.org/codeberg.org/dropwhile/go-droplibs/chrono?status.png)](https://godoc.org/codeberg.org/dropwhile/go-droplibs/chrono)

## About

A unix epoch time structure that is updated in 1 second intervals.

This can be useful in cases where high request volume may result in many calls to get the system time, where it is fine to tradeoff to save some cycles and render a slightly less precise time.

## Usage

``` go
package main

import (
	"fmt"
	"net/http"

	"codeberg.org/dropwhile/go-droplibs/chrono"
)

type DumbHttpRouter struct {
	dateGenerator fmt.Stringer
	serverName string
}

// ServeHTTP fulfills the http server interface
func (dr *DumbHttpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Date", dr.dateGenerator.String())
	h.Set("Server", dr.serverName)
	w.WriteHeader(http.StatusOK)
}

const timeformat = "Mon, 02 Jan 2006 15:04:05 GMT"

func main() {
	dr := &DumbHttpRouter{
		dateGenerator: chrono.NewTimeNowString(timeformat),
		serverName:    "timestamp-generator",
	}

	http.ListenAndServe(":8090", dr)
}
```

Curl output:
```
└─ % curl -I 'http://127.0.0.1:8090'
HTTP/1.1 200 OK
Date: Wed, 30 Sep 2026 04:04:35 GMT
Server: timestamp-generator
```

## License

Released under a [MIT license][2]. See `LICENSE.md` file for details.


[1]: https://cr.yp.to/libtai/tai64.html
[2]: https://choosealicense.com/licenses/mit/
