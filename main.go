package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	PayhipToken    = flag.String("payhip", "", "Payhip API Token")
	BotToken       = flag.String("token", "", "Bot access token")
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RoleID         = flag.String("role", "", "Role ID to give to verified users")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
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

			if err != nil {
				log.Print(err)
				return
			}
		},
		"verify-cli": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			product := i.ApplicationCommandData().Options[0].StringValue()
			license := i.ApplicationCommandData().Options[1].StringValue()

			// Verify the license
			verified, err := VerifyLicense(product, license, *PayhipToken)
			if err != nil {
				log.Printf("Error verifying license: %v", err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error verifying license",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			// Send the response
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "License vertification: " + verified,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

			log.Print("Vertification of license: " + verified + " for user: " + i.Member.User.Username + "#" + i.Member.User.Discriminator)
			if verified == "Success" {
				log.Print("Gave User: " + i.Member.User.Username + "#" + i.Member.User.Discriminator + " the Verified role")
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, *RoleID)
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
			if err != nil {
				log.Print(err)
				return
			}
		},
	}

	modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"vertify_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ModalSubmitData()
			product := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			license := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			// Verify the license
			verified, err := VerifyLicense(product, license, *PayhipToken)
			if err != nil {
				log.Printf("Error verifying license: %v", err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error verifying license",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			// Send the response
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "License Vertification: " + verified,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

			log.Print("Vertification of license: " + verified + " for user: " + i.Member.User.Username + "#" + i.Member.User.Discriminator)
			if verified == "Success" {
				log.Print("Gave User: " + i.Member.User.Username + "#" + i.Member.User.Discriminator + " the Verified role")
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, *RoleID)
			}
		},
	}
)

func init() {
	flag.Parse()

	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if h, ok := modalHandlers[i.ModalSubmitData().CustomID]; ok {
				h(s, i)
			}
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
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
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}

func VerifyLicense(product string, license string, PayhipToken string) (string, error) {
	// Create socket
	netClient := &http.Client{
		Timeout: time.Second * 5,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://payhip.com/api/v1/license/verify?product_link=%s&license_key=%s", product, license), nil)
	req.Header.Add("payhip-api-key", PayhipToken)

	resp, err := netClient.Do(req)
	if err != nil {
		return "", err
	}

	data := message{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(body), &data)
	defer resp.Body.Close()

	if data.Data.Buyer_email != "" {
		return "Success", nil
	}
	return "Failed", nil
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
