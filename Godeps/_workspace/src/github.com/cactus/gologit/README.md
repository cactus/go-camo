gologit
=======

A simple wrapper for Go "log" that provides toggle-able debug support.

## Installation

    $ go get github.com/cactus/gologit

## Usage

Simplest:

    import (
        "flag"
        "github.com/cactus/gologit"
    )
   
    func main() {
        debug := flag.Bool("debug", false, "Enable Debug Logging")
        flag.Parse()

        // set debug true/false
        gologit.Logger.Set(*debug)
        // this prints only if debug is true
        gologit.Logger.Debugln("Debug Logging enabled!")
    }


Simple:

    import (
        "flag"
        "github.com/cactus/gologit"
    )
   
    // alias exported gologit Logger to a short name for convenience
    var logger = gologit.Logger

    func main() {
        debug := flag.Bool("debug", false, "Enable Debug Logging")
        flag.Parse()

        logger.Set(*debug)
        logger.Debugln("Logging enabled")
    }


When you don't want to share:

    import (
        "flag"
        "github.com/cactus/gologit"
    )
   
    // make a new one (unique to this module)
    var logger = gologit.New(false)

    func main() {
        debug := flag.Bool("debug", false, "Enable Debug Logging")
        flag.Parse()

        logger.Set(*debug)
        logger.Debugln("Logging enabled")
    }


Pass it like a logging potato:

    import (
        "flag"
        "github.com/cactus/gologit"
    )
   
    // make a new one (unique to this module)
    var logger = gologit.New(false)

    func main() {
        debug := flag.Bool("debug", false, "Enable Debug Logging")
        flag.Parse()

        logger := gologit.New(*debug)
        logger.Debugln("Logging enabled")
        SomeOtherFunc(logger)
    }

## Documentation

More documentation available at [go.pkgdoc.org][1]

## License

Released under the [MIT license][2]. See `LICENSE` file for details.

[1]: http://go.pkgdoc.org/github.com/cactus/gologit
[2]: http://www.opensource.org/licenses/mit-license.php

