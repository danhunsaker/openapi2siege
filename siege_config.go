package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type SiegeConfig struct {
	Variables       SiegeVars     `siege:"-"`
	Verbose         SiegeBoolTF   `siege:"verbose"`
	Color           SiegeBoolOO   `siege:"color"`
	Quiet           SiegeBoolTF   `siege:"quiet"`
	JsonOutput      SiegeBoolTF   `siege:"json_output"`
	ShowLogfile     SiegeBoolTF   `siege:"show-logfile"`
	Logging         SiegeBoolTF   `siege:"logging"`
	LogFile         string        `siege:"logfile,omitempty"`
	GetMethod       string        `siege:"gmethod"`
	UseParser       SiegeBoolTF   `siege:"parser"`
	NoFollow        stringSlice   `siege:"nofollow"`
	CSVOutput       SiegeBoolTF   `siege:"csv,omitempty"`
	ShowTimestamp   SiegeBoolTF   `siege:"timestamp,omitempty"`
	ShowFullURL     SiegeBoolTF   `siege:"fullurl,omitempty"`
	ShowID          SiegeBoolTF   `siege:"display-id,omitempty"`
	ThreadLimit     int           `siege:"limit"`
	Protocol        string        `siege:"protocol"`
	Chunked         SiegeBoolTF   `siege:"chunked"`
	Cache           SiegeBoolTF   `siege:"cache"`
	Connection      string        `siege:"connection"`
	Concurrent      int           `siege:"concurrent"`
	Duration        time.Duration `siege:"time,omitempty"`
	Reps            int           `siege:"reps,omitempty"`
	Delay           float64       `siege:"delay"`
	UrlFile         string        `siege:"file,omitempty"`
	SingleUrl       string        `siege:"url,omitempty"`
	Timeout         int           `siege:"timeout,omitempty"`
	ExpireSession   SiegeBoolTF   `siege:"expire-session,omitempty"`
	UseCookies      SiegeBoolTF   `siege:"cookies"`
	AllowedFailures int           `siege:"failures,omitempty"`
	UseRandomUrls   SiegeBoolTF   `siege:"internet"`
	BenchmarkMode   SiegeBoolTF   `siege:"benchmark"`
	UserAgent       string        `siege:"user-agent,omitempty"`
	AcceptEncoding  string        `siege:"accept-encoding"`
	EscapeUrls      SiegeBoolTF   `siege:"url-escaping"`
	LoginInfo       SiegeCreds    `siege:"login,omitempty"`
	LoginUrls       urlList       `siege:"login-url,omitempty"`
	FtpLoginInfo    SiegeCreds    `siege:"ftp-login,omitempty"`
	FtpUnique       SiegeBoolTF   `siege:"unique"`
	SslUserCert     string        `siege:"ssl-cert,omitempty"`
	SslUserKey      string        `siege:"ssl-key,omitempty"`
	SslTimeout      int           `siege:"ssl-timeout,omitempty"`
	SslCiphers      string        `siege:"ssl-ciphers,omitempty"`
	ProxyHost       string        `siege:"proxy-host,omitempty"`
	ProxyPort       int16         `siege:"proxy-port,omitempty"`
	ProxyLogin      SiegeCreds    `siege:"proxy-login,omitempty"`
	FollowRedirects SiegeBoolTF   `siege:"follow-location"`
	Headers         http.Header   `siege:"header,omitempty"`
}

// NewSiegeConfig creates a SiegeConfig with the same defaults as Siege itself
func NewSiegeConfig() *SiegeConfig {
	noFollowList := []string{
		"ad.doubleclick.net",
		"pagead2.googlesyndication.com",
		"ads.pubsqrd.com",
		"ib.adnxs.com",
	}

	return &SiegeConfig{
		Verbose:         true,
		Color:           true,
		Quiet:           false,
		JsonOutput:      true,
		ShowLogfile:     true,
		Logging:         false,
		GetMethod:       "HEAD",
		UseParser:       true,
		NoFollow:        noFollowList,
		ThreadLimit:     255,
		Protocol:        "HTTP/1.1",
		Chunked:         true,
		Cache:           false,
		Connection:      "close",
		Concurrent:      25,
		Delay:           0,
		UseCookies:      true,
		UseRandomUrls:   false,
		BenchmarkMode:   false,
		AcceptEncoding:  "gzip, deflate",
		EscapeUrls:      true,
		FtpUnique:       true,
		FollowRedirects: true,
		Headers:         make(http.Header),
	}
}

func (c *SiegeConfig) String() string {
	writer := new(strings.Builder)

	writer.WriteString(c.Variables.String())

	reflected := reflect.ValueOf(*c)
	for i := 0; i < reflected.Type().NumField(); i++ {
		fieldType := reflected.Type().Field(i)
		fieldVal := reflected.Field(i)

		tag := fieldType.Tag.Get("siege")
		if tag == "-" {
			continue
		}

		name, opts := parseTag(tag)

		if opts.OmitEmpty && fieldVal.IsZero() {
			continue
		}

		switch fieldType.Type.Name() {
		case "string":
			writer.WriteString(fmt.Sprintf("%s = %s\n", name, fieldVal.String()))
		case "stringSlice":
			for idx := 0; idx < fieldVal.Len(); idx++ {
				writer.WriteString(fmt.Sprintf("%s = %s\n", name, fieldVal.Index(idx).String()))
			}
		case "int", "int16":
			writer.WriteString(fmt.Sprintf("%s = %d\n", name, fieldVal.Int()))
		case "float64":
			writer.WriteString(fmt.Sprintf("%s = %f\n", name, fieldVal.Float()))
		case "SiegeBoolTF", "SiegeBoolOO", "SiegeCreds", "urlList", "Duration":
			result := fieldVal.MethodByName("String").Call([]reflect.Value{})
			writer.WriteString(fmt.Sprintf("%s = %s\n", name, result[0].String()))
		case "Header":
			for _, key := range fieldVal.MapKeys() {
				valueList := fieldVal.MapIndex(key)
				for idx := 0; idx < valueList.Len(); idx++ {
					writer.WriteString(fmt.Sprintf("%s = %s: %s\n", name, key.String(), valueList.Index(idx).String()))
				}
			}
		default:
			log.Printf("Failed to export Siege config: got a '%s' when trying to export '%s'", fieldType.Type.Name(), name)
		}
	}

	return writer.String()
}

type SiegeVars map[string]string

func (v SiegeVars) String() string {
	output := ""

	for variable, value := range v {
		output += fmt.Sprintf("%s = %s\n", variable, value)
	}

	return output
}

type SiegeCreds struct {
	User     string
	Password string
	Realm    string
}

func (c SiegeCreds) String() string {
	if c.Realm != "" {
		return fmt.Sprintf("%s:%s:%s", c.User, c.Password, c.Realm)
	}

	return fmt.Sprintf("%s:%s", c.User, c.Password)
}

type SiegeBoolTF bool

func (b SiegeBoolTF) String() string {
	if b {
		return "true"
	}

	return "false"
}

type SiegeBoolOO bool

func (b SiegeBoolOO) String() string {
	if b {
		return "on"
	}

	return "off"
}

type stringSlice []string

type siegeConfigFieldOptions struct {
	OmitEmpty bool
}

func parseTag(tag string) (string, siegeConfigFieldOptions) {
	parsedTag := strings.Split(tag, ",")
	name := parsedTag[0]
	opts := parseOpts(parsedTag[1:])

	return name, opts
}

func parseOpts(opts []string) siegeConfigFieldOptions {
	out := siegeConfigFieldOptions{}

	for _, opt := range opts {
		if opt == "omitempty" {
			out.OmitEmpty = true
		}
	}

	return out
}
