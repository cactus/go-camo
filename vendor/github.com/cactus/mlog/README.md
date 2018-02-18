mlog
====

[![Build Status](https://travis-ci.org/cactus/mlog.png?branch=master)](https://travis-ci.org/cactus/mlog)
[![GoDoc](https://godoc.org/github.com/cactus/mlog?status.png)](https://godoc.org/github.com/cactus/mlog)
[![Go Report Card](https://goreportcard.com/badge/cactus/mlog)](https://goreportcard.com/report/cactus/mlog)
[![License](https://img.shields.io/github/license/cactus/mlog.svg)](https://github.com/cactus/mlog/blob/master/LICENSE.md)

## About

A purposefully basic logging library for Go.

mlog only has 3 logging levels: Debug, Info, and Fatal.

### Why only 3 levels?

Dave Cheney [wrote a great post][1], that made me rethink my own approach to
logging, and prompted me to start writing mlog.

### How does it work?

Logging methods are:

*   `Debug` - conditionally (if debug is enabled) logs message at level
    "debug".
*   `Debugf` - similar to `Debug`, but supports printf formatting.
*   `Debugm` - similar to `Debug`, but logs an mlog.Map as extra data.
*   `Info` - logs message at level "info". `Print` is an alias for `Info`.
*   `Infof` - similar to `Info`, but supports printf formatting. `Printf` is an
    alias for `Infof`.
*   `Infom` - similar to `Info`, but logs an mlog.Map as extra data. `Printm`
    is an alias for `Infom`.
*   `Fatal` - logs message at level "fata", then calls `os.Exit(1)`.
*   `Fatalf` - similar to `Fatal`, but supports printf formatting.
*   `Fatalm` - similar to `Fatal`, but logs an mlog.Map as extra data.

That's it!

For more info, check out the [docs][3].

## Usage

``` go
import (
    "bytes"

    "github.com/cactus/mlog"
)

func main() {
    mlog.Info("this is a log")

    mlog.Infom("this is a log with more data", mlog.Map{
        "interesting": "data",
        "something":   42,
    })

    thing := mlog.Map(
        map[string]interface{}{
            "what‽":       "yup",
            "this-works?": "as long as it is a mlog.Map",
        },
    )

    mlog.Infom("this is also a log with more data", thing)

    mlog.Debug("this won't print")

    // set flags for the default logger
    // alternatively, you can create your own logger
    // and supply flags at creation time
    mlog.SetFlags(mlog.Ltimestamp | mlog.Ldebug)

    mlog.Debug("now this will print!")

    mlog.Debugm("can it print?", mlog.Map{
        "how_fancy": []byte{'v', 'e', 'r', 'y', '!'},
        "this_too":  bytes.NewBuffer([]byte("if fmt.Print can print it!")),
    })

    // you can use a more classical Printf type log method too.
    mlog.Debugf("a printf style debug log: %s", "here!")
    mlog.Infof("a printf style info log: %s", "here!")

    // how about logging in json?
    mlog.SetEmitter(&mlog.FormatWriterJSON{})
    mlog.Infom("something", mlog.Map{
        "one": "two",
        "three":  3,
    })

    mlog.Fatalm("time for a nap", mlog.Map{"cleanup": false})
}
```

Output:

```
time="2016-04-29T19:59:11.474362716-07:00" level="I" msg="this is a log"
time="2016-04-29T19:59:11.474506079-07:00" level="I" msg="this is a log with more data" interesting="data" something="42"
time="2016-04-29T19:59:11.474523514-07:00" level="I" msg="this is also a log with more data" this-works?="as long as it is a mlog.Map" what‽="yup"
time="2016-04-29T19:59:11.474535676-07:00" msg="now this will print!"
time="2016-04-29T19:59:11.474542467-07:00" msg="can it print?" how_fancy="[118 101 114 121 33]" this_too="if fmt.Print can print it!"
time="2016-04-29T19:59:11.474551625-07:00" msg="a printf style debug log: here!"
time="2016-04-29T19:59:11.474578991-07:00" msg="a printf style info log: here!"
{"time": "2016-04-29T19:59:11.474583762-07:00", "msg": "something" "extra": {"one": "two", "three": "3"}}
{"time": "2016-04-29T19:59:11.474604928-07:00", "msg": "time for a nap" "extra": {"cleanup": "false"}}
exit status 1
```

## License

Released under the [MIT license][2]. See `LICENSE.md` file for details.

[1]: http://dave.cheney.net/2015/11/05/lets-talk-about-logging
[2]: http://www.opensource.org/licenses/mit-license.php
[3]: https://godoc.org/github.com/cactus/mlog
