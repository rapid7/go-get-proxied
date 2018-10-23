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
