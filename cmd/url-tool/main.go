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
	Base         string            `short:"b" long:"base" default:"base64" description:"Encode/Decode base. Either hex or base64"`
	Prefix       string            `short:"p" long:"prefix" default:"" description:"Optional url prefix used by encode output"`
	ExtraHeaders map[string]string `short:"H" long:"header" default:"" description:"Add an extra header to the encoded url. May be supplied multiple times. Requires camo additional headers support."`
	Positional   struct {
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

	extraHeaders := encoding.SimpleHeader{}
	if len(c.ExtraHeaders) > 0 {
		for k, v := range c.ExtraHeaders {
			if len(k) == 0 || len(v) == 0 {
				continue
			}
			extraHeaders[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	hmacKeyBytes := []byte(opts.HmacKey)
	var outURL string
	var err error
	switch c.Base {
	case "base64":
		outURL, err = encoding.EncodeURL(encoding.B64Codec, hmacKeyBytes, c.Positional.Url, extraHeaders)
	case "hex":
		outURL, err = encoding.EncodeURL(encoding.HexCodec, hmacKeyBytes, c.Positional.Url, extraHeaders)
	default:
		return errors.New("invalid base provided")
	}
	if err != nil {
		return err
	}
	fmt.Println(strings.TrimRight(c.Prefix, "/") + outURL)
	return nil
}

// DecodeCommand holds command options for the decode command
type DecodeCommand struct {
	PrintHeaders bool `short:"x" long:"print-headers"  description:"Print any encoded extra headers, if present"`
	Positional   struct {
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
	comp := strings.Split(u.Path, "/")
	if len(comp) < 3 || len(comp) > 4 {
		return errors.New("bad url provided")
	}

	encExtraHdr := ""
	if len(comp) == 4 {
		encExtraHdr = comp[3]
	}
	decURL, extraHdr, valid := encoding.DecodeURL(hmacKeyBytes, comp[1], comp[2], encExtraHdr)
	if !valid {
		return errors.New("hmac is invalid")
	}
	fmt.Println(decURL)
	if c.PrintHeaders {
		if len(extraHdr) > 0 {
			fmt.Println("Additional Headers: ")
			for k, v := range extraHdr {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	}
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
