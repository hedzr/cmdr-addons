// Copyright © 2020 Hedzr Yeh.

package cfg

import (
	"github.com/hedzr/cmdr-addons/pkg/vxconf"
	"github.com/hedzr/log/dir"
	"gopkg.in/yaml.v3"
)

// YAML -----------------------------------------------------------------------

// ParseYaml reads a YAML configuration from the given string.
func ParseYaml(cfg string) (*AppConfig, error) {
	return parseYaml([]byte(cfg))
}

// ParseYamlFile reads a YAML configuration from the given filename.
func ParseYamlFile(filename string) (*AppConfig, error) {
	cfg, err := dir.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseYaml(cfg)
}

// parseYaml performs the real YAML parsing.
func parseYaml(cfgbin []byte) (*AppConfig, error) {
	var out interface{}
	var err error
	if err = yaml.Unmarshal(cfgbin, &out); err != nil {
		return nil, err
	}
	if out, err = NormalizeValue(out); err != nil {
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
	return vxconf.UnescapeUnicode(b), nil
}
