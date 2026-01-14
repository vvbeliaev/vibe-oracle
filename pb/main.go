package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"svpb-tmpl/pkg/config"
	"svpb-tmpl/pkg/indexer"
	"svpb-tmpl/pkg/parser"
	"svpb-tmpl/pkg/rag"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	_ "svpb-tmpl/migrations"
)

func main() {
	godotenv.Load("../.env")

	app := pocketbase.New()

	// Load configuration
	cfg := config.Load()

	// Static file serving with SPA fallback
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", func(e *core.RequestEvent) error {
			path := e.Request.PathValue("path")
			fsys := os.DirFS("./pb_public")

			// PWA treatment block - no cache for service worker files
			if path == "sw.js" || strings.HasPrefix(path, "workbox-") {
				e.Response.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				e.Response.Header().Set("Pragma", "no-cache")
				e.Response.Header().Set("Expires", "0")
				return e.FileFS(fsys, path)
			}

			// Try to serve static file
			err := e.FileFS(fsys, path)
			if err == nil {
				return nil
			}

			// SPA fallback (ignore API and admin routes)
			if !strings.HasPrefix(path, "api/") && !strings.HasPrefix(path, "_/") && !strings.Contains(path, ".") {
				return e.FileFS(fsys, "index.html")
			}

			return nil
		})

		return se.Next()
	})

	// Auto-migrations in dev mode
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
		Dir:         "migrations",
	})

	// --- Telegram Integration ---

	// Add tg-login command for interactive Telegram authentication
	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "tg-login",
		Short: "Login to Telegram and save session",
		Long:  "Performs interactive Telegram authentication. Use this to generate session.json before deploying to server.",
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			if cfg.TgAPIID == 0 || cfg.TgAPIHash == "" {
				logger.Fatal("TG_API_ID and TG_API_HASH must be set")
			}

			parserCfg := parser.Config{
				APIID:       cfg.TgAPIID,
				APIHash:     cfg.TgAPIHash,
				Phone:       cfg.TgPhone,
				SessionPath: cfg.TgSessionPath,
			}

			client := parser.NewClient(parserCfg, logger)

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			if err := client.Login(ctx); err != nil {
				logger.Fatal("Login failed", zap.Error(err))
			}
		},
	})

	// Start services when server starts
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		logger, _ := zap.NewProduction()

		// Initialize indexer service (needed for both parser and RAG)
		indexerSvc, err := indexer.NewService(app, cfg, logger)
		if err != nil {
			log.Printf("Failed to initialize indexer: %v", err)
			return se.Next()
		}

		// Ensure MeiliSearch index is configured
		ctx := context.Background()
		if err := indexerSvc.EnsureIndex(ctx); err != nil {
			log.Printf("Failed to configure MeiliSearch index: %v", err)
		}

		// Initialize RAG service
		ragSvc := rag.NewService(app, indexerSvc, cfg, logger)

		// Register chat API route
		se.Router.POST("/api/chat", ragSvc.HandleChat)
		se.Router.GET("/api/chats/{chatId}/sse", ragSvc.HandleChatSSE)

		// Start Telegram parser if configured
		if cfg.TgAPIID != 0 && cfg.TgAPIHash != "" {
			// Check if session file exists
			if _, err := os.Stat(cfg.TgSessionPath); os.IsNotExist(err) {
				log.Printf("Session file not found at %s - run 'tg-login' first", cfg.TgSessionPath)
			} else {
				// Start the parser in background
				go startTelegramParser(cfg, indexerSvc, logger)
			}
		} else {
			log.Println("Telegram not configured (TG_API_ID/TG_API_HASH missing), skipping parser")
		}

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// startTelegramParser runs the Telegram message listener in the background.
func startTelegramParser(cfg *config.Config, indexerSvc *indexer.Service, logger *zap.Logger) {
	defer logger.Sync()

	// Create handler
	handler := parser.NewHandler(cfg, indexerSvc, logger)

	// Create Telegram client
	parserCfg := parser.Config{
		APIID:       cfg.TgAPIID,
		APIHash:     cfg.TgAPIHash,
		Phone:       cfg.TgPhone,
		SessionPath: cfg.TgSessionPath,
	}
	tg := parser.NewClient(parserCfg, logger)
	tg.OnNewMessage(handler.HandleMessage)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Info("Starting Telegram parser...",
		zap.Int("targetChats", len(cfg.TargetChatIDs)),
	)

	if err := tg.Start(ctx); err != nil {
		if err != context.Canceled {
			logger.Error("Telegram parser error", zap.Error(err))
		}
	}

	logger.Info("Telegram parser stopped")
}
