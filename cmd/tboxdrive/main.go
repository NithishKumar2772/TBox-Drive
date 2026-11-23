package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/NithishKumar2772/TBox-Drive/internal/app"
	"github.com/NithishKumar2772/TBox-Drive/internal/app/web"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

func main() {
	if err := run(); err != nil {
		log.Printf("TBox Drive web server stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	clientID := os.Getenv("TBOX_TERABOX_CLIENT_ID")
	clientSecret := os.Getenv("TBOX_TERABOX_CLIENT_SECRET")
	privateSecret := os.Getenv("TBOX_TERABOX_PRIVATE_SECRET")
	applicationRoot := os.Getenv("TBOX_TERABOX_APP_FOLDER")
	if clientID == "" || clientSecret == "" || privateSecret == "" || applicationRoot == "" {
		return fmt.Errorf("configure TBOX_TERABOX_CLIENT_ID, TBOX_TERABOX_CLIENT_SECRET, TBOX_TERABOX_PRIVATE_SECRET and TBOX_TERABOX_APP_FOLDER before starting")
	}

	configDir, err := os.UserConfigDir()
	if err != nil { return fmt.Errorf("locate user configuration directory: %w", err) }
	dataDir := filepath.Join(configDir, "TBox Drive")
	if err := os.MkdirAll(dataDir, 0700); err != nil { return fmt.Errorf("create application data directory: %w", err) }
	store, err := sqlite.Open(filepath.Join(dataDir, "tbox.db"))
	if err != nil { return fmt.Errorf("open local database: %w", err) }
	defer store.Close()

	oauth, err := terabox.NewOAuthClient(terabox.OAuthConfig{
		ClientID: clientID, ClientSecret: clientSecret, PrivateSecret: privateSecret,
	})
	if err != nil { return fmt.Errorf("configure TeraBox OAuth: %w", err) }
	secrets, err := app.NewOSSecretStore()
	if err != nil { return fmt.Errorf("initialize operating-system credential vault: %w", err) }
	service, err := app.NewTeraBoxService(oauth, secrets, store, store, applicationRoot)
	if err != nil { return fmt.Errorf("initialize TeraBox service: %w", err) }
	server, err := web.NewServer(service, store, applicationRoot)
	if err != nil { return fmt.Errorf("initialize local web server: %w", err) }

	address := os.Getenv("TBOX_WEB_ADDR")
	if address == "" { address = "127.0.0.1:8765" }
	host, _, err := net.SplitHostPort(address)
	if err != nil { return fmt.Errorf("invalid TBOX_WEB_ADDR: %w", err) }
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() { return fmt.Errorf("TBOX_WEB_ADDR must use a loopback IP address; remote access is disabled") }

	log.Printf("TBox Drive browser client: http://%s", address)
	log.Printf("Local database: %s", filepath.Join(dataDir, "tbox.db"))
	return server.ListenAndServe(ctx, address)
}
