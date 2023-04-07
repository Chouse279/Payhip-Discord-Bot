# Payhip-Discord-Bot
Open scource Discord bot for Payhip, providig a way to automate sale vertifications, very simple in it's current form

To run the bot:
 - Install Golang: https://go.dev/dl/
 - Go to the folder with the main.go file and open a terminal
 - run ```go install```
 - run ```go mod tidy```

Then run the program:

```go run main.go -payhip YOUR_PAYHIP_APIKEY -token YOUR_DISCORD_BOT_TOKEN -guild DISCORD_SERVER_ID -role DISCORD_SERVER_ROLE_ID```

When it's running you will see something like this in the console:

```
2023/04/06 11:38:50 Adding commands...
2023/04/06 11:38:50 Logged in as: BOT_NAME
2023/04/06 11:38:50 Press Ctrl+C to exit
```

Once you see that in the console, your ready to go to your server and add the vertification button.
Create a text channel and run the action ```/spawnverify``` that will give you a button with a UI and all for handeling it.
If you wanna do it via a chat message instead you can use the: ```/verify-cli``` command instead

To close the bot simply press Ctrl+C then it will close itself again
