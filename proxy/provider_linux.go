// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
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