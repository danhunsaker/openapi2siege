package main

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"
)

type requestData struct {
	MediaType string
	Payload   string
}

type urlData struct {
	URL       url.URL
	Method    string
	MediaType string
	Payload   string
	Cookies   []*http.Cookie
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

func (l urlList) StringByMediaType(mediaType string) string {
	return strings.Join(mapSlice(l, func(data urlData) string {
		if data.MediaType != mediaType && data.MediaType != "" {
			return ""
		}

		return data.String()
	}), "\n")
}

func (l urlList) MediaTypes() []string {
	return uniqueSlice(mapSlice(l, func(data urlData) string {
		return data.MediaType
	}))
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
		v := f(e)
		if !reflect.ValueOf(v).IsZero() {
			res = append(res, v)
		}
	}

	return res
}

func uniqueSlice[T comparable](data []T) []T {
	filter := make(map[T]bool)
	res := make([]T, 0)

	for _, e := range data {
		if _, isSet := filter[e]; !isSet {
			filter[e] = true
			res = append(res, e)
		}
	}

	return res
}
