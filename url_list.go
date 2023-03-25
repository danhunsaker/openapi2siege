package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"
)

type urlData struct {
	URL     url.URL
	Method  string
	Payload string
	Cookies []*http.Cookie
}

type urlList []urlData

func (d urlData) String() string {
	if d.Method == "GET" {
		return d.URL.String()
	}

	return fmt.Sprintf("%s %s %s", d.URL.String(), d.Method, d.Payload)
}

func (l urlList) String() string {
	return strings.Join(mapSlice(l, func(data urlData) string {
		return data.String()
	}), "\n")
}

func (l urlList) CookieJar(file string) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{Filename: file})
	if err != nil {
		return nil, err
	}

	for _, data := range l {
		if len(data.Cookies) > 0 {
			jar.SetCookies(&data.URL, data.Cookies)
		}
	}

	return jar, nil
}

func mapSlice[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}
