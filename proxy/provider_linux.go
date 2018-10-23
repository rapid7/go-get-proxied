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

type providerLinux struct {
	provider
}

/*
Create a new Provider which is used to retrieve Proxy configurations.
Params:
	configFile: Optional. Path to a configuration file which specifies proxies.
 */
func NewProvider(configFile string) (Provider) {
	c := new(providerLinux)
	c.init(configFile)
	return c
}

/*
Returns the Proxy configuration for the given proxy protocol and targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
Params:
	protocol: The proxy's protocol (i.e. https)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerLinux) Get(protocol string, targetUrlStr string) (Proxy) {
	return p.provider.get(protocol, ParseTargetURL(targetUrlStr))
}

/*
Returns the HTTPS Proxy configuration for the given targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerLinux) GetHTTPS(targetUrl string) (Proxy) {
	return p.Get("https", targetUrl)
}