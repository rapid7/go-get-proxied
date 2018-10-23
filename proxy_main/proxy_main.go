// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rapid7/go-get-proxied/proxy"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	protocolP := flag.String("p", "https", "Optional. The proxy protocol you wish to lookup. Default: https")
	configP := flag.String("c", "", "Optional. Path to configuration file.")
	targetP := flag.String("t", "", "Optional. Target URL which the proxy will be used for. Default: *")
	quietP := flag.Bool("q", false, "Optional. Quiet mode; only write the URL of a proxy to stdout (if found). Default: false")
	flag.Parse()
	var (
		protocol 	string
		config 		string
		target 		string
		quiet 		bool
	)
	if protocolP != nil {
		protocol = *protocolP
	}
	if configP != nil {
		config = *configP
	}
	if targetP != nil {
		target = *targetP
	}
	if quietP != nil {
		quiet = *quietP
	}
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	p := proxy.NewProvider(config).Get(protocol, target)
	var exit int
	if p != nil {
		if !quiet {
			fmt.Printf("Proxy: %s\n", p)
			b, _ := json.MarshalIndent(p, "", "   ")
			fmt.Printf("Proxy JSON: %s", string(b))
		} else {
			println(p.URL().String())
		}
	} else {
		exit = 1
		if !quiet {
			println("Proxy: nil")
		}
	}
	os.Exit(exit)
}