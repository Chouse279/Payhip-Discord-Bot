# Payhip-Discord-Bot
Open source Discord bot for Payhip, providing a way to automate sale verifications, very simple in its current form

***This app is in no way associated with [Payhip](https://payhip.com/), the company itself, and only relies on the public license API provided***

To run the bot:
 - Install Golang: https://go.dev/dl/
 - Go to the folder with the main.go file and open a terminal
 - run ```go install```
 - run ```go mod tidy```

Then run the program from the source:

```go run main.go -payhip YOUR_PAYHIP_APIKEY -token YOUR_DISCORD_BOT_TOKEN -guild DISCORD_SERVER_ID -role DISCORD_SERVER_ROLE_ID```

If you download a release pre-compiled release version for your system use this, just change the file in front to match, your downloaded file:

```./payhip-discord-bot-windows-amd64.exe -payhip YOUR_PAYHIP_APIKEY -token YOUR_DISCORD_BOT_TOKEN -guild DISCORD_SERVER_ID -role DISCORD_SERVER_ROLE_ID```

If you don't specify anything when running the program, it will generate a config.json file for you that you can fill out with the relevant info, and then that will be used instead.

If you would rather use a .env file, then sure, go ahead anf make one, then the program will use that instead og the config.json file an empty .env should look like this:

```
PAYHIP_TOKEN=
BOT_TOKEN=
GUILD_ID=
ROLE_ID=
REMOVE_COMMANDS=
```

When it's running you will see something like this in the console:

```
2023/04/06 11:38:50 Adding commands...
2023/04/06 11:38:50 Logged in as: BOT_NAME
2023/04/06 11:38:50 Press Ctrl+C to exit
```

Once you see that in the console, you are ready to go to your server and add the verification button.
Create a text channel and run the action ```/spawnverify``` that will give you a button with a UI and all for handling it.
If you wanna do it via a chat message instead you can use the: ```/verify-cli``` command instead

To close the bot simply press Ctrl+C then it will close itself again

Usefull links: 
- [Discord Developer](https://discord.com/developers/applications)
- [Discord Bot Setup Guide](https://discordpy.readthedocs.io/en/stable/discord.html)
- [Payhip License Key Setup & API](https://help.payhip.com/article/114-software-license-keys)

Initial idea and insperation for this bot comes from the intergration of the Gumroad bot: [GumCord](https://github.com/benaclejames/GumCord)
