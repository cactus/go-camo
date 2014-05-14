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

type EncodeCommand struct {
	Base    string `short:"b" long:"base" default:"hex" description:"Encode/Decode base. Either hex or base64"`
}

func (c *EncodeCommand) Execute(args []string) error {
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
	var outUrl string
	switch c.Base {
	case "base64":
		outUrl = encoding.B64EncodeUrl(hmacKeyBytes, oUrl)
	case "hex":
		outUrl = encoding.HexEncodeUrl(hmacKeyBytes, oUrl)
	default:
		return errors.New("Invalid base provided")
	}
	fmt.Println(opts.Prefix + outUrl)
	return nil
}

type DecodeCommand struct {}

func (c *DecodeCommand) Execute(args []string) error {
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

	u, err := url.Parse(oUrl)
	if err != nil {
		return err
	}
	comp := strings.SplitN(u.Path, "/", 3)
	decUrl, valid := encoding.DecodeUrl(hmacKeyBytes, comp[1], comp[2])
	if !valid {
		return errors.New("hmac is invalid")
	}
	log.Println(decUrl)
	return nil
}

var opts struct {
	HmacKey string `short:"k" long:"key" description:"HMAC key"`
	Prefix  string `short:"p" long:"prefix" default:"" description:"Optional url prefix used by encode output"`
}

func main() {
	// clear log prefix -- not needed for tool
	log.SetFlags(0)

	parser := flags.NewParser(&opts, flags.Default)
	parser.AddCommand("encode", "Encode a url and print result",
		"Encode a url and print result", &EncodeCommand{})
	parser.AddCommand("decode", "Decode a url and print result",
		"Decode a url and print result", &DecodeCommand{})

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
