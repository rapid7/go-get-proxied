// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
package proxy

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

var dataNewProxy = []struct {
	protocol	string
	u 			*url.URL
	expectP		Proxy
	expectErr	error
}{
	// All input
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:8999", User:url.UserPassword("user1", "password1")},
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.UserPassword("user1", "password1"), src:"Test"}, nil,
	},
	// No password
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:8999", User:url.User("user1")},
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.User("user1"), src:"Test"}, nil,
	},
	// No user
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:8999"},
		&proxy{protocol:"https", host:"testProxy", port:8999, user:nil, src:"Test"}, nil,
	},
	// No port
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy"},
		&proxy{protocol:"https", host:"testProxy", port:8443, user:nil, src:"Test"}, nil,
	},
	// No port - Default protocol
	{
		"gopher", &url.URL{Scheme:"gopher", Host:"testProxy"},
		&proxy{protocol:"gopher", host:"testProxy", port:8080, user:nil, src:"Test"}, nil,
	},
	// 0 port
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:0"},
		&proxy{protocol:"https", host:"testProxy", port:8443, user:nil, src:"Test"}, nil,
	},
	// No URL protocol
	{
		"https", &url.URL{Host:"testProxy"},
		&proxy{protocol:"https", host:"testProxy", port: 8443, src:"Test"}, nil,
	},
	// Uppercase expected protocol
	{
		" HTTPS ", &url.URL{Scheme:"https", Host:"testProxy"},
		&proxy{protocol:"https", host:"testProxy", port:8443, user:nil, src:"Test"}, nil,
	},
	// Uppercase and whitespace URL protocol
	{
		" https ", &url.URL{Scheme:"  HTTPS  ", Host:"testProxy"},
		&proxy{protocol:"https", host:"testProxy", port:8443, user:nil, src:"Test"}, nil,
	},
	// No expected protocol
	{
		" ", &url.URL{Scheme:"https", Host:"testProxy"},
		&proxy{protocol:"https", host:"testProxy", port:8443, user:nil, src:"Test"}, nil,
	},
	// Mis-matched protocol
	{
		"http", &url.URL{Scheme:"https", Host:""},
		nil, errors.New("expected protocol \"http\", got \"https\""),
	},
	// Invalid port
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:testPort"},
		nil, errors.New("SplitHostPort testProxy:testPort: strconv.ParseUint: parsing \"testPort\": invalid syntax"),
	},
	// Negative port
	{
		"https", &url.URL{Scheme:"https", Host:"testProxy:-1"},
		nil, errors.New("SplitHostPort testProxy:-1: strconv.ParseUint: parsing \"-1\": invalid syntax"),
	},
	// Empty host
	{
		"https", &url.URL{Scheme:"https", Host:""},
		nil, errors.New("empty host"),
	},
	// Nil URL
	{
		"https", nil,
		nil, errors.New("nil URL"),
	},
}
func TestNewProxy(t *testing.T) {
	for _, tt := range dataNewProxy {
		var tName = tt.protocol + " "
		if tt.u == nil {
			tName = tName + "nil"
		} else {
			tName = tName + tt.u.String()
		}
		t.Run(tName, func(t *testing.T) {
			a := assert.New(t)
			p, err := NewProxy(tt.protocol, tt.u, "Test")
			if tt.expectP == nil {
				a.Nil(p)
			} else {
				a.Equal(tt.expectP, p)
			}
			if tt.expectErr == nil {
				a.Nil(err)
			} else {
				if a.NotNil(err) {
					a.Equal(tt.expectErr.Error(), err.Error())
				}
			}
		})
	}
}

var dataUsername = []struct {
	u				*url.Userinfo
	expectUsername	string
	expectExists	bool
} {
	{
		url.User("user1"), "user1", true,
	},
	{
		url.UserPassword("user1", "password1"), "user1", true,
	},
	{
		url.User(""), "", true,
	},
	{
		nil, "", false,
	},
}

func TestProxy_Username(t *testing.T) {
	for _, tt := range dataUsername {
		var tName string
		if tt.u == nil {
			tName = "nil"
		} else {
			tName = tt.u.String()
		}
		t.Run(tName, func(t *testing.T) {
			a := assert.New(t)
			username, exists := (&proxy{user:tt.u}).Username()
			a.Equal(tt.expectUsername, username)
			a.Equal(tt.expectExists, exists)
		})
	}
}

var dataPassword = []struct {
	u				*url.Userinfo
	expectPassword	string
	expectExists	bool
} {
	{
		url.User("user1"), "", false,
	},
	{
		url.UserPassword("user1", "password1"), "password1", true,
	},
	{
		url.UserPassword("user1", ""), "", true,
	},
	{
		url.User(""), "", false,
	},
	{
		nil, "", false,
	},
}

func TestProxy_Password(t *testing.T) {
	for _, tt := range dataPassword {
		var tName string
		if tt.u == nil {
			tName = "nil"
		} else {
			tName = tt.u.String()
		}
		t.Run(tName, func(t *testing.T) {
			a := assert.New(t)
			username, exists := (&proxy{user:tt.u}).Password()
			a.Equal(tt.expectPassword, username)
			a.Equal(tt.expectExists, exists)
		})
	}
}

var dataURL = []struct {
	p			Proxy
	expectU		*url.URL
	expectStr	string
} {
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.UserPassword("user1", "password1"), src:"Test"},
		&url.URL{Scheme:"https", Host:"testProxy:8999", User:url.UserPassword("user1", "password1")},
		"https://user1:password1@testProxy:8999",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.User("user1"), src:"Test"},
		&url.URL{Scheme:"https", Host:"testProxy:8999", User:url.User("user1")},
		"https://user1@testProxy:8999",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999},
		&url.URL{Scheme:"https", Host:"testProxy:8999"},
		"https://testProxy:8999",
	},
	{
		&proxy{port:0},
		&url.URL{Host:":0"},
		"//:0",
	},
}

func TestProxy_URL(t *testing.T) {
	for _, tt := range dataURL {
		t.Run(tt.p.String(), func(t *testing.T) {
			a := assert.New(t)
			u := tt.p.URL()
			if a.Equal(tt.expectU, u) {
				a.Equal(tt.expectStr, u.String())
			}
		})
	}
}

var dataString = []struct {
	p		Proxy
	expect 	string
} {
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.UserPassword("user1", "password1"), src:"Test"},
		"Test|https://<username>:<password>@testProxy:8999",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.User("user1"), src:"Test"},
		"Test|https://<username>@testProxy:8999",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.User("user1"), src:"Test"},
		"Test|https://<username>@testProxy:8999",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, src:"Test"},
		"Test|https://testProxy:8999",
	},
	{
		&proxy{},
		"|://:0",
	},
}

func TestProxy_String(t *testing.T) {
	for _, tt := range dataString {
		t.Run(tt.p.String(), func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tt.expect, tt.p.String())
		})
	}
}

var dataProxyMarshalJSON = []struct {
	p		Proxy
	expect 	string
} {
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.UserPassword("user1", "password1"), src:"Test"},
		"{\"host\":\"testProxy\",\"password\":\"password1\",\"port\":8999,\"protocol\":\"https\",\"src\":\"Test\",\"username\":\"user1\"}",
	},
	{
		&proxy{protocol:"https", host:"testProxy", port:8999, user:url.User("user1"), src:"Test"},
		"{\"host\":\"testProxy\",\"password\":null,\"port\":8999,\"protocol\":\"https\",\"src\":\"Test\",\"username\":\"user1\"}",
	},
	{
		&proxy{},
		"{\"host\":\"\",\"password\":null,\"port\":0,\"protocol\":\"\",\"src\":\"\",\"username\":null}",
	},
}

func TestProxy_MarshalJSON(t *testing.T) {
	for _, tt := range dataProxyMarshalJSON {
		t.Run(tt.p.String(), func(t *testing.T) {
			a := assert.New(t)
			b, err := tt.p.MarshalJSON()
			if a.Nil(err) {
				a.Equal(tt.expect, string(b[:]))
			}
		})
	}
}