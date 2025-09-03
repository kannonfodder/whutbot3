Whutbot3 — Discord bot skeleton

This repository contains a minimal Go Discord bot that connects to a single channel and listens for messages beginning with `https://stashdb.org/scenes`.

Configuration

- `DISCORD_TOKEN` — your bot token
- `TARGET_CHANNEL_ID` — the channel ID the bot should monitor

Quick start (PowerShell):

1. Set environment variables for this session: `$env:DISCORD_TOKEN = "<token>"; $env:TARGET_CHANNEL_ID = "<channel id>"`
2. Run the bot: `go run main.go`

Notes

- The bot uses `github.com/bwmarrin/discordgo`. Run `go mod tidy` to fetch dependencies.
- You'll need to enable the Message Content intent for your bot in the Discord Developer Portal if you want to read message content.
- Replace the placeholder processing in `main.go` with your own logic.

## Docker

Build and run using Docker (PowerShell):

1. Copy `.env.example` to `.env` and fill in `DISCORD_TOKEN` and `TARGET_CHANNEL_ID`.
2. From PowerShell run:

   ```powershell
   .\run-docker.ps1 -Build -Run
   ```

   Or build and run manually:

   ```powershell
   docker build -t whutbot3 .
   docker run --rm --env-file .env --name whutbot3 whutbot3
   ```

The container runs the compiled `whutbot3` binary. Ensure the bot has the Message Content intent enabled in the Discord Developer Portal if you need to read message contents.
