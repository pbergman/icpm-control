package main

import (
	"crypto/ed25519"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

type ScriptConfig struct {
	Vars  map[string]interface{} `toml:"var"`
	User  string
	Group string
	Shell string
	Src   string
	Dst   string
	Code  uint8
}

type Script struct {
	ScriptConfig
	Exec *template.Template
}

type ClientKey struct {
	Id  [md5.Size]byte
	Key ed25519.PublicKey
}

func (c *ClientKey) MarshalText() ([]byte, error) {
	var text = make([]byte, base64.StdEncoding.EncodedLen(len(c.Key)))
	base64.StdEncoding.Encode(text, c.Key)
	return text, nil
}

func (c *ClientKey) UnmarshalText(text []byte) error {
	var key = make([]byte, ed25519.PublicKeySize)
	if _, err := base64.StdEncoding.Decode(key, text); err != nil {
		return err
	}
	c.Key = key
	c.Id = md5.Sum(key)
	return nil
}

type AppConfig struct {
	Debug     bool
	Clients   []*ClientKey
	Interface string
	Block     map[string]string      `toml:"block"`
	Vars      map[string]interface{} `toml:"var"`
}

type Config struct {
	AppConfig
	Scripts []*Script
}

func getConfig(location string) (*Config, error) {

	var stub struct {
		AppConfig
		Scripts []struct {
			ScriptConfig
			Exec toml.Primitive
		} `toml:"script"`
	}

	meta, err := toml.DecodeFile(location, &stub)

	if err != nil {
		return nil, err
	}

	var config = &Config{
		AppConfig: stub.AppConfig,
		Scripts:   make([]*Script, 0),
	}

	if config.Interface == "" {
		return nil, errors.New("missing network interface in config")
	}

	if config.Clients == nil || len(config.Clients) == 0 {
		return nil, errors.New("no client keys defined")
	}

	for i, c := 0, len(stub.Scripts); i < c; i++ {

		var script = &Script{
			ScriptConfig: stub.Scripts[i].ScriptConfig,
		}

		if stub.Scripts[i].Code == 0xff {
			return nil, fmt.Errorf("script %d uses resever code", i)
		}

		config.Scripts = append(config.Scripts, script)

		value, err := getStrings(meta, stub.Scripts[i].Exec)

		if err != nil {
			return nil, err
		}

		tmpl := template.New(fmt.Sprintf("script[%d]", i))
		tmpl.Funcs(getFuncMap())

		for name, block := range config.Block {
			if _, err := tmpl.Parse(fmt.Sprintf(`{{- define "%s" -}}%s{{- end -}}`, name, block)); err != nil {
				return nil, err
			}
		}

		if _, err := tmpl.Parse(strings.Join(value, "\n") + "\n"); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		script.Exec = tmpl
	}

	return config, nil
}

func getFuncMap() template.FuncMap {
	return map[string]any{
		"filepath_join": filepath.Join,
		"filepath_base": filepath.Base,
		"filepath_dir":  filepath.Dir,
		"array": func(args ...interface{}) []interface{} {
			return args
		},
	}
}

func getStrings(meta toml.MetaData, primitive toml.Primitive) ([]string, error) {
	var str string

	if err := meta.PrimitiveDecode(primitive, &str); err == nil {
		return []string{str}, nil
	}

	var strs []string

	if err := meta.PrimitiveDecode(primitive, &strs); err == nil {
		return strs, nil
	} else {
		return nil, err
	}
}
