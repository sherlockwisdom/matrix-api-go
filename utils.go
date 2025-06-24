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
	BotName          string            `yaml:"botname"`
	UsernameTemplate string            `yaml:"username_template"`
	Cmd              map[string]string `yaml:"cmd"` // ← map instead of slice of maps
}

type Tls struct {
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}

type ServerWebsocket struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Tls  Tls    `yaml:"tls"`
}

type Server struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Tls  Tls    `yaml:"tls"`
}

type User struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	AccessToken string `yaml:"access_token"`
}

type Conf struct {
	Server           Server                    `yaml:"server"`
	Websocket        ServerWebsocket           `yaml:"websocket"`
	KeystoreFilepath string                    `yaml:"keystore_filepath"`
	HomeServer       string                    `yaml:"homeserver"`
	HomeServerDomain string                    `yaml:"homeserver_domain"`
	Bridges          []map[string]BridgeConfig `yaml:"bridges"`
	User             User                      `yaml:"user"`
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
		return nil, err
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

func (c *Conf) CheckOngoingPattern(bridgeType string, input string) (bool, error) {
	config, ok := c.GetBridgeConfig(bridgeType)
	if !ok {
		return false, fmt.Errorf("bridge type %s not found in configuration", bridgeType)
	}

	ongoingPattern, ok := config.Cmd["ongoing"]
	if !ok {
		return false, fmt.Errorf("ongoing pattern not found for bridge type %s", bridgeType)
	}

	matched, err := regexp.MatchString(ongoingPattern, input)
	if err != nil {
		return false, fmt.Errorf("error matching pattern: %v", err)
	}

	return matched, nil
}

func (c *Conf) CheckUsernameTemplate(bridgeType string, username string) (bool, error) {
	config, ok := c.GetBridgeConfig(bridgeType)
	if !ok {
		return false, fmt.Errorf("bridge type %s not found in configuration", bridgeType)
	}

	if config.UsernameTemplate == "" {
		return false, fmt.Errorf("username template not found for bridge type %s", bridgeType)
	}

	// Convert template pattern to regex pattern
	// Replace {{.}} with .* to match any characters
	regexPattern := strings.ReplaceAll(config.UsernameTemplate, "{{.}}", ".*")
	// Escape any other special regex characters
	regexPattern = regexp.QuoteMeta(regexPattern)
	// Restore the .* pattern
	regexPattern = strings.ReplaceAll(regexPattern, "\\.\\*", ".*")

	matched, err := regexp.MatchString(regexPattern, username)
	if err != nil {
		return false, fmt.Errorf("error matching username pattern: %v", err)
	}

	return matched, nil
}

func (c *Conf) FormatUsername(bridgeType string, username string) (string, error) {
	config, ok := c.GetBridgeConfig(bridgeType)
	if !ok {
		return "", fmt.Errorf("bridge type %s not found in configuration", bridgeType)
	}

	if config.UsernameTemplate == "" {
		return "", fmt.Errorf("username template not found for bridge type %s", bridgeType)
	}

	// Replace {{.}} with the actual username
	formattedUsername := strings.ReplaceAll(config.UsernameTemplate, "{{.}}", username)

	// Ensure the username is properly formatted as a Matrix user ID
	// Matrix user IDs should be in the format @localpart:domain
	if !strings.HasPrefix(formattedUsername, "@") {
		formattedUsername = "@" + formattedUsername
	}
	if !strings.Contains(formattedUsername, ":") {
		formattedUsername = formattedUsername + ":" + c.HomeServerDomain
	}

	return formattedUsername, nil
}

// ExtractBracketContent extracts the content inside the first pair of parentheses in the input string.
func ExtractBracketContent(input string) (string, error) {
	start := strings.Index(input, "(")
	end := strings.Index(input, ")")
	if start == -1 || end == -1 || end <= start+1 {
		return "", fmt.Errorf("no content found in brackets")
	}
	content := input[start+1 : end]
	// Remove the "+" character from the content
	content = strings.ReplaceAll(content, "+", "")
	return content, nil
}

func ReverseAliasForEventSubscriber(username, bridgeName, homeserver string) string {
	// @username:bridgeName:homeserver.com -> username_bridgeName
	return fmt.Sprintf("@%s:%s:%s", username, bridgeName, homeserver)
}
