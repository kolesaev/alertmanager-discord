package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

// DiscordChannel contains the necessary Discord DiscordChannel properties
// for the application
type DiscordChannel struct {
	Name                        string   `json:"name" yaml:"name"`
	WebhookURL                  string   `json:"webhookURL" yaml:"webhookURL"`
	RolesToMention              []string `json:"rolesToMention" yaml:"rolesToMention"`
	SeveritiesToMention         []string `json:"severitiesToMention" yaml:"severitiesToMention"`
	SeveritiesToIgnoreWhenAlone []string `json:"severitiesToIgnoreWhenAlone" yaml:"severitiesToIgnoreWhenAlone"`
}

// StatusAppearance defines the Embed's color and Emoji to be used in the title
type StatusAppearance struct {
	Color int    `json:"color" yaml:"color"`
	Emoji string `json:"emoji" yaml:"emoji"`
}

// SeverityAppearance defines the Embed's color and Emoji to be used in the title
type SeverityAppearance struct {
	Color    int    `json:"color" yaml:"color"`
	Emoji    string `json:"emoji" yaml:"emoji"`
	Priority int    `json:"priority" yaml:"priority"`
}

type SeverityDefinition struct {
	Label  string                        `json:"label" yaml:"label"`
	Values map[string]SeverityAppearance `json:"values" yaml:"values"`
}

// DashboardLinkConfig defines configuration for dashboard links
type DashboardLinkConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Label   string `json:"label" yaml:"label"`
	Text    string `json:"text" yaml:"text"`
	// Position for dashboard link: "content", "embed_top", or "embed_bottom"
	Position string `json:"position" yaml:"position"`
}

// GeneratorLinkConfig defines configuration for generator links
type GeneratorLinkConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Text    string `json:"text" yaml:"text"`
	// Position for generator link: "content", "embed_top", or "embed_bottom"
	Position string `json:"position" yaml:"position"`
}

// TimeDisplayConfig defines configuration for time display
type TimeDisplayConfig struct {
	Enabled             bool     `json:"enabled" yaml:"enabled"`
	StartsAtText        string   `json:"startsAtText" yaml:"startsAtText"`
	EndsAtText          string   `json:"endsAtText" yaml:"endsAtText"`
	DurationText        string   `json:"durationText" yaml:"durationText"`
	HiddenForSeverities []string `json:"hiddenForSeverities" yaml:"hiddenForSeverities"`
}

// Config defines the (.yaml|.json) config structured to be used by the app
type Config struct {
	AvatarURL                   string                      `json:"avatarURL" yaml:"avatarURL"`
	ListenAddress               string                      `json:"listenAddress,omitempty" yaml:"listenAddress,omitempty"`
	Username                    string                      `json:"username" yaml:"username"`
	MessageType                 string                      `json:"messageType" yaml:"messageType"`
	Status                      map[string]StatusAppearance `json:"status" yaml:"status"`
	FiringCountToMention        int                         `json:"firingCountToMention" yaml:"firingCountToMention"`
	RolesToMention              []string                    `json:"rolesToMention" yaml:"rolesToMention"`
	SeveritiesToMention         []string                    `json:"severitiesToMention" yaml:"severitiesToMention"`
	SeveritiesToIgnoreWhenAlone []string                    `json:"severitiesToIgnoreWhenAlone" yaml:"severitiesToIgnoreWhenAlone"`
	Severity                    SeverityDefinition          `json:"severity" yaml:"severity"`
	DashboardLink               DashboardLinkConfig         `json:"dashboardLink" yaml:"dashboardLink"`
	GeneratorLink               GeneratorLinkConfig         `json:"generatorLink" yaml:"generatorLink"`
	TimeDisplay                 TimeDisplayConfig           `json:"timeDisplay" yaml:"timeDisplay"`
	DiscordChannels             map[string]DiscordChannel   `json:"channels" yaml:"channels"`
}

var defaultConfig = Config{
	ListenAddress:        ":8080",
	MessageType:          "status",
	AvatarURL:            "https://raw.githubusercontent.com/kolesaev/alertmanager-discord/master/assets/images/prometheus-logo.png",
	Username:             "alertmanager",
	FiringCountToMention: -1,
	Status: map[string]StatusAppearance{
		"firing": {
			Emoji: ":rotating_light:",
			Color: 10038562, // EmbedColorDarkRed
		},
		"resolved": {
			Emoji: ":white_check_mark:",
			Color: 3066993, // EmbedColorGreen
		},
	},
	Severity: SeverityDefinition{
		Label: "severity",
		Values: map[string]SeverityAppearance{
			"unknown": {
				Color: 9807270, // EmbedColorGrey
				Emoji: ":grey_question:",
			},
			"information": {
				Color: 3447003, // EmbedColorBlue
				Emoji: ":information_source:",
			},
			"info": {
				Color: 3447003, // EmbedColorBlue
				Emoji: ":information_source:",
			},
			"warning": {
				Color:    15844367, // EmbedColorGold
				Emoji:    ":warning:",
				Priority: 1,
			},
			"warn": {
				Color:    15844367, // EmbedColorGold
				Emoji:    ":warning:",
				Priority: 1,
			},
			"critical": {
				Color:    11027200, // EmbedColorDarkOrange
				Emoji:    ":rotating_light:",
				Priority: 2,
			},
			"disaster": {
				Color:    10038562, // EmbedColorDarkRed
				Emoji:    ":fire:",
				Priority: 3,
			},
		},
	},
	DashboardLink: DashboardLinkConfig{
		Enabled:  false,
		Label:    "url",
		Text:     "Open in Dashboard",
		Position: "content",
	},
	GeneratorLink: GeneratorLinkConfig{
		Enabled:  false,
		Text:     "Open in PromQL",
		Position: "content",
	},
	TimeDisplay: TimeDisplayConfig{
		Enabled:             false,
		StartsAtText:        "Started at:",
		EndsAtText:          "Ended at:",
		DurationText:        "Duration:",
		HiddenForSeverities: []string{},
	},
}

// LoadUserConfig provides a Config struct to be used throughout the application
func LoadUserConfig() *Config {
	configFilePath := getEnv("CONFIG_PATH", "./config.yaml")

	userConfig := loadConfigurationFile(configFilePath)

	config := defaultConfig

	err := mergo.Merge(&config, userConfig, mergo.WithOverride)
	if err != nil {
		log.Fatalln(err)
	}

	yamlConfig, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using the following config:\n\n=======\n\n%s\n\n========\n\n", string(yamlConfig))

	return &config
}

func loadConfigurationFile(file string) Config {
	var config Config

	configFile, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}

	defer configFile.Close()

	if strings.HasSuffix(file, ".json") {
		jsonParser := json.NewDecoder(configFile)
		err := jsonParser.Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		yamlParser := yaml.NewDecoder(configFile)
		err := yamlParser.Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	}

	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
