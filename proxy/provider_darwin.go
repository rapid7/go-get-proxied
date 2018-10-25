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
package proxy

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"regexp"
	"strings"
	"time"
)

type providerDarwin struct {
	provider
}

/*
Create a new Provider which is used to retrieve Proxy configurations.
Params:
	configFile: Optional. Path to a configuration file which specifies proxies.
 */
func NewProvider(configFile string) (Provider) {
	c := new(providerDarwin)
	c.init(configFile)
	return c
}

/*
Returns the Proxy configuration for the given proxy protocol and targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
	* scutil: The Network settings
Params:
	protocol: The proxy's protocol (i.e. https)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) Get(protocol string, targetUrlStr string) (Proxy) {
	if proxy := p.provider.get(protocol, ParseTargetURL(targetUrlStr)); proxy != nil {
		return proxy
	}
	return p.readDarwinNetworkSettingProxy(protocol)
}

/*
Returns the HTTPS Proxy configuration for the given targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
	* scutil: The Network settings
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) GetHTTPS(targetUrl string) (Proxy) {
	return p.Get("https", targetUrl)
}


/*
Returns the Network Setting Proxy found.
If none is found, or an error occurs, nil is returned.
Params:
	protocol: The protocol of interest
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) readDarwinNetworkSettingProxy(protocol string) (Proxy) {
	proxy, err := p.parseScutildata(protocol, "scutil", "--proxy")
	if err != nil {
		if isNotFound(err){
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: %s proxy is not enabled.\n", protocol)
		}else if isTimedOut(err){
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: Operation timed out. \n")
		} else {
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: Failed to parse Scutil data, %s\n", err)
		}
	}
	return proxy
}

/*
Returns the Proxy found by parsing the Scutil output.
If none is found, or an error occurs, nil is returned.
Params:
	protocol: The protocol of interest
	name: The name of the program (scutil)
	arg: The list of the arguments (--proxy)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) parseScutildata(protocol string, name string, arg ...string) (Proxy, error) {
	lookupProtocol := strings.ToUpper(protocol) // to cover search for http, HTTP, https, HTTPS

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 1) // Die after one second
	defer cancel()

	cmd := p.proc(ctx, name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, new(timeoutError)
	}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	/* init values */
	var enable bool
	var port string
	var host string

	regexEnable, err := regexp.Compile(lookupProtocol + "Enable:1")
	if err != nil {
		return nil, err
	}
	regexDisable, err := regexp.Compile(lookupProtocol + "Enable:0")
	if err != nil {
		return nil, err
	}
	regexPort, err := regexp.Compile(lookupProtocol + "Port:")
	if err != nil {
		return nil, err
	}
	regexProxy, err := regexp.Compile(lookupProtocol + "Proxy:")
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		str := strings.Replace(scanner.Text(), " ", "", -1) // removing spaces
		if !enable { // don't search if already found
			// if proxy is disabled, stop the search
			protocolDisableFound := regexDisable.FindStringIndex(str)
			if protocolDisableFound != nil {
				break
			}
			protocolEnableFound := regexEnable.FindStringIndex(str)
			if protocolEnableFound != nil {
				enable = true
			}
		}
		if port == "" { // don't search if already found
			portFoundLoc := regexPort.FindStringIndex(str)
			if portFoundLoc != nil {
				port = str[portFoundLoc[1]:]
			}
		}
		if host == "" { // don't search if already found
			proxyFoundLoc := regexProxy.FindStringIndex(str)
			if proxyFoundLoc != nil {
				host = str[proxyFoundLoc[1]:]
			}
		}
	}
	if !enable {
		return nil, new(notFoundError)
	}

	proxyUrlStr := host + ":" + port
	proxyUrl, err := ParseURL(proxyUrlStr, protocol)
	if err != nil {
		return nil, err
	}
	src := "State:/Network/Global/Proxies"
	proxy, err := NewProxy(protocol, proxyUrl, src)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}