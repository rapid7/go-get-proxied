go-get-proxied - Cross platform proxy configurations
================================

[![Build Status](https://travis-ci.org/rapid7/go-get-proxied.svg)](https://travis-ci.org/rapid7/go-get-proxied) [![Go Report Card](https://goreportcard.com/badge/github.com/rapid7/go-get-proxied)](https://goreportcard.com/report/github.com/rapid7/go-get-proxied)

Go code (golang) package which facilitates the retrieval of system proxy configurations.

#### Installation

* Install this package using `go get github.com/rapid7/go-get-proxied`
* Or, simply import `go get github.com/rapid7/go-get-proxied`, and use `dep ensure` to include it in your package

#### Usage: 

```go
package main
import (
    "fmt"
    "github.com/rapid7/go-get-proxied/proxy"
)
func main() {
    p := proxy.NewProvider("").Get("https", "https://rapid7.com")
    if p != nil {
        fmt.Printf("Found proxy: %s\n", p)
    }
}
```

#### Command Line Usage:
```bash
> ./go-get-proxied -h
Usage of ./go-get-proxied:
  -c string
    	Optional. Path to configuration file.
  -j	Optional. If a proxy is found, write it as JSON instead of a URL.
  -p string
    	Optional. The proxy protocol you wish to lookup. Default: https (default "https")
  -t string
    	Optional. Target URL which the proxy will be used for. Default: *
  -v	Optional. If set, log content will be sent to stderr.
```
```bash
> netsh winhttp set proxy testProxy:8999

Current WinHTTP proxy settings:

    Proxy Server(s) :  testProxy:8999
    Bypass List     :  (none)
> ./go-get-proxied
//testProxy:8999
> ./go-get-proxied -j
{
   "host": "testProxy",
   "password": null,
   "port": 8999,
   "protocol": "",
   "src": "WinHTTP:WinHttpDefault",
   "username": null
}
```
```bash
> echo '{"https":"http://testProxy:8999"}' > proxy.config
> ./go-get-proxied -c proxy.config
http://testProxy:8999
> ./go-get-proxied -c proxy.config -j
{
   "host": "testProxy",
   "password": null,
   "port": 8999,
   "protocol": "http",
   "src": "ConfigurationFile",
   "username": null
}
```

#### Configuration:

The priority of retrieval is the following.
-  **Windows**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.
   - Internet Options: Automatically detect settings (`WPAD`)
   - Internet Options: Use automatic configuration script (`PAC`)
   - Internet Options: Manual proxy server
   - WINHTTP: (`netsh winhttp`)
- **Linux**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.
- **MacOS**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.
   - Network Settings: `scutil`
