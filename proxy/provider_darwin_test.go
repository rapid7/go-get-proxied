package proxy

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"strings"
	"testing"
)

const (
	ScutilDataHttpsHttp = "ScutilDataHttpsHttp"
	ScutilDataHttps = "ScutilDataHttps"
	ScutilDataHttp = "ScutilDataHttp"
)

var providerDarwinTestCases = []struct {
	testType string
	test string
}{
	{ScutilDataHttpsHttp,
	fmt.Sprintf("<dictionary> {\n  HTTPSEnable : 1\n HTTPSPort : %s\n  HTTPSProxy : %s\n" +
		" HTTPEnable : 1\n  HTTPPort : %s\n  HTTPProxy : %s\n}", "1234", "1.2.3.4", "1234", "1.2.3.4")},
	{ScutilDataHttpsHttp,
	fmt.Sprintf("         <dictionary> {   \n  HTTPSEnable:1    \nHTTPSPort :  %s\n  HTTPSProxy :  %s\n " +
		"HTTPEnable : 1   \n  HTTPPort :      %s\nHTTPProxy: %s   \n    }", "1234", "1.2.3.4",  "1234", "1.2.3.4")},
	{ScutilDataHttps,
	fmt.Sprintf("<dictionary> {\n  HTTPEnable : 0\n  HTTPSEnable : 1\n  HTTPSPort : %s\n  " +
		"HTTPSProxy : %s\n}", "1234", "1.2.3.4")},
	{ScutilDataHttps,
	fmt.Sprintf("<dictionary> {\n      HTTPEnable: 0\n  HTTPSEnable: 1\n  " +
		"HTTPSPort :        %s\n        HTTPSProxy:   %s\n}", "1234", "1.2.3.4")},
	{ScutilDataHttp,
	fmt.Sprintf("<dictionary> {\n  HTTPSEnable : 0\n  HTTPEnable : 1\n  HTTPPort : %s\n  " +
		"HTTPProxy : %s\n}", "1234", "1.2.3.4")},
	{ScutilDataHttp, 	fmt.Sprintf("<dictionary> {\n      HTTPSEnable: 0\n  HTTPEnable: 1\n  " +
		"HTTPPort :        %s\n        HTTPProxy:   %s\n}", "1234", "1.2.3.4")},
}

func getDarwinProviderTests(key string)([]string){
	var s []string
	for _, v := range providerDarwinTestCases {
		if v.testType == key {
			s = append(s, v.test)
		}
	}
	return s
}

/*
Below tests cover cases when both https and http proxies are present.
following tests are being performed:
- Test https and http proxies are not nil,
- Test https and http proxies match expected,
- Test for lower case and upper case, i.e. https/HTTPS/...,
- Test no errors are returned
*/
func TestParseScutildata_Read_HTTPS_HTTP(t *testing.T) {
	a := assert.New(t)

	c := newDarwinTestProvider()

	commands := getDarwinProviderTests(ScutilDataHttpsHttp)

	protocols := [4]string{"http", "https", "HTTP", "HTTPS"}
	for _, protocol := range protocols {
		for _, command := range commands {
			expectedProxy, err := c.parseScutildata(protocol, "echo", command)
			// test error is nil
			a.Nil(err)
			// test expected https proxy matches hardcoded proxy, test lowercase
			a.Equal(&proxy{src: "State:/Network/Global/Proxies", protocol: strings.ToLower(protocol), host: "1.2.3.4", port: 1234},
				expectedProxy)
			// test https and https proxies are not nil
			a.NotNil(c.parseScutildata(protocol, "echo", command))
		}
	}
}

/*
Below tests cover cases when only https proxy is present.
following tests are being performed:
- Test https proxy is not nil,
- Test http proxy is nil,
- Test https proxy match expected,
- Test for lower case and upper case, i.e. https/HTTPS/...,
- Test no errors are returned
*/
func TestParseScutildata_Read_HTTPS(t *testing.T) {
	a := assert.New(t)

	c := newDarwinTestProvider()

	commands := getDarwinProviderTests(ScutilDataHttps)

	protocols := [4]string{"http", "https", "HTTP", "HTTPS"}
	for _, protocol := range protocols {
		for _, command := range commands {
			if strings.ToLower(protocol) == "https" {
				expectedProxy, err := c.parseScutildata(protocol, "echo", command)
				// test error is nil
				a.Nil(err)
				// test expected https proxy matches hardcoded proxy
				a.Equal(&proxy{src: "State:/Network/Global/Proxies", protocol: strings.ToLower(protocol), host: "1.2.3.4", port: 1234},
					expectedProxy)
				// test https proxy is not nil
				a.NotNil(c.parseScutildata(protocol, "echo", command))
			}else{
				// test http proxy is nil
				a.Nil(c.parseScutildata(protocol, "echo", command))
			}

		}
	}
}

/*
Below tests cover cases when only http proxy is present.
following tests are being performed:
- Test http proxy is not nil,
- Test https proxy is nil,
- Test http proxy match expected,
- Test for lower case and upper case, i.e. https/HTTPS/...,
- Test no errors are returned
*/
func TestParseScutildata_Read_HTTP(t *testing.T) {
	a := assert.New(t)

	c := newDarwinTestProvider()

	commands := getDarwinProviderTests(ScutilDataHttp)

	protocols := [4]string{"http", "https", "HTTP", "HTTPS"}
	for _, protocol := range protocols {
		for _, command := range commands {
			if strings.ToLower(protocol) == "http" {
				expectedProxy, err := c.parseScutildata(protocol, "echo", command)
				// test error is nil
				a.Nil(err)
				// test expected http proxy matches hardcoded proxy
				a.Equal(&proxy{src: "State:/Network/Global/Proxies", protocol: strings.ToLower(protocol), host: "1.2.3.4", port: 1234},
					expectedProxy)
				// test http proxy is not nil
				a.NotNil(c.parseScutildata(protocol, "echo", command))
			}else{
				// test https proxy is nil
				a.Nil(c.parseScutildata(protocol, "echo", command))
			}
		}
	}
}

/*
Tests whether the timeout property functions as expected
*/
func TestExecCommandsHandledProperly(t *testing.T) {
	a := assert.New(t)

	c := newDarwinTestProvider()

	expectedProxy, err := c.parseScutildata("", "exit", "")

	a.Equal(isTimedOut(err), true)
	a.Equal(expectedProxy, nil)
}

func newDarwinTestProvider() (*providerDarwin) {
	Cmd := func (ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, name, args...)
	}
	c := new(providerDarwin)
	c.proc = Cmd
	return c
}