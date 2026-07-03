package cmd

import (
	"log"
	"time"

	apiv1 "a-share-assistant/backend/api"
	"a-share-assistant/backend/cache"
	"a-share-assistant/backend/config"
	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/ws"

	"github.com/gin-gonic/gin"
)

func StartServer(cfg *config.Config) {
	// Initialize data sources
	ds := datasource.NewManager(
		datasource.NewEastMoney(),
		datasource.NewSina(),
		datasource.NewTencent(),
	)

	// Initialize Python microservice client (mootdx + akshare)
	var pyClient *datasource.PyClient
	if cfg.PythonURL != "" {
		pyClient = datasource.NewPyClient(cfg.PythonURL)
		ds.SetPyClient(pyClient)
		if pyClient.IsReady() {
			log.Printf("Python microservice connected: %s", cfg.PythonURL)
		} else {
			log.Printf("WARNING: Python microservice not available at %s", cfg.PythonURL)
		}
	} else {
		log.Printf("WARNING: PYTHON_SERVICE_URL not set, mootdx/akshare disabled")
	}

	// Initialize 同花顺热点
	ithomer := datasource.NewIthomer()
	ds.SetIthomer(ithomer)

	// Initialize iwencai (NL search)
	iwencai := datasource.NewIwencai()
	ds.SetIwencai(iwencai)

	// Set default manager for news package
	datasource.SetDefaultManager(ds)

	// Initialize cache (5 min TTL for quotes, 30 min for K-lines)
	store, err := cache.NewStore(cache.DefaultDSN())
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
	quoteCache := apiv1.NewQuoteCache(store)

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handler
	handler := apiv1.NewHandler(ds, quoteCache, hub)
	handler.Portfolio = apiv1.NewPortfolioManager()

	// Set Python service client for full market data
	if cfg.PythonURL != "" {
		handler.SetPyClient(pyClient)
	}

	// Setup routes
	r := gin.Default()
	r.StaticFile("/favicon.ico", "")

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		handler.RegisterQuoteRoutes(api)
		handler.RegisterPortfolioRoutes(api)
		handler.RegisterBacktestRoutes(api)
		apiv1.RegisterChatRoutes(api, handler)
		handler.RegisterCapitalRoutes(api)
		handler.RegisterDataRoutes(api)
		handler.RegisterSectorRoutes(api)
		handler.Alerts.RegisterAlertRoutes(api)
	}

	r.GET("/ws", handler.WSHandler)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start background refresh for watchlist
	go startBackgroundRefresh(handler, cfg.RefreshSec)

	log.Printf("Server starting on port %s", cfg.HTTPPort)
	log.Fatal(r.Run(":" + cfg.HTTPPort))
}

// startBackgroundRefresh periodically refreshes quotes for cached stocks
func startBackgroundRefresh(handler *apiv1.Handler, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		codes := handler.Cache.GetCachedCodes()
		for _, code := range codes {
			q, err := handler.DS.GetQuote(code)
			if err != nil {
				continue
			}
			handler.Cache.SetQuote(code, q)
			handler.Hub.BroadcastQuote(q)
		}
	}
}
