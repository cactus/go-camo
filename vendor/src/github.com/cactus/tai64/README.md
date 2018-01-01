tai64
=====

[![Build Status](https://travis-ci.org/cactus/tai64.svg?branch=master)](https://travis-ci.org/cactus/tai64)
[![GoDoc](https://godoc.org/github.com/cactus/tai64?status.png)](https://godoc.org/github.com/cactus/tai64)
[![Go Report Card](https://goreportcard.com/badge/github.com/cactus/tai64)](https://goreportcard.com/report/github.com/cactus/tai64)

## About

Formats and parses [TAI64 and TAI64N][1] timestamps.

## Usage

``` go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cactus/tai64"
)

func main() {
	t := time.Now()
	fmt.Println(t)

	s := tai64.FormatNano(t)
	fmt.Println(s)

	p, err := tai64.Parse(s)
	if err != nil {
		fmt.Println("Failed to decode time")
		os.Exit(1)
	}

    // tai64 times are in UTC
    fmt.Println(p)

    // time.Equal properly compares times with different locations.
	if t.Equal(p) {
		fmt.Println("equal")
	} else {
		fmt.Println("not equal")
	}
}
```

Output:

```
2016-05-25 13:44:01.281160355 -0700 PDT
@4000000057460eb510c22aa3
2016-05-25 20:44:01.281160355 +0000 UTC
equal
```


[1]: http://www.tai64.com
