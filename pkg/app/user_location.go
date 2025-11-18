package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	config "osom/pkg"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUserLocation() {
	// Spawn a web server with two endpoints:
	// 1. GET /location - serves an HTML page with a button to request access to user location, then posts it to the other endpoint
	// 2. POST /location - receives the user's latitude and longitude and stores them in config.Config.DefaultLatLong

	router := gin.New()
	router.LoadHTMLGlob("templates/*")

	var userLatLong string

	server := &http.Server{
		Addr:    ":4749",
		Handler: router,
	}

	router.GET("/location", func(c *gin.Context) {
		// Return JSON response
		c.HTML(http.StatusOK, "request_location.html", nil)
	})

	router.POST("/location", func(c *gin.Context) {
		userLatLong = c.Query("latitude") + "," + c.Query("longitude")

		// Close the server in a goroutine to prevent deadlocks (server is waiting for this handler to finish)
		go func() {
			ctx := context.Background()
			if err := server.Shutdown(ctx); err != nil {
				log.Fatalf("Server Shutdown Failed:%v", err)
				return
			}
		}()

		c.String(http.StatusNoContent, "")
	})

	url := "http://localhost:4749/location"
	fmt.Printf("Please grant permission to access your current location in your browser at %s...", url)
	go func() {
		// time.Sleep(200 * time.Millisecond) // small delay to ensure server is listening
		// Overkill version:
		timer := time.NewTimer(200 * time.Millisecond)
		<-timer.C
		if err := openBrowser(url); err != nil {
			log.Fatalf("failed to open browser: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if userLatLong != "" {
			config.Config.DefaultLatLong = userLatLong
		} else {
			log.Fatalf("failed to run server: %v", err)
		}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
