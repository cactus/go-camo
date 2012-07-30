// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cactus/go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
)

func main() {
	// command line flags
	hmacKey := flag.String("hmac-key", "", "HMAC Key")
	configFile := flag.String("config-file", "", "JSON Config File")
	prefix := flag.String("prefix", "", "Optional url prefix used by encode output")
	encode := flag.Bool("encode", false, "Encode a url and print result")
	decode := flag.Bool("decode", false, "Decode a url and print result")
	// parse said flags
	flag.Parse()

	// clear log prefix -- not needed for tool
	log.SetFlags(0)

	// Anonymous struct Container for holding configuration parameters
	// parsed from JSON config file.
	config := camoproxy.Config{}

	if *configFile != "" {
		b, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatal("Could not read configFile", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatal("Could not parse configFile", err)
		}
	}

	if *encode == true && *decode == true {
		log.Fatal("Encode and Decode are mutually exclusive. Doing nothing.")
	}

	if *encode == false && *decode == false {
		log.Fatal("No action requested. Doing nothing.")
	}

	// flags override config file
	if *hmacKey != "" {
		config.HmacKey = *hmacKey
	}

	oUrl := flag.Arg(0)
	if oUrl == "" {
		log.Fatal("No url argument provided")
	}

	hmacKeyBytes := []byte(config.HmacKey)

	if *encode == true {
		outUrl := camoproxy.EncodeUrl(&hmacKeyBytes, oUrl)
		fmt.Println(*prefix + outUrl)
	}

	if *decode == true {
		u, err := url.Parse(oUrl)
		if err != nil {
			log.Fatal(err)
		}
		comp := strings.SplitN(u.Path, "/", 3)
		decUrl, valid := camoproxy.DecodeUrl(&hmacKeyBytes, comp[1], comp[2])
		if !valid {
			log.Fatal("hmac is invalid")
		}
		log.Println(decUrl)
	}
}
