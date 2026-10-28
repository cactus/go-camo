# AssertHttp - Some helper http test functions.

[![GoDoc](https://godoc.org/codeberg.org/dropwhile/go-droplibs/asserthttp?status.png)](https://godoc.org/codeberg.org/dropwhile/go-droplibs/asserthttp)

## Usage

```go
package main

import (
	"net/http"
	"testing"

	"codeberg.org/dropwhile/go-droplibs/assert"
	"codeberg.org/dropwhile/go-droplibs/asserthttp"
)

func TestSomething(t *testing.T) {
    t.Parallel()

    resp, err := http.Get("http://example.com/")
    assert.Nil(t, err) // test assertion that something is nil
    // output => &errors.errorString{s:"my error"}; want: <nil>;

    // assert http header matches
    asserthttp.Header(resp, "server-name", "example server")
    // output => http header ('server-name') mismatch. got 'foo-server 1.0'; want 'example server';

    // assert http status code matches
    asserthttp.StatusCode(resp, 404)
    // output => http status code mismatch. got '204'; want '404';

    // assert http response body matches
    asserthttp.Body(resp, "this an example")
    // output => http response body mismatch. got 'some example'; want 'this is an example';
}
```
