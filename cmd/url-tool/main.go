// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// url-tool
package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cactus/go-camo/pkg/camo/encoding"

	flags "github.com/jessevdk/go-flags"
)

// EncodeCommand holds command options for the encode command
type EncodeCommand struct {
	Base   string `short:"b" long:"base" default:"hex" description:"Encode/Decode base. Either hex or base64"`
	Prefix string `short:"p" long:"prefix" default:"" description:"Optional url prefix used by encode output"`
}

// Execute runs the encode command
func (c *EncodeCommand) Execute(args []string) error {
	if opts.HmacKey == "" {
		return errors.New("Empty HMAC")
	}

	if len(args) == 0 {
		return errors.New("No url argument provided")
	}

	oURL := args[0]
	if oURL == "" {
		return errors.New("No url argument provided")
	}

	hmacKeyBytes := []byte(opts.HmacKey)
	var outURL string
	switch c.Base {
	case "base64":
		outURL = encoding.B64EncodeURL(hmacKeyBytes, oURL)
	case "hex":
		outURL = encoding.HexEncodeURL(hmacKeyBytes, oURL)
	default:
		return errors.New("Invalid base provided")
	}
	fmt.Println(c.Prefix + outURL)
	return nil
}

// DecodeCommand holds command options for the decode command
type DecodeCommand struct{}

// Execute runs the decode command
func (c *DecodeCommand) Execute(args []string) error {
	if opts.HmacKey == "" {
		return errors.New("Empty HMAC")
	}

	if len(args) == 0 {
		return errors.New("No url argument provided")
	}

	oURL := args[0]
	if oURL == "" {
		return errors.New("No url argument provided")
	}

	hmacKeyBytes := []byte(opts.HmacKey)

	u, err := url.Parse(oURL)
	if err != nil {
		return err
	}
	comp := strings.SplitN(u.Path, "/", 3)
	decURL, valid := encoding.DecodeURL(hmacKeyBytes, comp[1], comp[2])
	if !valid {
		return errors.New("hmac is invalid")
	}
	fmt.Println(decURL)
	return nil
}

var opts struct {
	HmacKey string `short:"k" long:"key" description:"HMAC key"`
}

func main() {
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
