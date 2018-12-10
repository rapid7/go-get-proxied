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
	jsonP := flag.Bool("j", false, "Optional. If a proxy is found, write it as JSON instead of a URL.")
	verboseP := flag.Bool("v", false, "Optional. If set, log content will be sent to stderr.")
	flag.Parse()
	var (
		protocol string
		config   string
		target   string
		jsonOut  bool
		verbose  bool
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
	if jsonP != nil {
		jsonOut = *jsonP
	}
	if verboseP != nil {
		verbose = *verboseP
	}
	if verbose {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	p := proxy.NewProvider(config).GetProxy(protocol, target)
	var exit int
	if p != nil {
		if jsonOut {
			b, _ := json.MarshalIndent(p, "", "   ")
			fmt.Println(string(b))
		} else {
			println(p.URL().String())
		}
	} else {
		exit = 1
	}
	os.Exit(exit)
}
