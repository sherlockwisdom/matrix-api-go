package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
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
	KeystoreFilepath string                    `yaml:"keystore_filepath"`
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

func (c *Conf) GetBridges() []*Bridges {
	var bridges []*Bridges
	for _, entry := range c.Bridges {
		for name, _ := range entry {
			bridges = append(bridges, &Bridges{Name: name})
		}
	}
	return bridges
}

func ParseImage(client *mautrix.Client, url string) ([]byte, error) {
	fmt.Printf(">>\tParsing image for: %v\n", url)
	contentUrl, err := id.ParseContentURI(url)
	if err != nil {
		panic(err)
	}
	return client.DownloadBytes(context.Background(), contentUrl)
}

func (c *Conf) CheckSuccessPattern(bridgeType string, input string) (bool, error) {
	config, ok := c.GetBridgeConfig(bridgeType)
	if !ok {
		return false, fmt.Errorf("bridge type %s not found in configuration", bridgeType)
	}

	successPattern, ok := config.Cmd["success"]
	if !ok {
		return false, fmt.Errorf("success pattern not found for bridge type %s", bridgeType)
	}

	// Replace %s with .* to create a regex pattern
	regexPattern := strings.ReplaceAll(successPattern, "%s", ".*")
	matched, err := regexp.MatchString(regexPattern, input)
	if err != nil {
		return false, fmt.Errorf("error matching pattern: %v", err)
	}

	return matched, nil
}
