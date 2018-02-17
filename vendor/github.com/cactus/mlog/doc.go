// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

/*
Package mlog provides a purposefully basic logging library for Go.

mlog only has 3 logging levels: debug, info, and fatal.

Each logging level has 3 logging methods. As an example, the following methods
log at the "info" level: Info, Infof, Infom. There are similar methods for
the fatal and debug levels.

Example usage:

    import (
        "bytes"

        "github.com/cactus/mlog"
    )

    func main() {
        mlog.Infom("this is a log", mlog.Map{
            "interesting": "data",
            "something": 42,
        })

        mlog.Debugm("this won't print")

        // set flags for the default logger
        // alternatively, you can create your own logger
        // and supply flags at creation time
        mlog.SetFlags(mlog.Ldebug)

        mlog.Debugm("this will print!")

        mlog.Debugm("can it print?", mlog.Map{
            "how_fancy": []byte{'v', 'e', 'r', 'y', '!'},
            "this_too": bytes.NewBuffer([]byte("if fmt.Print can print it!")),
        })

        // you can use a more classical Printf type log method too.
        mlog.Debugf("a printf style debug log: %s", "here!")
        mlog.Infof("a printf style info log: %s", "here!")

        mlog.Fatalm("time for a nap", mlog.Map{"cleanup": false})
    }
*/
package mlog
