# Payhip-Discord-Bot
Open source Discord bot for Payhip, providing a way to automate sale verifications, very simple in its current form

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
