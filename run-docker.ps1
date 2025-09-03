<#
Build and run the whutbot3 Docker image.
Usage (PowerShell):
  .\run-docker.ps1 -Build -Run

This script expects a `.env` file with DISCORD_TOKEN and TARGET_CHANNEL_ID, or you may set those environment variables on the host.
#>

param(
    [switch]$Build,
    [switch]$Run
)

if ($Build) {
    Write-Host "Building Docker image 'whutbot3'..."
    docker build -t whutbot3 .
}

if ($Run) {
    if (-Not (Test-Path ".env")) {
        Write-Warning ".env not found. Make sure DISCORD_TOKEN and TARGET_CHANNEL_ID are provided via .env or host environment variables."
    }

    Write-Host "Running container 'whutbot3' (will remove on exit)..."
    docker run --rm --env-file .env --name whutbot3 whutbot3
}
