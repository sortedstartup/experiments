package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed ui/*
var uiFS embed.FS

// NewApp creates and configures the Echo application
func NewApp() *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files from ui folder
	uiSubFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		panic(err)
	}
	e.GET("/", echo.WrapHandler(http.FileServer(http.FS(uiSubFS))))

	// API Routes
	e.GET("/api/health", healthHandler)

	return e
}

// healthHandler returns server health status
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Server is running",
	})
}
