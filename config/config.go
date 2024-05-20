package config

import (
	"os"

	"github.com/joho/godotenv"
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
		log.Error("Config is empty or missing info please fill your info in the config.json file")
		return
	}

	log.Warn("Config file loaded")
}

// Use godotenv to read .env file
func ReadEnvConfig() {
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")

	if err != nil {
		log.Error("Error loading .env file, skipping...")
		return
	}

	// Get the environment variables
	LocalConfig := &configStruct{
		PayhipToken:    os.Getenv("PAYHIP_TOKEN"),
		BotToken:       os.Getenv("BOT_TOKEN"),
		GuildID:        os.Getenv("GUILD_ID"),
		RoleID:         os.Getenv("ROLE_ID"),
		RemoveCommands: os.Getenv("REMOVE_COMMANDS") == "true",
	}

	// store non empty values
	if Config.PayhipToken == "" {
		Config.PayhipToken = LocalConfig.PayhipToken
	}
	if Config.BotToken == "" {
		Config.BotToken = LocalConfig.BotToken
	}
	if Config.GuildID == "" {
		Config.GuildID = LocalConfig.GuildID
	}
	if Config.RoleID == "" {
		Config.RoleID = LocalConfig.RoleID
	}
	if LocalConfig.RemoveCommands { // Only set to true if it is true
		Config.RemoveCommands = LocalConfig.RemoveCommands
	}

	// Check if all values are filled
	if Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		log.Error("Config is empty or missing info please fill your info in the .env file")
		return
	}

	log.Warn("Env file loaded")
}

// Check for missing config values
func ConfigIsValid() bool{
	if Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		log.Error("Config is empty or missing info please fill your info in the config.json file or the .env file")
		return false
	}
	return true
}