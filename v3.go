package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"
)

func handleV3Spec(c *cli.Context, spec *libopenapi.DocumentModel[v3.Document]) (urlList, *SiegeConfig, error) {
	urls := urlList{}
	conf := NewSiegeConfig()
	paths := spec.Model.Paths.PathItems

	baseUrl, err := getV3BaseUrl(c, spec.Model.Servers)
	if err != nil {
		return nil, nil, err
	}

	pathsConfig, isType := c.Generic("paths").(PathsConfig)
	if !isType {
		return nil, nil, fmt.Errorf("Paths not configured.\n\tNeed `paths.{path}.{method}.params.{name}` and/or `paths.{path}.{method}.payloads.{mediaType}`\n")
	}

	// Iterate paths in the same order every invocation
	pathList := maps.Keys(paths)
	sort.Strings(pathList)

	for _, rawPath := range pathList {
		pathData := paths[rawPath]
		pathConfig, exists := pathsConfig[rawPath]
		if !exists {
			return nil, nil, fmt.Errorf("Path `%s` not configured.\n\tNeed `paths.%s.{method}.params.{name}` and/or `paths.%s.{method}.payloads.{mediaType}`\n", rawPath, rawPath, rawPath)
		}

		if pathData.Get != nil && !*pathData.Get.Deprecated {
			urls, err = getV3RequestNoPayload(c, "get", rawPath, baseUrl, urls, pathData.Get, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Post != nil && !*pathData.Post.Deprecated {
			urls, err = getV3RequestWithPayload(c, "post", rawPath, baseUrl, urls, pathData.Post, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Delete != nil && !*pathData.Delete.Deprecated {
			urls, err = getV3RequestWithPayload(c, "delete", rawPath, baseUrl, urls, pathData.Delete, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Patch != nil && !*pathData.Patch.Deprecated {
			urls, err = getV3RequestWithPayload(c, "patch", rawPath, baseUrl, urls, pathData.Patch, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Put != nil && !*pathData.Put.Deprecated {
			urls, err = getV3RequestWithPayload(c, "put", rawPath, baseUrl, urls, pathData.Put, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Trace != nil && !*pathData.Trace.Deprecated {
			log.Printf("TRACE operations are unsupported by Siege; your tests will be incomplete\n\tSkipping TRACE for %s\n", rawPath)
		}

		if pathData.Head != nil && !*pathData.Head.Deprecated {
			urls, err = getV3RequestNoPayload(c, "head", rawPath, baseUrl, urls, pathData.Head, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		if pathData.Options != nil && !*pathData.Options.Deprecated {
			urls, err = getV3RequestWithPayload(c, "options", rawPath, baseUrl, urls, pathData.Options, pathConfig)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	for _, sec := range spec.Model.Security {
		auth, isType := c.Generic("auth").(AuthConfig)
		if !isType {
			return nil, nil, fmt.Errorf("Auth not configured!\n\tNeed `auth.{scheme_name}.*\n")
		}

		for name := range sec.Requirements {
			scheme, exists := spec.Model.Components.SecuritySchemes[name]
			if !exists {
				return nil, nil, fmt.Errorf("Auth scheme %s not configured\n\tNeed `auth.%s.*\n", name, name)
			}

			switch scheme.Type {
			case "apiKey":
				key, exists := auth[name]["apikey"]
				if !exists {
					return nil, nil, fmt.Errorf("API Key not configured for %s scheme\n\tNeed `auth.%s.apikey`\n", name, name)
				}

				switch scheme.In {
				case "query":
					for _, url := range urls {
						query := url.URL.Query()
						query.Add(scheme.Name, key)
						url.URL.RawQuery = query.Encode()
					}
				case "header":
					conf.Headers.Add(scheme.Name, key)
				case "cookie":
					for _, url := range urls {
						cookie := http.Cookie{
							Name:  scheme.Name,
							Value: key,
						}
						url.Cookies = append(url.Cookies, &cookie)
					}
				}
			case "http":
				creds, exists := auth[name]["creds"]
				if !exists {
					return nil, nil, fmt.Errorf("Credentials not configured for %s scheme\n\tNeed `auth.%s.creds`\n", name, name)
				}

				switch scheme.Scheme {
				case "basic":
					basicCreds := strings.SplitN(creds, ":", 2)
					if len(basicCreds) < 2 {
						return nil, nil, fmt.Errorf("Credentials incorrect for %s scheme\n\tNeed `auth.%s.creds` to be `{user}:{pass}` or `{user}:{pass}:{realm}\n", name, name)
					}
					conf.LoginInfo.User = basicCreds[0]
					conf.LoginInfo.Password = basicCreds[1]
					conf.LoginInfo.Realm = basicCreds[2]
				case "digest":
					digestCreds := strings.SplitN(creds, ":", 2)
					if len(digestCreds) < 2 {
						return nil, nil, fmt.Errorf("Credentials incorrect for %s scheme\n\tNeed `auth.%s.creds` to be `{user}:{pass}` or `{user}:{pass}:{realm}\n", name, name)
					}
					conf.LoginInfo.User = digestCreds[0]
					conf.LoginInfo.Password = digestCreds[1]
					conf.LoginInfo.Realm = digestCreds[2]
				case "bearer":
					if creds != "command" {
						conf.Headers.Add("Authorization", fmt.Sprintf("Bearer %s", creds))
						log.Println("The HTTP auth scheme `bearer` is supported on a best-effort basis.\n\tSiege does NOT actively support bearer tokens; expiration handling is up to you.")
					} else {
						conf.Headers.Add("Authorization", "Bearer ${OA2S_TOKEN}")
						log.Println("The HTTP auth scheme `bearer` is supported on a best-effort basis.\n\tSiege does NOT actively support bearer tokens; you need to manually set your current token in the OA2S_TOKEN environment variable.")
					}
				default:
					return nil, nil, fmt.Errorf("The HTTP auth scheme %s (used in %s) is not currently supported.\n\tContact us to get it added!\n", scheme.Scheme, name)
				}
			case "mutualTLS":
				cert, certExists := auth[name]["cert"]
				key, keyExists := auth[name]["key"]
				if !certExists || !keyExists {
					return nil, nil, fmt.Errorf("Certificate and/or key not configured for %s scheme\n\tNeed `auth.%s.cert` and `auth.%s.key`\n", name, name, name)
				}

				conf.SslUserCert = cert
				conf.SslUserKey = key
			case "oauth2":
				return nil, nil, fmt.Errorf("Unsupported security scheme `oauth2` used in %s\n\tSiege doesn't currently support this authentication mechanism.\n", name)
			case "openIdConnect":
				return nil, nil, fmt.Errorf("Unsupported security scheme `openIdConnect` used in %s\n\tSiege doesn't currently support this authentication mechanism.\n", name)
			default:
				return nil, nil, fmt.Errorf("Unrecognized security scheme `%s` used in %s\n\tOpenAPI v3 doesn't support this authentication type, so we don't know how to proceed\n", scheme.Type, name)
			}
		}
	}

	return urls, conf, nil
}

func getV3RequestNoPayload(c *cli.Context, method, rawPath string, baseUrl *url.URL, urls urlList, methodData *v3.Operation, pathConfig PathConfig) (urlList, error) {
	var err error

	methodConfig, exists := pathConfig[method]
	if !exists {
		return nil, fmt.Errorf("`%s %s` not configured.\n\tNeed `paths.%s.%s.params.{name}`\n", strings.ToUpper(method), rawPath, rawPath, strings.ToLower(method))
	}

	pathBaseUrl := baseUrl

	if len(methodData.Servers) > 0 {
		pathBaseUrl, err = getV3BaseUrl(c, methodData.Servers)
		if err != nil {
			return nil, err
		}
	}

	path, query, cookies, err := getV3PathParams(c, method, rawPath, methodData.Parameters, methodConfig)
	if err != nil {
		return nil, err
	}

	pathUrl := pathBaseUrl.JoinPath(path)

	for _, cookie := range cookies {
		cookie.Path = pathUrl.String()
	}

	pathUrl.RawQuery = query.Encode()

	urls = append(urls, urlData{
		Method:  strings.ToUpper(method),
		URL:     *pathUrl,
		Payload: "",
		Cookies: cookies,
	})

	return urls, nil
}

func getV3RequestWithPayload(c *cli.Context, method, rawPath string, baseUrl *url.URL, urls urlList, methodData *v3.Operation, pathConfig PathConfig) (urlList, error) {
	var err error

	methodConfig, exists := pathConfig[method]
	if !exists {
		return nil, fmt.Errorf("`%s %s` not configured.\n\tNeed `paths.%s.%s.params.{name}` and/or `paths.%s.%s.payloads.{mediaType}`\n", strings.ToUpper(method), rawPath, rawPath, strings.ToLower(method), rawPath, strings.ToLower(method))
	}

	pathBaseUrl := baseUrl

	if len(methodData.Servers) > 0 {
		pathBaseUrl, err = getV3BaseUrl(c, methodData.Servers)
		if err != nil {
			return nil, err
		}
	}

	path, query, cookies, err := getV3PathParams(c, method, rawPath, methodData.Parameters, methodConfig)
	if err != nil {
		return nil, err
	}

	pathUrl := pathBaseUrl.JoinPath(path)

	for _, cookie := range cookies {
		cookie.Path = pathUrl.String()
	}

	pathUrl.RawQuery = query.Encode()

	payloads := make([]string, 0)
	if methodData.RequestBody != nil {
		payloads, err = getV3PathPayloads(c, method, rawPath, *methodData.RequestBody, methodConfig)
		if err != nil {
			return nil, err
		}
	}

	for _, payload := range payloads {
		urls = append(urls, urlData{
			Method:  strings.ToUpper(method),
			URL:     *pathUrl,
			Payload: payload,
			Cookies: cookies,
		})
	}

	return urls, nil
}

func getV3BaseUrl(c *cli.Context, servers []*v3.Server) (*url.URL, error) {
	for _, server := range servers {
		rawUrl := server.URL
		for name, variable := range server.Variables {
			var value string

			config, isType := c.Generic("server.variables").(ServerVarsConfig)
			if isType {
				value = config[name]
			}

			if value == "" {
				value = variable.Default
			}

			if value == "" {
				return nil, fmt.Errorf("Server variable not set and default is empty.\n\tCheck your configuration for `server.variables.%s`\n", name)
			}

			rawUrl = strings.ReplaceAll(rawUrl, fmt.Sprintf("{%s}", name), value)
		}

		baseUrl, err := url.Parse(rawUrl)
		if err != nil {
			return nil, err
		}

		if len(servers) == 1 || c.String("server.description") == server.Description || c.Bool("server.useFirst") {
			return baseUrl, nil
		}
	}

	return nil, fmt.Errorf("Couldn't determine which server to use.\n\tCheck your configuration for `server.description` or `server.useFirst`.\n")
}

func getV3PathParams(c *cli.Context, method, rawPath string, params []*v3.Parameter, config PathMethodConfig) (string, url.Values, []*http.Cookie, error) {
	path := rawPath
	query := make(url.Values)
	cookies := make([]*http.Cookie, 0)

	// log.Printf("Processing path: %s %s - %v - %v\n", method, rawPath, params, config.Params)

	for _, param := range params {
		var paramValue interface{}

		configValue, exists := config.Params[param.Name]
		if !exists && param.Required {
			switch {
			case param.Example != nil:
				paramValue = param.Example
			case len(param.Examples) > 0:
				for _, value := range param.Examples {
					paramValue = value.Value
					break
				}
			case param.AllowEmptyValue:
				paramValue = ""
			default:
				return "", nil, nil, fmt.Errorf("Unconfigured value for %s in %s %s, with no examples to draw from\n\tNeed paths.%s.%s.params.%s\n", param.Name, strings.ToUpper(method), rawPath, rawPath, strings.ToLower(method), param.Name)
			}
		}

		if exists {
			paramValue = configValue
		}

		if exists || param.Required {
			switch param.In {
			case "path":
				path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", param.Name), interfaceToString(paramValue))
			case "query":
				query.Add(param.Name, interfaceToString(paramValue))
			case "header":
				log.Printf("Per-request headers are unsupported by Siege; your tests may not work as expected\n\tSkipping %s for %s\n", param.Name, rawPath)
			case "cookie":
				cookies = append(cookies, &http.Cookie{
					Name:  param.Name,
					Value: interfaceToString(paramValue),
				})
			}
		}
	}

	return path, query, cookies, nil
}

func getV3PathPayloads(c *cli.Context, method, rawPath string, body v3.RequestBody, config PathMethodConfig) ([]string, error) {
	payloads := make([]string, 0)
	var err error

	// log.Printf("Processing path: %s %s - %v - %v\n", method, rawPath, body.Content, config.Payloads)

	for mediatype, details := range body.Content {
		payload, exists := config.Payloads[mediatype]

		if body.Required && !exists {
			addedPayloads := false
			if details.Example != nil {
				payload, err = getPayloadFromType(mediatype, details.Example)
				if err != nil {
					return nil, err
				}

				addedPayloads = true
			}

			for _, example := range details.Examples {
				var examplePayload string

				examplePayload, err = getPayloadFromType(mediatype, example.Value)
				if err != nil {
					return nil, err
				}

				payloads = append(payloads, examplePayload)

				addedPayloads = true
			}

			if !addedPayloads {
				var fakePayload interface{}

				fakePayload, err = createFakePayload(details.Schema)
				if err != nil {
					return nil, err
				}

				if fakePayload != nil {
					payload, err = getPayloadFromType(mediatype, fakePayload)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		if payload != "" {
			payloads = append(payloads, payload)
		}

		if body.Required && len(payloads) < 1 {
			return nil, fmt.Errorf("Unconfigured payload for %s in %s %s, and couldn't generate one\n\tNeed paths.%s.%s.payloads.%s\n", mediatype, strings.ToUpper(method), rawPath, rawPath, method, mediatype)
		}
	}

	if len(payloads) < 1 {
		payloads = append(payloads, "\"\"")
	}

	// log.Printf("inserting payload(s) for: %s %s: %v\n", method, rawPath, payloads)

	return payloads, nil
}

func interfaceToString(in interface{}) string {
	if out, isType := in.(string); isType {
		return out
	}

	if out, isType := in.(fmt.Stringer); isType {
		return out.String()
	}

	return ""
}
