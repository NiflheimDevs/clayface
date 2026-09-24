// Command portal is the Clayface Portal: one binary serving the public
// customer portal, the JSON API and the internal admin panel on a single
// listener.
//
// It runs as one process inside one container on the lab's application host,
// with PostgreSQL alongside it in a second container. There is no supervisor
// and no reload: configuration comes from the environment at start-up, and
// changing a weakness toggle means recreating the container.
//
// Start-up is deliberately ordered so that a misconfigured deployment fails
// loudly and early rather than half-working:
//
//  1. load and validate the configuration,
//  2. connect to PostgreSQL (retrying briefly, in case it is still starting),
//  3. apply the schema and the seed if the database is empty,
//  4. materialise the document store if it is empty,
//  5. serve.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clayface/app/internal/config"
	"clayface/app/internal/db"
	"clayface/app/internal/seed"
	"clayface/app/internal/web"
)

func main() {
	// UTC in the container log so that log lines line up with the audit table,
	// which stores timestamptz.
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("portal: ")

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	// The listener stays up until SIGINT or SIGTERM, and the same context
	// cancels the database work below, so a signal during start-up stops the
	// retry loop instead of waiting it out.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	defer pool.Close()

	if err := seed.Ensure(ctx, pool, cfg); err != nil {
		log.Fatalf("seeding database: %v", err)
	}

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: web.NewServer(cfg, pool),
		// Slow-loris guard. The portal sits behind nginx, which buffers
		// requests, so a long header read is only ever an idle client.
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Printf("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	reportPosture(cfg)
	log.Printf("listening on %s, data directory %s", cfg.ListenAddr, cfg.DataDir)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server: %v", err)
	}
}

// reportPosture writes the enabled weakness toggles to the log at start-up.
//
// It exists because the lab's most common confusion is not knowing which
// posture a deployment is in: the same image runs both, and a toggle that was
// meant to be flipped on and was not looks exactly like a vulnerability that
// does not work. Printing the enabled set turns that into a one-line check.
func reportPosture(cfg *config.Config) {
	enabled := make([]string, 0, len(cfg.WeaknessReport()))
	for name, on := range cfg.WeaknessReport() {
		if on {
			enabled = append(enabled, name)
		}
	}

	if len(enabled) == 0 {
		log.Printf("posture: hardened — every weakness toggle is off")
		return
	}

	log.Printf("posture: vulnerable — %d weakness toggles enabled: %v", len(enabled), enabled)
}
