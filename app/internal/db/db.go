// Package db opens and holds the PostgreSQL connection pool.
//
// The driver is pgx through its database/sql compatibility layer. Using
// database/sql rather than pgx's native interface is deliberate: the weakness
// toggles swap between parameterised queries and concatenated SQL, and both
// are expressed the same way through database/sql, so the vulnerable and the
// hardened path differ by one string concatenation rather than by two
// different driver APIs.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// Registers the "pgx" driver with database/sql. Imported for its side
	// effect, which is why it is blank.
	_ "github.com/jackc/pgx/v5/stdlib"

	"clayface/app/internal/config"
)

// Open connects to PostgreSQL and waits for it to answer.
//
// The wait is not redundant with the Compose healthcheck. Compose's
// depends_on does hold the portal back until postgres reports healthy, but that
// guarantee only covers the first start: if postgres is later restarted (or the
// portal outlives a `docker compose restart postgres`), the pool's existing
// connections die and a fresh container has nothing to hold it back. Retrying
// for a bounded window makes the application survive that without a restart
// loop, and still fails with a clear message instead of hanging forever.
func Open(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	pool, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening database handle: %w", err)
	}

	// The portal is a low-traffic internal application, so a small pool is
	// plenty; the limits are mostly there to keep a runaway loop in a
	// deliberately vulnerable handler from exhausting postgres connections.
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(30 * time.Minute)

	const (
		attempts = 30
		delay    = 2 * time.Second
	)
	for attempt := 1; ; attempt++ {
		err = pool.PingContext(ctx)
		if err == nil {
			return pool, nil
		}
		if attempt >= attempts {
			pool.Close()
			return nil, fmt.Errorf("database at %s:%d did not become reachable after %d attempts: %w",
				cfg.DBHost, cfg.DBPort, attempts, err)
		}
		if attempt == 1 {
			log.Printf("waiting for database at %s:%d (%v)", cfg.DBHost, cfg.DBPort, err)
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
}
