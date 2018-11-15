// Copyright 2018, Rapid7, Inc.
// License: BSD-3-clause
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// * Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
// * Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software
// without specific prior written permission.
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
		protocol string
		config   string
		target   string
		quiet    bool
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
			b, _ := json.MarshalIndent(p, "", "   ")
			fmt.Println(string(b))
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
