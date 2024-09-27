package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mchauge/payhip-discord-bot/config"
	"github.com/mchauge/payhip-discord-bot/version"
	log "github.com/s00500/env_logger"
)

//go:generate sh injectGitVars.sh

// Payhip Bot made by:
var maker = "McHauge (mc-hauge@hotmail.com)"

// Bot parameters
var (
	PayhipToken    = flag.String("payhip", "", "Payhip API Token")
	BotToken       = flag.String("token", "", "Bot access token")
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RoleID         = flag.String("role", "", "Role ID to give to verified users")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

	// Version flag
	Version = flag.Bool("version", false, "Print version and exit")
	v       = flag.Bool("v", false, "Print version and exit")
)

var s *discordgo.Session

var (
	// integerOptionMinValue          = 1.0
	// dmPermission                   = false
	// defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{ // Spawnverify command
			Name:        "spawnverify",
			Description: "Spawn a vertify command",
			Type:        discordgo.ChatApplicationCommand,
		},
		{ // Vertify-cli command
			Name:        "verify-cli",
			Description: "Verify license via chat",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "product",
					Description: "Product key",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "license",
					Description: "License key",
					Required:    true,
				},
			},
		},
		{ // Ping command
			Name:        "ping",
			Description: "ping pong",
			Type:        discordgo.ChatApplicationCommand,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"spawnverify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			button := discordgo.Button{
				Label:    "Verify",
				Style:    discordgo.PrimaryButton,
				Disabled: false,
				CustomID: "verify_button",
			}

			embed := discordgo.MessageEmbed{
				Title:       "Verify your purchase",
				Description: "Click the button below to begin verifying your purchase",
				Color:       0x2fdf0c,
			}

			// Spawn the message
			_, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
				Embed: &embed,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							button,
						},
					},
				},
			})

			// Send the response
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Verification message created",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			log.Should(err)
		},
		"verify-cli": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			product := i.ApplicationCommandData().Options[0].StringValue()
			license := i.ApplicationCommandData().Options[1].StringValue()

			// Verify the license
			verified, err := VerifyLicense(product, license, *PayhipToken, config.Config.MaxLicenseUses)
			if err != nil {
				log.Errorf("Verifying license: %v", err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error verifying license",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			} else {
				// Send the response
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "License vertification: " + verified,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				Username := i.Member.User.Username
				if i.Member.User.Discriminator != "0" {
					Username = i.Member.User.Username + "#" + i.Member.User.Discriminator
				}

				log.Info("Vertification of license: " + verified + " for user: " + Username)
				if verified == "Success" {
					log.Info("Gave User: " + Username + " the Verified role")
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, *RoleID)
				}
			}
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		},
	}

	componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"verify_button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "vertify_modal",
					Title:    "Verify Your License",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "product-key",
									Label:       "Product key",
									Style:       discordgo.TextInputShort,
									Placeholder: "XXXXX",
									Required:    true,
									MinLength:   5,
									MaxLength:   5,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "license-key",
									Label:       "License Key",
									Style:       discordgo.TextInputShort,
									Placeholder: "XXXXX-XXXXX-XXXXX-XXXXX",
									Required:    true,
									MinLength:   23,
									MaxLength:   23,
								},
							},
						},
					},
				},
			})
			log.Should(err)
		},
	}

	modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"vertify_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ModalSubmitData()
			product := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			license := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			// Verify the license
			verified, err := VerifyLicense(product, license, *PayhipToken, config.Config.MaxLicenseUses)
			if err != nil {
				log.Errorf("Error verifying license: %v", err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error verifying license",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			} else {

				Username := i.Member.User.Username
				if i.Member.User.Discriminator != "0" {
					Username = i.Member.User.Username + "#" + i.Member.User.Discriminator
				}

				if verified == "Success" {
					log.Info("Vertification of license: " + verified + " for user: " + Username)
					err := s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, *RoleID)
					if !log.Should(err) {
						log.Info("Gave User: " + Username + " the Verified role")
						verified = "Success"
					} else {
						log.Error("Failed to give User: " + Username + " the Verified role")
						verified = "Valid license but failed to give the Verified role"
					}

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "License Vertification: " + verified,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})

				} else {
					log.Info("Failed to Vertify license: " + verified + " for user: " + Username)

					// Send the response
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "License Vertification: " + verified,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				}
			}
		},
	}
)

func init() {
	branch := ""
	if gitBranch != "master" && gitBranch != "main" {
		branch = "- branch: " + gitBranch + " "
	}

	flag.Parse()
	if *Version || *v {
		// Version contains version and Git commit information.
		//
		// The placeholders are replaced on `git archive` using the `export-subst` attribute.
		var intVersion = version.Version(fmt.Sprintf("%s (%s) %s", gitTag, gitRevision, branch), "$Format:%(describe)$", "$Format:%H$")

		intVersion.Print(maker)
		os.Exit(0)
	}

	log.Infof("Payhip Discord bot by %s, version %s (%s) %s", maker, gitTag, gitRevision, branch)

	if *BotToken == "" {
		log.Warn("No bot token provided, using config file instead")
		config.ReadConfig()
		config.ReadEnvConfig() // override config file with .env file if it has values

		if config.ConfigIsValid() {
			*PayhipToken = config.Config.PayhipToken
			*BotToken = config.Config.BotToken
			*RoleID = config.Config.RoleID
			*PayhipToken = config.Config.PayhipToken
			*GuildID = config.Config.GuildID
			*RemoveCommands = config.Config.RemoveCommands
		} else {
			log.Fatalf("Invalid bot parameters, please check your config or env file")
		}
	}

	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.GuildID != *GuildID {
				log.Debugf("Ignoring command from different guild, expected: %v, got: %v", *GuildID, i.GuildID)
				return
			}

			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if i.GuildID != *GuildID {
				log.Debugf("Ignoring command from different guild, expected: %v, got: %v", *GuildID, i.GuildID)
				return
			}

			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if i.GuildID != *GuildID {
				log.Debugf("Ignoring command from different guild, expected: %v, got: %v", *GuildID, i.GuildID)
				return
			}

			if h, ok := modalHandlers[i.ModalSubmitData().CustomID]; ok {
				h(s, i)
			}
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Infoln("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Infoln("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Infoln("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Infoln("Gracefully shutting down.")
}

func VerifyLicense(product string, license string, PayhipToken string, MaxLicenseUses int) (string, error) {
	// Create socket
	netClient := &http.Client{
		Timeout: time.Second * 5,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://payhip.com/api/v1/license/verify?product_link=%s&license_key=%s", product, license), nil)
	req.Header.Add("payhip-api-key", PayhipToken)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Should(err)
		return "Payhip API Error", err
	}

	data := message{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal([]byte(body), &data)
	defer resp.Body.Close()

	log.Debugln(log.Indent(data))

	// Needs to be a valid key and have a buyer email
	if data.Data.Enabled && data.Data.Buyer_email != "" {

		// Check if the license key has been used more than the max allowed uses
		if MaxLicenseUses != 0 && data.Data.Uses > MaxLicenseUses-1 {
			return "Max uses reached", nil
		}

		// Add usage to the license key if it's not unlimited
		if MaxLicenseUses != 0 {
			msg, err := data.AddUsage(product, license, PayhipToken)
			if err != nil {
				return "Failed to add usage", err
			}

			if msg != "" {
				return msg, nil
			}
		}

		return "Success", nil
	}
	return "Failed", nil
}

func (m message) AddUsage(product string, license string, PayhipToken string) (string, error) {

	// Create socket
	netClient := &http.Client{
		Timeout: time.Second * 5,
	}

	req, _ := http.NewRequest("PUT", "https://payhip.com/api/v1/license/usage", strings.NewReader(fmt.Sprintf("product_link=%s&license_key=%s", product, license)))
	req.Header.Add("payhip-api-key", PayhipToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := netClient.Do(req)
	if err != nil {
		log.Should(err)
		return "Add Usage: Payhip API Error", err
	}

	data := message{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal([]byte(body), &data)
	defer resp.Body.Close()

	log.Debugln("Add Usage: ", log.Indent(data))

	if data.Data.Product_link == "" {
		return fmt.Sprintf("Failed to add usage: %s", string(body)), nil
	}

	return "", nil
}

type message struct {
	Data data
}

type data struct {
	Enabled      bool
	Product_link string
	License_key  string
	Buyer_email  string
	Uses         int
	Date         string
}
