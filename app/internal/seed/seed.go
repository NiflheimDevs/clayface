// Package seed applies the application schema and the demo data, and
// materialises the sample documents the portal serves.
//
// Everything here runs once, at first boot: the schema and the seed are applied
// only when the users table is absent, so a restart never duplicates data and
// never overwrites what a student has since changed. The one exception is the
// application service-account row, which is rewritten on every boot from the
// environment — see ensureServiceAccountRow for why.
//
// The SQL lives in embedded files rather than in Go string literals so that it
// can be read, reviewed and pasted into psql without going through the binary.
package seed

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"clayface/app/internal/config"
)

// files carries the SQL and the sample documents into the binary. The
// Dockerfile copies no SQL at all: what the container runs is exactly what was
// compiled in, which is what makes the vulnerability toggles the only way the
// data set can differ between deployments.
//
//go:embed schema.sql seed.sql documents
var files embed.FS

// serviceAccountHashPlaceholder stands in for the application service account's
// credential when the WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB toggle is off.
//
// It is deliberately shaped like a bcrypt digest ($2a$12$ followed by 53
// characters of salt-and-hash alphabet) and deliberately is not one: it is a
// fixed string chosen for this file, not a hash of any password, so there is no
// plaintext anywhere that it corresponds to. The point of the hardened posture
// is that the table no longer yields a usable credential, and a placeholder
// that merely looks like a hash makes that visible in a dump without pretending
// the app has a password-verification path it does not have.
const serviceAccountHashPlaceholder = "$2a$12$Xk8fVb2mQq4hY0wR3nJ5eOa9tL1cZ6pS7dG8uH2iK4vB0nM3jQ5y2"

// Ensure brings the database up to date with the embedded schema and seed, then
// materialises the document store on disk.
func Ensure(ctx context.Context, db *sql.DB, cfg *config.Config) error {
	applied, err := schemaApplied(ctx, db)
	if err != nil {
		return err
	}

	if !applied {
		log.Printf("empty database: applying schema.sql and seed.sql")
		if err := applyFile(ctx, db, "schema.sql"); err != nil {
			return fmt.Errorf("applying schema: %w", err)
		}
		if err := applyFile(ctx, db, "seed.sql"); err != nil {
			return fmt.Errorf("applying seed data: %w", err)
		}
	}

	if err := ensureServiceAccountRow(ctx, db, cfg); err != nil {
		return fmt.Errorf("seeding service account row: %w", err)
	}
	if err := ensureDocuments(cfg.DocumentsDir()); err != nil {
		return fmt.Errorf("materialising documents: %w", err)
	}

	return nil
}

// schemaApplied reports whether the schema is already there, by looking for the
// users table specifically. Any of the tables would do; users is the one every
// other table hangs off, so it is the one that is meaningless to have alone.
func schemaApplied(ctx context.Context, db *sql.DB) (bool, error) {
	const query = `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'users'
	)`
	var exists bool
	if err := db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking for existing schema: %w", err)
	}
	return exists, nil
}

// applyFile runs one embedded .sql file statement by statement, inside a single
// transaction so that a file which fails halfway leaves no partial schema
// behind.
//
// The statements are split here rather than by the driver because the pgx
// driver only accepts one statement per call when it uses the extended
// protocol, and it uses the extended protocol as soon as a query carries
// parameters. Splitting is a deliberate simplification: it understands
// full-line comments and a statement terminator at end of line, which is all
// these two files use. It would need replacing before it could swallow
// functions, dollar-quoted bodies or a trailing comment on the same line as a
// statement.
func applyFile(ctx context.Context, db *sql.DB, name string) error {
	content, err := files.ReadFile(name)
	if err != nil {
		return fmt.Errorf("reading embedded %s: %w", name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		// Rolled back explicitly on the error paths below; this covers the
		// case where a later statement fails and the deferred call is what
		// guarantees the transaction does not stay open.
		_ = tx.Rollback()
	}()

	for _, statement := range splitStatements(string(content)) {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("%s: %w", firstLine(statement), err)
		}
	}

	return tx.Commit()
}

// splitStatements breaks a .sql file into statements at end-of-line semicolons,
// dropping blank lines and full-line comments. Comments are dropped rather than
// passed through because a comment-only "statement" is a wasted round trip, and
// because dropping them keeps the statement text that reaches the driver and
// the error messages readable.
func splitStatements(content string) []string {
	var (
		statements []string
		current    strings.Builder
	)

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		statements = append(statements, current.String())
	}

	return statements
}

// firstLine names a statement in an error message without dumping the whole
// thing. The first line of every statement in these files is the part that says
// what it was doing.
func firstLine(statement string) string {
	if idx := strings.IndexByte(statement, '\n'); idx >= 0 {
		return strings.TrimSpace(statement[:idx])
	}
	return strings.TrimSpace(statement)
}

// ensureServiceAccountRow writes the application service account's credential
// into api_keys, and is the only piece of seeding that runs on every boot —
// because the credential comes from the environment, and the environment is
// allowed to change between deployments.
//
// --- DELIBERATE WEAKNESS: WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB ---
//
// With the toggle on, the row is updated to hold the plaintext
// SERVICE_ACCOUNT_PASSWORD. A service account's password sitting in a business
// table, in clear text, next to API keys that are also in clear text, is the
// weakness: the credential is reachable by anything that can read the table —
// a UNION SELECT through the search endpoint, a database dump, a backup — and
// it is the same credential the application itself uses to authenticate to
// other internal services, so it is a real pivot and not a decoy.
//
// With the toggle off the row keeps the placeholder digest instead, so the
// table's shape is identical in both postures and only its contents differ.
// That is what lets a validation playbook assert the difference by reading one
// column rather than by comparing schemas.
func ensureServiceAccountRow(ctx context.Context, db *sql.DB, cfg *config.Config) error {
	credential := serviceAccountHashPlaceholder
	if cfg.Weak.ServiceAccountCredentialInDB {
		if cfg.ServiceAccountPassword == "" {
			// Not fatal: the rest of the application works, and the hardened
			// posture is a legitimate deployment too. But it is worth a line in
			// the log, because the intended lab deployment always sets it and a
			// silent fall back to the placeholder would look like the toggle
			// had not been applied.
			log.Printf("SERVICE_ACCOUNT_PASSWORD is unset: the %s row keeps its placeholder",
				cfg.ServiceAccountUser)
		} else {
			credential = cfg.ServiceAccountPassword
		}
	}

	const upsert = `
		INSERT INTO api_keys (key_name, service_account, api_key, credential_type, notes)
		VALUES ($1, $2, $3, 'application-account', $4)
		ON CONFLICT (service_account) DO UPDATE
			SET api_key = EXCLUDED.api_key,
			    key_name = EXCLUDED.key_name,
			    notes = EXCLUDED.notes`

	notes := "Credential for the portal's own service account. Rewritten at every boot from SERVICE_ACCOUNT_PASSWORD."
	_, err := db.ExecContext(ctx, upsert, "portal-internal-api", cfg.ServiceAccountUser, credential, notes)
	return err
}

// ensureDocuments writes the embedded sample documents into the data directory
// if, and only if, it is empty.
//
// The emptiness check is the whole point: this directory is inside a named
// volume that survives redeployment, and a real deployment would have real
// documents in it. Overwriting whatever is there would destroy the thing a
// student is supposed to be exfiltrating, so the embedded set is a first-boot
// convenience and never a sync.
func ensureDocuments(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	existing, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		log.Printf("document store at %s already has %d entries: leaving it alone", dir, len(existing))
		return nil
	}

	entries, err := fs.ReadDir(files, "documents")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := files.ReadFile("documents/" + entry.Name())
		if err != nil {
			return err
		}
		// 0640: the files are readable by the portal's own user and by whoever
		// shares its group, which is how a container running as a non-root user
		// can serve them without the directory being world-readable.
		target := filepath.Join(dir, entry.Name())
		if err := os.WriteFile(target, content, 0o640); err != nil {
			return err
		}
		log.Printf("wrote sample document %s", target)
	}

	return nil
}
