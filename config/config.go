package config

import (
	"os"
	"strconv"

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
	MaxLicenseUses int    `json:"MaxLicenseUses"`
}

func ReadConfig() {
	path, err := os.Getwd()
	log.ShouldWarn(err)
	helpers.ReadJson(&Config, path, "config.json")

	if Config == nil || Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		Config = &configStruct{
			PayhipToken:    "caf89e1951d4cd6eecb2f14bbd7ded1fd0f60546",
			BotToken:       "MTM0NTYzOTQ2MTUzNDQzNzM4Ng.GBo0f8.Br0_6QniBevDNYDrWHTX1BZPX3fqco7DSDxFMc",
			GuildID:        "1255171192965566494",
			RoleID:         "1255171193003315284",
			RemoveCommands: false,
			MaxLicenseUses: 0, // 0 means unlimited, and will not add uses to the license key
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
		log.Warn("Error loading .env file, skipping... Checking for environment variables")
	}

	maxlicense, err := strconv.Atoi(os.Getenv("MAX_LICENSE_USES"))
	if err != nil {
		maxlicense = 0
	}

	// Get the environment variables
	LocalConfig := &configStruct{
		PayhipToken:    os.Getenv("PAYHIP_TOKEN"),
		BotToken:       os.Getenv("BOT_TOKEN"),
		GuildID:        os.Getenv("GUILD_ID"),
		RoleID:         os.Getenv("ROLE_ID"),
		RemoveCommands: os.Getenv("REMOVE_COMMANDS") == "false", // Only set to true if it is true
		MaxLicenseUses: int(maxlicense),
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
	if LocalConfig.MaxLicenseUses != 0 { // Only set if it is not 0
		Config.MaxLicenseUses = LocalConfig.MaxLicenseUses
	}

	// Check if all values are filled
	if Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		log.Error("Config is empty or missing info please fill your info in the .env file")
		return
	}

	log.Warn("Env file loaded")
}

// Check for missing config values
func ConfigIsValid() bool {
	if Config.PayhipToken == "" || Config.BotToken == "" || Config.GuildID == "" || Config.RoleID == "" {
		log.Error("Config is empty or missing info please fill your info in the config.json file or the .env file")
		return false
	}
	return true
}
