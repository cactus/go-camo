// Copyright (c) 2012-2019 Eli Janssen
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

	"github.com/cactus/go-camo/v2/pkg/camo/encoding"

	flags "github.com/jessevdk/go-flags"
)

// EncodeCommand holds command options for the encode command
type EncodeCommand struct {
	Base       string `short:"b" long:"base" default:"hex" description:"Encode/Decode base. Either hex or base64"`
	Prefix     string `short:"p" long:"prefix" default:"" description:"Optional url prefix used by encode output"`
	Positional struct {
		Url string `positional-arg-name:"URL"`
	} `positional-args:"yes" required:"true"`
}

// Execute runs the encode command
func (c *EncodeCommand) Execute(args []string) error {
	if opts.HmacKey == "" {
		return errors.New("empty HMAC")
	}

	if len(c.Positional.Url) == 0 {
		return errors.New("no url argument provided")
	}

	hmacKeyBytes := []byte(opts.HmacKey)
	var outURL string
	switch c.Base {
	case "base64":
		outURL = encoding.B64EncodeURL(hmacKeyBytes, c.Positional.Url)
	case "hex":
		outURL = encoding.HexEncodeURL(hmacKeyBytes, c.Positional.Url)
	default:
		return errors.New("invalid base provided")
	}
	fmt.Println(strings.TrimRight(c.Prefix, "/") + outURL)
	return nil
}

// DecodeCommand holds command options for the decode command
type DecodeCommand struct {
	Positional struct {
		Url string `positional-arg-name:"URL"`
	} `positional-args:"yes" required:"true"`
}

// Execute runs the decode command
func (c *DecodeCommand) Execute(args []string) error {
	if opts.HmacKey == "" {
		return errors.New("empty HMAC")
	}

	if len(c.Positional.Url) == 0 {
		return errors.New("no url argument provided")
	}

	hmacKeyBytes := []byte(opts.HmacKey)

	u, err := url.Parse(c.Positional.Url)
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

// #nosec G104
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
