// +build windows

// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
package winhttp

import (
	"golang.org/x/sys/windows"
	"unicode/utf16"
	"unsafe"
)

//noinspection SpellCheckingInspection,GoNameStartsWithPackageName,GoSnakeCaseUsage
const (
	// AutoProxyOptions.DwFlags
	WINHTTP_AUTOPROXY_AUTO_DETECT		= 0x00000001
	WINHTTP_AUTOPROXY_CONFIG_URL		= 0x00000002
	// AutoProxyOptions.DwAutoDetectFlags
	WINHTTP_AUTO_DETECT_TYPE_DHCP		= 0x00000001
	WINHTTP_AUTO_DETECT_TYPE_DNS_A 		= 0x00000002
	// AutoProxyOptions.dwAccessType
	WINHTTP_ACCESS_TYPE_NO_PROXY 		= 0x00000001
	// Set a sane ceiling in case we don't find \x00\x00
	// Maximum number of bytes for any returned lpwstr
	lpwstrMaxBytes = 1024*512
)

//noinspection SpellCheckingInspection
type (
	Lpwstr 		*uint16
	Dword 		uint32
	HInternet 	uintptr
	Allocated interface {
		Free() (error)
	}
	AutoProxyOptions struct {
		DwFlags                Dword
		DwAutoDetectFlags      Dword
		LpszAutoConfigUrl      Lpwstr
		lpvReserved            uintptr
		dwReserved             uint32
		FAutoLogonIfChallenged bool
	}
	ProxyInfo struct {
		DwAccessType    Dword
		LpszProxy       Lpwstr
		LpszProxyBypass Lpwstr
	}
	CurrentUserIEProxyConfig struct {
		FAutoDetect			bool
		LpszAutoConfigUrl 	Lpwstr
		LpszProxy			Lpwstr
		LpszProxyBypass 	Lpwstr
	}
)

//noinspection SpellCheckingInspection
func Open(pszAgentW Lpwstr, dwAccessType Dword, pszProxyW Lpwstr, pszProxyBypassW Lpwstr, dwFlags Dword) (HInternet, error) {
	if err := openP.Find() ; err != nil {
		return 0, err
	}
	r, _, err := openP.Call(
		uintptr(unsafe.Pointer(pszAgentW)),
		uintptr(dwAccessType),
		uintptr(unsafe.Pointer(pszProxyW)),
		uintptr(unsafe.Pointer(pszProxyBypassW)),
		uintptr(dwFlags),
	)
	if rNil(r) {
		return 0, err
	}
	return HInternet(r), nil
}

func CloseHandle(hInternet HInternet) (error) {
	if err := closeHandleP.Find() ; err != nil {
		return err
	}
	r, _, err := closeHandleP.Call(uintptr(hInternet))
	if rTrue(r) {
		return nil
	}
	return err
}

func SetTimeouts(hInternet HInternet, nResolveTimeout int, nConnectTimeout int, nSendTimeout int, nReceiveTimeout int) (error) {
	if err := setTimeoutsP.Find() ; err != nil {
		return err
	}
	r, _, err := setTimeoutsP.Call(
		uintptr(hInternet),
		uintptr(nResolveTimeout),
		uintptr(nConnectTimeout),
		uintptr(nSendTimeout),
		uintptr(nReceiveTimeout))
	if rTrue(r) {
		return nil
	}
	return err
}

//noinspection SpellCheckingInspection
func GetProxyForUrl(hInternet HInternet, lpcwszUrl Lpwstr, winhttpAutoProxyOptions *AutoProxyOptions) (*ProxyInfo, error) {
	if err := getProxyForUrlP.Find() ; err != nil {
		return nil, err
	}
	p := new(ProxyInfo)
	r, _, err := getProxyForUrlP.Call(
		uintptr(hInternet),
		uintptr(unsafe.Pointer(lpcwszUrl)),
		uintptr(unsafe.Pointer(winhttpAutoProxyOptions)),
		uintptr(unsafe.Pointer(p)))
	if rTrue(r) {
		return p, nil
	}
	return nil, err
}

//noinspection SpellCheckingInspection
func GetIEProxyConfigForCurrentUser() (*CurrentUserIEProxyConfig, error) {
	if err := getIEProxyConfigForCurrentUserP.Find() ; err != nil {
		return nil, err
	}
	p := new(CurrentUserIEProxyConfig)
	r, _, err := getIEProxyConfigForCurrentUserP.Call(uintptr(unsafe.Pointer(p)))
	if rTrue(r) {
		return p, nil
	}
	return nil, err
}

func GetDefaultProxyConfiguration() (*ProxyInfo, error) {
	pInfo := new(ProxyInfo)
	if err := getDefaultProxyConfigurationP.Find() ; err != nil {
		return nil, err
	}
	r, _, err := getDefaultProxyConfigurationP.Call(uintptr(unsafe.Pointer(pInfo)))
	if rTrue(r) {
		return pInfo, nil
	}
	return nil, err
}

//noinspection SpellCheckingInspection
func LpwstrToString(d Lpwstr) (string) {
	if d == nil {
		return ""
	}
	s := make([]uint16, 0, 256)
	p := uintptr(unsafe.Pointer(d))
	pMax := p + lpwstrMaxBytes
	for ; p < pMax ; p += 2 {
		c := *(*uint16)(unsafe.Pointer(p))
		// NUL char is EOF
		if c == 0 {
			return string(utf16.Decode(s))
		}
		s = append(s, c)
	}
	return ""
}

//noinspection SpellCheckingInspection
func StringToLpwstr(s string) (*uint16) {
	if s == "" {
		return nil
	}
	// If s contains \x00, we'll just silently return nil, better than a panic
	r, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return r
}

//noinspection SpellCheckingInspection
func (p *ProxyInfo) Free() (error) {
	if p == nil {
		return nil
	}
	rerr := globalFree(p.LpszProxy)
	if err := globalFree(p.LpszProxyBypass) ; rerr == nil && err != nil {
		rerr = err
	}
	return rerr
}

//noinspection SpellCheckingInspection
func (p *CurrentUserIEProxyConfig) Free() (error) {
	if p == nil {
		return nil
	}
	rerr := globalFree(p.LpszAutoConfigUrl)
	if err := globalFree(p.LpszProxy) ; rerr == nil && err != nil {
		rerr = err
	}
	if err := globalFree(p.LpszProxyBypass) ; rerr == nil && err != nil {
		rerr = err
	}
	return rerr
}


/************* BEGIN PRIVATE IMPL *************/

//noinspection SpellCheckingInspection
var (
	kd = windows.NewLazySystemDLL("kernel32.dll")
	globalFreeP = kd.NewProc("GlobalFree")
	whd = windows.NewLazySystemDLL("winhttp.dll")
	openP = whd.NewProc("WinHttpOpen")
	closeHandleP = whd.NewProc("WinHttpCloseHandle")
	setTimeoutsP = whd.NewProc("WinHttpSetTimeouts")
	getProxyForUrlP = whd.NewProc("WinHttpGetProxyForUrl")
	getIEProxyConfigForCurrentUserP = whd.NewProc("WinHttpGetIEProxyConfigForCurrentUser")
	getDefaultProxyConfigurationP = whd.NewProc("WinHttpGetDefaultProxyConfiguration")
)

func globalFree(hMem *uint16) (error) {
	if hMem == nil {
		return nil
	}
	if err := globalFreeP.Find() ; err != nil {
		return err
	}
	r, _, err := globalFreeP.Call(uintptr(unsafe.Pointer(hMem)))
	if rNil(r) {
		return nil
	}
	return err
}

func rNil(r uintptr) (bool) {
	return r == 0
}

func rTrue(r uintptr) (bool) {
	return r == 1
}
