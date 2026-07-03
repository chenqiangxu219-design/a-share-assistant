package main

import (
	"a-share-assistant/backend/cmd"
	"a-share-assistant/backend/config"
)

func main() {
	cfg := config.Load()
	cmd.StartServer(cfg)
}
