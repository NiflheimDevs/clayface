// Package config collects everything the portal is allowed to read from its
// environment.
//
// The whole application is configured through environment variables rather
// than a config file, because the deployment path is Docker Compose on the
// lab's APP01 VM: Ansible writes one .env file into the project directory and
// Compose turns it into environment variables for the containers.

// Every weakness toggle is an environment variable too. They exist so that a
// single binary can be run in two postures — deliberately vulnerable or hardened

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// --- DELIBERATE WEAKNESS: WEAK_HARDCODED_DB_CREDENTIALS ---
//
// The lab's database password also exists here, as a literal, compiled into
// the binary. It is used only when DB_PASSWORD is unset, which is what makes
// it a realistic finding rather than a decoration: the built image carries a
// working credential inside it, so anyone who obtains the image — or the
// source, or the /admin/config response — obtains a database password without
// ever touching the environment.
//
// It demonstrates the "secrets belong in configuration, not in code" rule.
// Turning the toggle off removes the fallback entirely (Load then refuses to
// start without DB_PASSWORD) and makes the two config endpoints redact the
// value instead of printing it.
const hardcodedDBPassword = "Clayface_DB_2024!"

// Weaknesses is the set of deliberately vulnerable code paths. Each field maps
// one-to-one onto an environment variable, and each is consulted at request
// time rather than at start-up so that the whole set can be flipped by
// recreating the container with a different .env.
//
// The names here are the short form; the environment variable names are the
// WEAK_* constants below.
type Weaknesses struct {
	SQLILogin                    bool
	SQLISearch                   bool
	CommandInjectionDiagnostics  bool
	IDORDocuments                bool
	BrokenAdminAuthz             bool
	PathTraversalDownload        bool
	HardcodedDBCredentials       bool
	ForgeableSessionToken        bool
	ServiceAccountCredentialInDB bool
	SSRFURLFetch                 bool
	VerboseErrorsDebugEndpoint   bool
}

// Config is the resolved configuration. Values are plain strings and ints:
// nothing here does any parsing beyond what Load already did, so handlers can
// read fields directly.
type Config struct {
	// Database connection. DBPassword is the effective password — either the
	// one from the environment or, under the hardcoded-credentials weakness,
	// the literal above. PasswordSource records which, so the config endpoints
	// can say so out loud.
	DBHost         string
	DBPort         int
	DBName         string
	DBUser         string
	DBPassword     string
	PasswordSource string

	// SessionSecret signs session cookies when the session token is not
	// forgeable. It is unused in the vulnerable posture, but it is still read
	// so that a misconfigured deployment is visible in /admin/config rather
	// than silently ignored.
	SessionSecret string

	// ServiceAccountUser and ServiceAccountPassword identify the application's
	// own service account. The password is seeded into the api_keys table when
	// the service-account weakness is on, which is how it becomes reachable
	// from a database dump.
	ServiceAccountUser     string
	ServiceAccountPassword string

	// DataDir is where the on-disk documents live; ListenAddr is the address
	// the HTTP listener binds to. Both are container-internal values.
	DataDir    string
	ListenAddr string

	Weak Weaknesses
}

// Environment variable names, in one place so the README, the deployment
// playbook and this file cannot drift apart.
const (
	EnvDBHost                     = "DB_HOST"
	EnvDBPort                     = "DB_PORT"
	EnvDBName                     = "DB_NAME"
	EnvDBUser                     = "DB_USER"
	EnvDBPassword                 = "DB_PASSWORD"
	EnvSessionSecret              = "SESSION_SECRET"
	EnvServiceAccountUser         = "SERVICE_ACCOUNT_USER"
	EnvServiceAccountPassword     = "SERVICE_ACCOUNT_PASSWORD"
	EnvDataDir                    = "DATA_DIR"
	EnvListenAddr                 = "LISTEN_ADDR"
	EnvWeakSQLILogin              = "WEAK_SQLI_LOGIN"
	EnvWeakSQLISearch             = "WEAK_SQLI_SEARCH"
	EnvWeakCommandInjection       = "WEAK_COMMAND_INJECTION_DIAGNOSTICS"
	EnvWeakIDORDocuments          = "WEAK_IDOR_DOCUMENTS"
	EnvWeakBrokenAdminAuthz       = "WEAK_BROKEN_ADMIN_AUTHZ"
	EnvWeakPathTraversal          = "WEAK_PATH_TRAVERSAL_DOWNLOAD"
	EnvWeakHardcodedDBCredentials = "WEAK_HARDCODED_DB_CREDENTIALS"
	EnvWeakForgeableSessionToken  = "WEAK_FORGEABLE_SESSION_TOKEN"
	EnvWeakServiceAccountInDB     = "WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB"
	EnvWeakSSRFURLFetch           = "WEAK_SSRF_URL_FETCH"
	EnvWeakVerboseErrors          = "WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT"
)

// Load reads the configuration from the process environment, applying the
// documented defaults. It does not validate; call Validate for that, so that
// the caller decides whether a problem is fatal.
func Load() *Config {
	cfg := &Config{
		DBHost:                 envOr(EnvDBHost, "postgres"),
		DBPort:                 envIntOr(EnvDBPort, 5432),
		DBName:                 envOr(EnvDBName, "clayface"),
		DBUser:                 envOr(EnvDBUser, "clayface_app"),
		SessionSecret:          envOr(EnvSessionSecret, ""),
		ServiceAccountUser:     envOr(EnvServiceAccountUser, "svc-app-portal"),
		ServiceAccountPassword: envOr(EnvServiceAccountPassword, ""),
		DataDir:                envOr(EnvDataDir, "/data"),
		ListenAddr:             envOr(EnvListenAddr, ":8080"),
	}

	cfg.Weak = Weaknesses{
		SQLILogin:                    envBoolOr(EnvWeakSQLILogin, true),
		SQLISearch:                   envBoolOr(EnvWeakSQLISearch, true),
		CommandInjectionDiagnostics:  envBoolOr(EnvWeakCommandInjection, true),
		IDORDocuments:                envBoolOr(EnvWeakIDORDocuments, true),
		BrokenAdminAuthz:             envBoolOr(EnvWeakBrokenAdminAuthz, true),
		PathTraversalDownload:        envBoolOr(EnvWeakPathTraversal, true),
		HardcodedDBCredentials:       envBoolOr(EnvWeakHardcodedDBCredentials, true),
		ForgeableSessionToken:        envBoolOr(EnvWeakForgeableSessionToken, true),
		ServiceAccountCredentialInDB: envBoolOr(EnvWeakServiceAccountInDB, true),
		SSRFURLFetch:                 envBoolOr(EnvWeakSSRFURLFetch, true),
		VerboseErrorsDebugEndpoint:   envBoolOr(EnvWeakVerboseErrors, true),
	}

	// Resolve the database password. The environment always wins when it is
	// set; the compiled-in literal is only the fallback, and only exists at all
	// while the hardcoded-credentials weakness is enabled. With the toggle off
	// there is deliberately no fallback — Validate rejects the empty password —
	// which is what a hardened deployment looks like.
	switch {
	case os.Getenv(EnvDBPassword) != "":
		cfg.DBPassword = os.Getenv(EnvDBPassword)
		cfg.PasswordSource = "environment"
	case cfg.Weak.HardcodedDBCredentials:
		cfg.DBPassword = hardcodedDBPassword
		cfg.PasswordSource = "hardcoded-fallback-in-binary"
	default:
		cfg.DBPassword = ""
		cfg.PasswordSource = ""
	}

	return cfg
}

// Validate reports the configuration problems that must stop the process. The
// only one that is fatal is a missing database password, and only because it
// means the container could never connect: failing here with a clear message is
// better than failing later with a driver-level authentication error that
// looks like a database problem.
func (c *Config) Validate() error {
	if c.DBPassword == "" {
		return fmt.Errorf("%s is required: no hardcoded fallback exists while %s=false",
			EnvDBPassword, EnvWeakHardcodedDBCredentials)
	}
	if !c.Weak.ForgeableSessionToken && c.SessionSecret == "" {
		return fmt.Errorf("%s is required to sign session cookies while %s=false",
			EnvSessionSecret, EnvWeakForgeableSessionToken)
	}
	return nil
}

// DSN renders the PostgreSQL connection string in the keyword/value form
// rather than the URL form. The password is generated by Ansible and may
// contain characters that would need percent-encoding in a URL, and the
// keyword/value form has no such requirement.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword)
}

// DocumentsDir is the directory the download endpoint serves from and the
// directory the seed materialises the sample documents into.
func (c *Config) DocumentsDir() string {
	return c.DataDir + "/documents"
}

// Settings renders the effective configuration as a flat map for the two
// config endpoints (/admin/config and /api/debug/config).
//
// --- DELIBERATE WEAKNESS: WEAK_HARDCODED_DB_CREDENTIALS ---
//
// With the toggle on, the map carries the live database password and the fact
// that it came from a compiled-in literal. Together those two fields are the
// disclosure: a reader of this response learns a working credential and learns
// that it is a constant, so it is the same on every deployment of this image.
// With the toggle off the value is replaced by a fixed redaction and the
// source is reported as "redacted" instead.
func (c *Config) Settings() map[string]string {
	password := "[redacted]"
	source := "redacted"
	if c.Weak.HardcodedDBCredentials {
		password = c.DBPassword
		source = c.PasswordSource
	}

	return map[string]string{
		"db_host":              c.DBHost,
		"db_port":              strconv.Itoa(c.DBPort),
		"db_name":              c.DBName,
		"db_user":              c.DBUser,
		"db_password":          password,
		"db_password_source":   source,
		"data_dir":             c.DataDir,
		"listen_addr":          c.ListenAddr,
		"service_account_user": c.ServiceAccountUser,
		"session_secret_set":   strconv.FormatBool(c.SessionSecret != ""),
	}
}

// WeaknessReport describes every toggle and its current state. The admin
// diagnostics page prints it, and it is the quickest way to answer "is this
// deployment vulnerable to X" without reading the .env on the host.
func (c *Config) WeaknessReport() map[string]bool {
	return map[string]bool{
		EnvWeakSQLILogin:              c.Weak.SQLILogin,
		EnvWeakSQLISearch:             c.Weak.SQLISearch,
		EnvWeakCommandInjection:       c.Weak.CommandInjectionDiagnostics,
		EnvWeakIDORDocuments:          c.Weak.IDORDocuments,
		EnvWeakBrokenAdminAuthz:       c.Weak.BrokenAdminAuthz,
		EnvWeakPathTraversal:          c.Weak.PathTraversalDownload,
		EnvWeakHardcodedDBCredentials: c.Weak.HardcodedDBCredentials,
		EnvWeakForgeableSessionToken:  c.Weak.ForgeableSessionToken,
		EnvWeakServiceAccountInDB:     c.Weak.ServiceAccountCredentialInDB,
		EnvWeakSSRFURLFetch:           c.Weak.SSRFURLFetch,
		EnvWeakVerboseErrors:          c.Weak.VerboseErrorsDebugEndpoint,
	}
}

// envOr returns the environment value, or the default when the variable is
// unset or empty. Empty counts as unset on purpose: Compose passes an empty
// string through for a variable that is declared but blank, and treating that
// as "use the default" is what makes a half-filled .env behave sensibly.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envIntOr is envOr for the one integer setting (the database port). A value
// that is not a number falls back to the default rather than failing the
// process, because a wrong port produces a clear connection error anyway.
func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return n
}

// envBoolOr reads a weakness toggle. Unset means enabled — the lab defaults to
// the vulnerable posture, so the interesting behaviour is what happens by
// default and hardening is the explicit act. Anything that is not a
// recognisable boolean is also treated as enabled, for the same reason: a typo
// should leave the lab vulnerable (and obvious) rather than quietly safe.
func envBoolOr(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
