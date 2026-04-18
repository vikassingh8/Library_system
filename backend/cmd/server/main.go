package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/library_system/internal/blob"
	"github.com/library_system/internal/config"
	"github.com/library_system/internal/handlers"
	"github.com/library_system/internal/middleware"

	"github.com/library_system/internal/storage/azuresql"
)

func main() {
	cfg := config.MustLoad() // reads from .env (local) or Azure App Settings (production)
	storage, err := azuresql.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("✅ Azure SQL connected",
		slog.String("server", cfg.Database.Server),
		slog.String("database", cfg.Database.Name),
	)

	// Initialise Azure Blob Storage client
	blobClient, err := blob.NewClientFromConnString(
		cfg.AzureBlob.ConnectionString,
		cfg.AzureBlob.ContainerName,
	)
	if err != nil {
		log.Fatal("❌ Blob client init failed:", err)
	}
	slog.Info("✅ Azure Blob Storage ready", slog.String("container", cfg.AzureBlob.ContainerName))

	router := http.NewServeMux()
	handler := middleware.Recover(middleware.CORS(cfg.CORS.AllowedOrigin)(middleware.LoggerMiddleware(router)))

	auth := middleware.AuthMiddleware(cfg, storage)

	// Public routes (no authentication)
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("✅ Backend is healthy"))
	})
	router.Handle("POST /register", handlers.Register(storage))
	router.Handle("POST /login", handlers.Login(storage, cfg))
	router.Handle("POST /logout", handlers.Logout())

	// Image upload (admin only — only admins add books)
	router.Handle("POST /upload-image", auth(middleware.AdminOnly(handlers.UploadImage(blobClient))))

	// Book routes — read: any authenticated user; write: admin only
	router.Handle("GET /books", auth(handlers.GetAllBooks(storage)))
	router.Handle("GET /books/search", auth(handlers.SearchBooks(storage)))
	router.Handle("GET /books/{id}", auth(handlers.GetBookByID(storage)))
	router.Handle("POST /books", auth(middleware.AdminOnly(handlers.CreateBook(storage))))
	router.Handle("PUT /books/{id}", auth(middleware.AdminOnly(handlers.UpdateBook(storage))))
	router.Handle("DELETE /books/{id}", auth(middleware.AdminOnly(handlers.DeleteBook(storage))))

	// Borrow routes (any authenticated user)
	router.Handle("POST /books/{id}/borrow", auth(handlers.BorrowBook(storage)))
	router.Handle("POST /borrows/{id}/return", auth(handlers.ReturnBook(storage)))
	router.Handle("GET /my-borrows", auth(handlers.GetMyBorrows(storage)))

	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// run server in goroutine
	go func() {
		log.Println("✅ Server started on", cfg.HttpServer.Address)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("❌ Server failed:", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("❌ Shutdown failed:", err)
	}

	log.Println("✅ Server stopped gracefully")
}
