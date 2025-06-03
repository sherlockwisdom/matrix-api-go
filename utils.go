package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BridgeConfig struct {
	BotName string            `yaml:"botname"`
	Cmd     map[string]string `yaml:"cmd"` // ← map instead of slice of maps
}

type Tls struct {
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}

type Server struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Tls  Tls    `yaml:"tls"`
}

type Conf struct {
	Server           Server                    `yaml:"server"`
	HomeServer       string                    `yaml:"homeserver"`
	HomeServerDomain string                    `yaml:"homeserver_domain"`
	Bridges          []map[string]BridgeConfig `yaml:"bridges"`
}

func (c *Conf) getConf() (*Conf, error) {
	yamlFile, err := os.ReadFile("conf.yaml")
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (c *Conf) GetBridgeConfig(bridgeType string) (*BridgeConfig, bool) {
	for _, entry := range c.Bridges {
		if config, ok := entry[bridgeType]; ok {
			return &config, true
		}
	}
	return nil, false
}
