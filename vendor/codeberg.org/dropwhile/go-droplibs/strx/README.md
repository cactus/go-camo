strx
=====

[![GoDoc](https://godoc.org/codeberg.org/dropwhile/go-droplibs/strx?status.png)](https://godoc.org/codeberg.org/dropwhile/go-droplibs/strx)

## About

Some helpful string utilities.

## Usage

``` go
package main

import (
	"fmt"
	"os"
	"time"

	"codeberg.org/dropwhile/go-droplibs/strx"
)

func main() {

	ok := strx.ContainsOneOf("a very fine lemon", []string{"radish", "cucumber", "lettuce"})
	fmt.Printf("%t\n", ok)

	ok = strx.ContainsOneOf("who put cardamom in my drink?", []string{"cardamom", "nutmeg", "cinnamon"})
	fmt.Printf("%t\n", ok)
}
```

Output:

```
false
true
```

## License

Released under the [ISC license][2]. See `LICENSE.md` file for details.


[1]: https://cr.yp.to/libtai/tai64.html
[2]: https://choosealicense.com/licenses/isc/
