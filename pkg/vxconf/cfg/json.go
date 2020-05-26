// Copyright © 2020 Hedzr Yeh.

package cfg

import (
	"encoding/json"
	"io/ioutil"
)

// JSON -----------------------------------------------------------------------

// ParseJSON reads a JSON configuration from the given string.
func ParseJSON(cfg string) (*AppConfig, error) {
	return parseJSON([]byte(cfg))
}

// ParseJSONFile reads a JSON configuration from the given filename.
func ParseJSONFile(filename string) (*AppConfig, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseJSON(cfg)
}

// parseJSON performs the real JSON parsing.
func parseJSON(cfgbin []byte) (*AppConfig, error) {
	var out interface{}
	var err error
	if err = json.Unmarshal(cfgbin, &out); err != nil {
		return nil, err
	}
	if out, err = NormalizeValue(out); err != nil {
		return nil, err
	}
	return &AppConfig{Root: out}, nil
}

// RenderJSON renders a JSON configuration.
func RenderJSON(cfg interface{}) (string, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
