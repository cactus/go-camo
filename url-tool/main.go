// go-camo daemon (go-camod)
package main

import (
	"errors"
	"fmt"
	"github.com/cactus/go-camo/camoproxy/encoding"
	flags "github.com/jessevdk/go-flags"
	"log"
	"net/url"
	"os"
	"strings"
)

type Command struct {
	Name string
}

func (c *Command) Usage() string {
	return fmt.Sprintf("[%s-OPTIONS] URL", c.Name)
}

func (c *Command) Execute(args []string) error {
	// clear log prefix -- not needed for tool
	log.SetFlags(0)

	if opts.HmacKey == "" {
		return errors.New("Empty HMAC")
	}

	if len(args) == 0 {
		return errors.New("No url argument provided")
	}

	oUrl := args[0]
	if oUrl == "" {
		return errors.New("No url argument provided")
	}

	hmacKeyBytes := []byte(opts.HmacKey)

	switch c.Name {
	case "encode":
		outUrl := encoding.EncodeUrl(&hmacKeyBytes, oUrl)
		fmt.Println(opts.Prefix + outUrl)
		return nil
	case "decode":
		u, err := url.Parse(oUrl)
		if err != nil {
			return err
		}
		comp := strings.SplitN(u.Path, "/", 3)
		decUrl, valid := encoding.DecodeUrl(&hmacKeyBytes, comp[1], comp[2])
		if !valid {
			return errors.New("hmac is invalid")
		}
		log.Println(decUrl)
		return nil
	}
	return errors.New("unknown command")
}

var opts struct {
	HmacKey string `short:"k" long:"key" description:"HMAC key"`
	Prefix  string `short:"p" long:"prefix" default:"" description:"Optional url prefix used by encode output"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.AddCommand("encode", "Encode a url and print result",
		"Encode a url and print result", &Command{Name: "encode"})
	parser.AddCommand("decode", "Decode a url and print result",
		"Decode a url and print result", &Command{Name: "decode"})

	// parse said flags
	_, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		os.Exit(1)
	}
}
