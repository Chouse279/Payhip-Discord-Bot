package config

import (
	"os"

	"github.com/mchauge/payhip-discord-bot/helpers"
	log "github.com/s00500/env_logger"
)

// Stores config data
var Config *configStruct

type configStruct struct {
	PayhipToken    string `json:"PayhipToken"`
	BotToken       string `json:"BotToken"`
	GuildID        string `json:"GuildID"`
	RoleID         string `json:"RoleID"`
	RemoveCommands bool   `json:"RemoveCommands"`
}

func ReadConfig() {
	path, err := os.Getwd()
	log.Should(err)
	helpers.ReadJson(&Config, path, "config.json")

	if Config == nil || Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		Config = &configStruct{
			PayhipToken:    "",
			BotToken:       "",
			GuildID:        "",
			RoleID:         "",
			RemoveCommands: true,
		}
		helpers.UpdateJson(&Config, path, "config.json")
		log.Fatal("Config is empty or missing info please fill your info in the config.json file")
	}

	log.Warn("Config file loaded")
}
