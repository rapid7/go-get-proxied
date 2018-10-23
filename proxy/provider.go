// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	proxyKeyFormat = "%s_PROXY"
	noProxyKeyUpper = "NO_PROXY"
	noProxyKeyLower = "no_proxy"
)

type Provider interface {
	/*
	Returns the Proxy configuration for the given proxy protocol and targetUrl.
	If none is found, or an error occurs, nil is returned.
	Params:
		protocol: The proxy's protocol (i.e. https)
		targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	Returns:
		Proxy: A proxy was found.
		nil: A proxy was not found, or an error occurred.
	*/
	Get(protocol string, targetUrl string) (Proxy)
	/*
	Returns the HTTPS Proxy configuration for the given targetUrl.
	If none is found, or an error occurs, nil is returned.
	Params:
		targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	Returns:
		Proxy: A proxy was found.
		nil: A proxy was not found, or an error occurred.
	*/
	GetHTTPS(targetUrl string) (Proxy)
}

type getEnvAdapter func(string) (string)

type provider struct {
	configFile	string
	getEnv		getEnvAdapter
}

func (p *provider) init(configFile string) {
	p.configFile = configFile
	p.getEnv = os.Getenv
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
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *provider) get(protocol string, targetUrl *url.URL) (Proxy) {
	proxy := p.readConfigFileProxy(protocol)
	if proxy != nil {
		return proxy
	}
	return p.readSystemEnvProxy(protocol, targetUrl)
}

/*
Unmarshal the proxy.config file, and return the first proxy matched for the given protocol.
If no proxy is found, or an error occurs reading the proxy.config file, nil is returned.
Params:
	protocol: The proxy's protocol (i.e. https)
Returns:
	Proxy: A proxy is found in proxy.config for the given protocol.
	nil: No proxy is found or an error occurs reading the proxy.config file.
*/
func (p *provider) readConfigFileProxy(protocol string) (Proxy) {
	proxyJson, err := p.unmarshalProxyConfigFile()
	if err != nil {
		log.Printf("[proxy.Provider.readConfigFileProxy]: %s\n", err)
		return nil
	}
	uStr, exists := proxyJson[protocol] ; if !exists {
		return nil
	}
	uUrl, uErr := ParseURL(uStr, protocol)
	var uProxy Proxy
	if uErr == nil {
		uProxy, uErr = NewProxy(protocol, uUrl, "ConfigurationFile")
	}
	if uErr != nil {
		log.Printf("[proxy.Provider.readConfigFileProxy]: invalid config file proxy, skipping \"%s\": \"%s\"\n", protocol, uStr)
		return nil
	}
	return uProxy
}

/*
Unmarshal the proxy.config file into a simple map[string]string structure.
Returns:
	map[string]string, nil: Unmarshal of proxy.config is successful.
	nil, error: Unmarshal of proxy.config is not successful.
*/
func (p *provider) unmarshalProxyConfigFile() (map[string]string, error) {
	m := map[string]string{}
	if p.configFile == "" {
		return m, nil
	}
	f := filepath.Join(p.configFile)
	stat, err := os.Stat(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("proxy configuration file not present: %s", f))
	} else if stat.IsDir() {
		return nil, errors.New(fmt.Sprintf("proxy configuration file is a directory: %s", f))
	} else if stat.Size() <= 0 {
		return nil, errors.New(fmt.Sprintf("proxy configuration file empty: %s", f))
	} else if stat.Size() > 1048576 {
		return nil, errors.New(fmt.Sprintf("proxy configuration file too large: %s", f))
	}
	out, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read proxy configuration file: %s: %s", f, err))
	}

	if err = json.Unmarshal(out, &m); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to unmarshal proxy configuration file: %s: %s", f, err))
	}
	// Sanitize the protocols so we can be case insensitive
	for protocol, v := range m {
		delete(m, protocol)
		m[strings.ToLower(protocol)] = v
	}
	return m, nil
}

/*
Find the proxy configured by environment variables for the given protocol and targetUrl.
If no proxy is found, or an error occurs, nil is returned.
Params:
	protocol: The proxy's protocol (i.e. https)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	proxy: A proxy is found through environment variables for the given protocol.
	nil: No proxy is found or an error occurs reading the environment variables.
*/
func (p *provider) readSystemEnvProxy(protocol string, targetUrl *url.URL) (Proxy) {
	keys := []string{
		strings.ToUpper(fmt.Sprintf(proxyKeyFormat, protocol)),
		strings.ToLower(fmt.Sprintf(proxyKeyFormat, protocol))}
	// TODO windows is case insensitive, waste of cycles here
	noProxyValues := map[string]string{
		noProxyKeyUpper: p.getEnv(noProxyKeyUpper),
		noProxyKeyLower: p.getEnv(noProxyKeyLower)}
	K: for _, key := range keys {
		proxy, err := p.parseEnvProxy(protocol, key)
		if err != nil {
			if !isNotFound(err) {
				log.Printf("[proxy.Provider.readSystemEnvProxy]: failed to parse \"%s\" value: %s\n", key, err)
			}
			continue
		}
		bypass := false
		for noProxyKey, proxyBypass := range noProxyValues {
			if proxyBypass == "" {
				continue
			}
			bypass = p.isProxyBypass(targetUrl, proxyBypass, ",")
			log.Printf("[proxy.Provider.readSystemEnvProxy]: \"%s\"=\"%s\", targetUrl=%s, bypass=%t", noProxyKey, proxyBypass, targetUrl, bypass)
			if bypass {
				continue K
			}
		}
		return proxy
	}
	return nil
}

/*
Return true if the given targetUrl should bypass a proxy for the given proxyBypass value and sep.
For example:
	("test.endpoint.rapid7.com", "rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", ".rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "*.rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "*", ",") -> true
	("test.endpoint.rapid7.com", "test.endpoint.rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "someHost,anotherHost", ",") -> false
	("test.endpoint.rapid7.com", "", ",") -> false
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	proxyBypass: The proxy bypass value.
	sep: The separator to use with the proxy bypass value.
Returns:
	true: The proxy should be bypassed for the given targetUrl
	false: Otherwise
*/
func (p *provider) isProxyBypass(targetUrl *url.URL, proxyBypass string, sep string) (bool) {
	targetHost, _, _ := SplitHostPort(targetUrl)
	for _, s := range strings.Split(proxyBypass, sep) {
		s = strings.TrimSpace(s)
		if s == "" {
			// No value
			continue
		} else if s == "<local>" {
			// Windows uses <local> for local domains
			if IsLoopbackHost(targetHost) {
				return true
			}
		}
		// Exact match
		if m, err :=  filepath.Match(s, targetHost) ; err != nil {
			return false
		} else if m {
			return true
		}
		// Prefix "* for wildcard matches (rapid7.com -> *.rapid7.com)
		if strings.Index(s, "*") != 0 {
			// (rapid7.com -> .rapid7.com)
			if strings.Index(s, ".") != 0 {
				s = "." + s
			}
			s = "*" + s
		}
		if m, err :=  filepath.Match(s, targetHost) ; err != nil {
			return false
		} else if m {
			return true
		}
	}
	return false
}

/*
Read the given environment variable by key and expected protocol, returning the proxy if it is valid.
Returns nil if no proxy is configured, or an error occurs.
Params:
	protocol: The proxy's expected protocol (i.e. https)
	key: The environment variable key
Returns:
	proxy: A proxy was found for the given environment variable key and is valid.
	false: Otherwise
*/
func (p *provider) parseEnvProxy(protocol string, key string) (Proxy, error) {
	proxyUrl, err := p.parseEnvURL(key)
	if err != nil {
		return nil, err
	}
	proxy, err := NewProxy(protocol, proxyUrl, fmt.Sprintf("Environment[%s]", key))
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

/*
Parse the optionally valid URL string from the given environment variable key's value.
Params:
	key: The name of the environment variable
Returns:
	url.URL: If the environment variable was populated, the parsed value. Otherwise nil.
	error: If the environment variable was populated, but we failed to parse it.
 */
func (p *provider) parseEnvURL(key string) (*url.URL, error) {
	value := strings.TrimSpace(p.getEnv(key))
	if value != "" {
		return ParseURL(value, "")
	}
	return nil, new(notFoundError)
}


type notFoundError struct {}

func (e notFoundError) Error() (string) {
	return "No proxy found"
}

/*
Returns:
	true: The error represents a Proxy not being found
	false: Otherwise
s*/
func isNotFound(e error) (bool) {
	switch e.(type) {
	case *notFoundError:
		return true
	case notFoundError:
		return true
	default:
		return false
	}
}
