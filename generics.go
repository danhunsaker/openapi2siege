package main

import "encoding/json"

type ServerVarsConfig map[string]string

type AuthConfig map[string]map[string]string

type PathsConfig map[string]PathConfig

type PathConfig map[string]PathMethodConfig

type PathMethodConfig struct {
	Params   map[string]string `json:"params"`
	Payloads map[string]string `json:"payloads"`
}

func (c ServerVarsConfig) Set(value string) error {
	return json.Unmarshal([]byte(value), &c)
}

func (c ServerVarsConfig) String() string {
	value, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(value)
}

func (c ServerVarsConfig) FromJson(raw []byte) error {
	return json.Unmarshal(raw, &c)
}

func (c AuthConfig) Set(value string) error {
	return json.Unmarshal([]byte(value), &c)
}

func (c AuthConfig) String() string {
	value, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(value)
}

func (c AuthConfig) FromJson(raw []byte) error {
	return json.Unmarshal(raw, &c)
}

func (c PathsConfig) Set(value string) error {
	return json.Unmarshal([]byte(value), &c)
}

func (c PathsConfig) String() string {
	value, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(value)
}

func (c PathsConfig) FromJson(raw []byte) error {
	return json.Unmarshal(raw, &c)
}
