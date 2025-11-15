package main

import (
	"log"
	"net/http"

	"github.com/BroBay24/WebsocketUTS/internal/config"
	"github.com/BroBay24/WebsocketUTS/internal/database"
	"github.com/BroBay24/WebsocketUTS/internal/handlers"
	"github.com/BroBay24/WebsocketUTS/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using environment variables")
	}

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	hub := websocket.NewHub()
	go hub.Run()

	r := gin.Default()

	r.Static("/css", "public/css")
	r.Static("/js", "public/js")
	r.Static("/assets", "public/assets")
	r.StaticFile("/", "public/form.html")
	r.StaticFile("/dashboard", "public/dashboard.html")

	handler := handlers.NewAttendanceHandler(db, hub)

	r.GET("/api/kehadiran", handler.List)
	r.GET("/export", handler.ExportExcel)
	r.GET("/ws", handler.HandleWebsocket)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	log.Printf("server running on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
