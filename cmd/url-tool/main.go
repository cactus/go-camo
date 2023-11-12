// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// url-tool
package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/cactus/go-camo/v2/pkg/camo/encoding"

	"github.com/alecthomas/kong"
)

// ServerVersion holds the server version string
var ServerVersion = "no-version"

// EncodeCommand holds command options for the encode command
type EncodeCmd struct {
	Base   string `name:"base" short:"b" enum:"hex,base64" default:"hex" help:"Encode/Decode base. One of: ${enum}"`
	Prefix string `name:"prefix" short:"p" default:"" help:"Optional url prefix used by encode output"`
	Url    string `arg:"" name:"URL" help:"URL to encode"`
}

// Execute runs the encode command
func (cmd *EncodeCmd) Run(cli *CLI) error {
	if cli.HmacKey == "" {
		return errors.New("empty HMAC")
	}

	if len(cmd.Url) == 0 {
		return errors.New("no url argument provided")
	}

	hmacKeyBytes := []byte(cli.HmacKey)
	var outURL string
	switch cmd.Base {
	case "base64":
		outURL = encoding.B64EncodeURL(hmacKeyBytes, cmd.Url)
	case "hex":
		outURL = encoding.HexEncodeURL(hmacKeyBytes, cmd.Url)
	default:
		return errors.New("invalid base provided")
	}
	fmt.Println(strings.TrimRight(cmd.Prefix, "/") + outURL)
	return nil
}

// DecodeCommand holds command options for the decode command
type DecodeCmd struct {
	Url string `arg:"" name:"URL" help:"URL to decode"`
}

// Execute runs the decode command
func (cmd *DecodeCmd) Run(cli *CLI) error {
	if cli.HmacKey == "" {
		return errors.New("empty HMAC")
	}

	if len(cmd.Url) == 0 {
		return errors.New("no url argument provided")
	}

	hmacKeyBytes := []byte(cli.HmacKey)

	u, err := url.Parse(cmd.Url)
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

type CLI struct {
	// global options
	Version kong.VersionFlag `name:"version" short:"V" help:"Print version information and quit"`
	HmacKey string           `name:"key" short:"k" help:"HMAC key"`

	// subcommands
	Encode EncodeCmd `cmd:"" aliases:"enc" help:"Encode a url and print result"`
	Decode DecodeCmd `cmd:"" aliases:"dec" help:"Decode a url and print result"`
}

// #nosec G104
func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("url-tool"),
		kong.Description("A simple way to work with signed go-camo URLs from the command line"),
		kong.UsageOnError(),
		kong.Vars{"version": ServerVersion},
	)
	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
