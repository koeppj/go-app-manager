package main

import (
	"flag"
	"log"

	"github.com/koeppj/go-app-manager/internal/trayui"
)

func main() {
	serviceURL := flag.String("service-url", "https://localhost:8443", "Service base URL")
	token := flag.String("token", "", "Bearer token override")
	ca := flag.String("ca", "", "Custom CA certificate")
	flag.Parse()

	if err := trayui.Run(*serviceURL, *token, *ca); err != nil {
		log.Fatalf("tray error: %v", err)
	}
}
