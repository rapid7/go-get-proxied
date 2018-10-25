go-get-proxied - Cross platform proxy configurations
================================

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
> ./proxymain -h                                                                                                                  Tue Oct 23 10:01:44 2018
Usage of ./proxymain:
  -c string
    	Optional. Path to configuration file.
  -p string
    	Optional. The proxy protocol you wish to lookup. Default: https (default "https")
  -q	Optional. Quiet mode; only write the URL of a proxy to stdout (if found). Default: false
  -t string
    	Optional. Target URL which the proxy will be used for. Default: *
```
```bash
> netsh winhttp set proxy testProxy:8999

Current WinHTTP proxy settings:

    Proxy Server(s) :  testProxy:8999
    Bypass List     :  (none)
> ./proxymain.exe
Proxy: WinHTTP:WinHttpDefault|https://testProxy:8999
Proxy JSON: {
   "host": "testProxy",
   "password": null,
   "port": 8999,
   "protocol": "https",
   "src": "WinHTTP:WinHttpDefault",
   "username": null
}
> ./proxy_main.exe -q
https://testProxy:8999
```
```bash
> echo '{"https":"testProxy:8999"}' > proxy.config
> ./proxymain -c proxy.config -q
https://testProxy:8999
> ./proxymain -c proxy.config
Proxy: ConfigurationFile|https://testProxy:8999
Proxy JSON: {
   "host": "testProxy",
   "password": null,
   "port": 8999,
   "protocol": "https",
   "src": "ConfigurationFile",
   "username": null
}
```

#### Configuration:

The priority of retrieval is the following.
-  **Windows**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY` and `NO_PROXY`
   - Internet Options: Automatically detect settings (`WPAD`)
   - Internet Options: Use automatic configuration script (`PAC`)
   - Internet Options: Manual proxy server
   - WINHTTP: (`netsh winhttp`)
- **Linux**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY` and `NO_PROXY`
- **MacOS**:
   - Configuration File
   - Environment Variable: `HTTPS_PROXY` and `NO_PROXY`
   - Network Settings: `scutil`
