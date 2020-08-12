// Copyright © 2019 Hedzr Yeh.

package vxconf

import (
	"encoding/json"
	"github.com/hedzr/cmdr"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func JsonToString(in interface{}, pretty bool) string {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(in, "", "  ")
	} else {
		b, err = json.Marshal(in)
	}
	if err != nil {
		cmdr.Logger.Errorf("json decode string failed: %v", err)
		return "<error-json-object>"
	} else {
		return string(b)
	}
}

func YamlToString(in interface{}, pretty bool) string {
	var b []byte
	var err error
	if pretty {
		b, err = yaml.Marshal(in)
	} else {
		b, err = yaml.Marshal(in)
	}
	if err != nil {
		cmdr.Logger.Errorf("yaml decode string failed: %v", err)
		return "<error-yaml-object>"
	} else {
		return string(b)
	}
}

// JSON -----------------------------------------------------------------------

// ParseJson reads a JSON configuration from the given string.
func ParseJson(cfg string) (*AppConfig, error) {
	return parseJson([]byte(cfg))
}

// ParseJsonFile reads a JSON configuration from the given filename.
func ParseJsonFile(filename string) (*AppConfig, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseJson(cfg)
}

// parseJson performs the real JSON parsing.
func parseJson(cfg []byte) (*AppConfig, error) {
	var out interface{}
	var err error
	if err = json.Unmarshal(cfg, &out); err != nil {
		return nil, err
	}
	if out, err = normalizeValue(out); err != nil {
		return nil, err
	}
	return &AppConfig{Root: out}, nil
}

// RenderJson renders a JSON configuration.
func RenderJson(cfg interface{}) (string, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// YAML -----------------------------------------------------------------------

// ParseYaml reads a YAML configuration from the given string.
func ParseYaml(cfg string) (*AppConfig, error) {
	return parseYaml([]byte(cfg))
}

// ParseYamlFile reads a YAML configuration from the given filename.
func ParseYamlFile(filename string) (*AppConfig, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseYaml(cfg)
}

// parseYaml performs the real YAML parsing.
func parseYaml(cfg []byte) (*AppConfig, error) {
	var out interface{}
	var err error
	if err = yaml.Unmarshal(cfg, &out); err != nil {
		return nil, err
	}
	if out, err = normalizeValue(out); err != nil {
		return nil, err
	}
	return &AppConfig{Root: out}, nil
}

// RenderYaml renders a YAML configuration.
func RenderYaml(cfg interface{}) (string, error) {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return UnescapeUnicode(b), nil
}
