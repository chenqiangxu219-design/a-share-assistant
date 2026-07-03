package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort        string
	WSPort          string
	FrontendURL     string
	CacheSize       int
	RefreshSec      int
	PythonURL       string // Python microservice URL
	IwencaiAPIKey   string // iwencai API key
}

func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if port == 0 {
		port = 8080
	}
	wsPort, _ := strconv.Atoi(os.Getenv("WS_PORT"))
	if wsPort == 0 {
		wsPort = 8090
	}
	cacheSize, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	if cacheSize == 0 {
		cacheSize = 10000
	}
	refreshSec, _ := strconv.Atoi(os.Getenv("REFRESH_SEC"))
	if refreshSec == 0 {
		refreshSec = 3
	}

	return &Config{
		HTTPPort:      strconv.Itoa(port),
		WSPort:        strconv.Itoa(wsPort),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
		CacheSize:     cacheSize,
		RefreshSec:    refreshSec,
		PythonURL:     os.Getenv("PYTHON_SERVICE_URL"),
		IwencaiAPIKey: os.Getenv("IWENCAI_API_KEY"),
	}
}
