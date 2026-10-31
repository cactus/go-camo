types
=====

[![GoDoc](https://godoc.org/codeberg.org/dropwhile/go-droplibs/types?status.png)](https://godoc.org/codeberg.org/dropwhile/go-droplibs/types)

## About

Some helpful types.

## Usage

``` go
package main

import (
	"fmt"
	"os"
	"time"

	"codeberg.org/dropwhile/go-droplibs/types"
)

func main() {
	t := types.Tuple[int, string]{1, "test"}
	t.Left // 1
	t.Right // test
	left, right := t.Unpack() // int(1) , string(test)
	t2 = t.Swap() // types.Tuple[string, int]{"test", 1}
	b, err := json.Marshal(&t) // ["test",1]
}
```

## License

Released under a [MIT license][1]. See `LICENSE.md` file for details.

[1]: https://choosealicense.com/licenses/mit/