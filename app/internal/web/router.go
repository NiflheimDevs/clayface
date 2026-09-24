// Package web holds the portal's HTTP surface: one listener, one binary, three
// route groups.
//
// The three groups are the public portal (/portal/*), the JSON API (/api/*)
// and the internal admin panel (/admin/*), and they are deliberately not
// separate services. The design decision recorded for this tier was "one
// binary, several route groups", because the point of the exercise is the
// application's weaknesses rather than its deployment topology, and because a
// single process is what makes a route-by-route hardening toggle meaningful —
// the same request path either concatenates a string or binds a parameter,
// depending on one environment variable.
//
// This file is the shared plumbing: the Server type, the routing table, the
// middleware, and the small helpers the three handler files have in common.
package web

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"strings"

	"clayface/app/internal/config"
)

// Server carries the two things every handler needs: the resolved configuration
// (which answers "is this weakness on") and the database pool.
type Server struct {
	cfg *config.Config
	db  *sql.DB
}

// NewServer builds the routing table and wraps it in the middleware chain.
func NewServer(cfg *config.Config, db *sql.DB) http.Handler {
	s := &Server{cfg: cfg, db: db}
	mux := http.NewServeMux()

	// The health endpoint is unauthenticated and deliberately trivial: the
	// container healthcheck runs as the unprivileged user with no credentials,
	// so anything that needed a session would be a healthcheck that always
	// fails. It reports readiness, not the state of the lab.
	mux.HandleFunc("GET /healthz", s.handleHealth)

	// The bare root is only a convenience so that browsing to the host does not
	// produce a 404 in the middle of a demo.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/portal/", http.StatusFound)
	})

	// --- /portal/* — the public portal -------------------------------------
	mux.HandleFunc("GET /portal/{$}", s.handlePortalIndex)
	mux.HandleFunc("GET /portal/login", s.handlePortalLoginForm)
	mux.HandleFunc("POST /portal/login", s.handlePortalLogin)
	mux.HandleFunc("GET /portal/logout", s.handlePortalLogout)
	mux.HandleFunc("GET /portal/profile", s.requireSession(s.handlePortalProfile))
	mux.HandleFunc("GET /portal/search", s.handlePortalSearch)
	mux.HandleFunc("GET /portal/download", s.handlePortalDownload)

	// --- /api/* — the JSON API --------------------------------------------
	mux.HandleFunc("POST /api/login", s.handleAPILogin)
	mux.HandleFunc("GET /api/profile", s.requireSession(s.handleAPIProfile))
	mux.HandleFunc("GET /api/orders", s.requireSession(s.handleAPIOrders))
	mux.HandleFunc("GET /api/orders/{id}", s.handleAPIOrder)
	mux.HandleFunc("GET /api/invoices/{id}", s.handleAPIInvoice)
	mux.HandleFunc("GET /api/search", s.handleAPISearch)
	mux.HandleFunc("GET /api/keys", s.requireSession(s.handleAPIKeys))
	mux.HandleFunc("GET /api/fetch", s.handleAPIFetch)
	mux.HandleFunc("GET /api/debug/config", s.handleAPIDebugConfig)

	// --- /admin/* — the internal panel ------------------------------------
	//
	// Every admin route goes through requireAdmin, which is where the
	// header-trust weakness lives. Wrapping all five in one place is what makes
	// the weakness a property of the route group rather than of five handlers.
	mux.HandleFunc("GET /admin/{$}", s.requireAdmin(s.handleAdminIndex))
	mux.HandleFunc("GET /admin/users", s.requireAdmin(s.handleAdminUsers))
	mux.HandleFunc("GET /admin/audit", s.requireAdmin(s.handleAdminAudit))
	mux.HandleFunc("GET /admin/diagnostics", s.requireAdmin(s.handleAdminDiagnostics))
	mux.HandleFunc("GET /admin/config", s.requireAdmin(s.handleAdminConfig))

	// The debug group exists only while its toggle is on, so that turning the
	// toggle off removes the surface rather than merely guarding it.
	if cfg.Weak.VerboseErrorsDebugEndpoint {
		// --- DELIBERATE WEAKNESS: WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT ---
		//
		// net/http/pprof is mounted on the same public listener, behind no
		// authentication at all. The heap and goroutine profiles expose memory
		// contents (including whatever was in a request buffer when the profile
		// was taken), /debug/pprof/cmdline exposes the container's arguments,
		// and the profile endpoints are a denial-of-service lever in their own
		// right. It demonstrates why the debug handlers belong on a separate,
		// authenticated listener — never on the application's own.
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// A panic on demand. The recovery middleware below is what renders it,
		// so this exists purely so that the shape of a rendered panic can be
		// demonstrated without waiting for a real bug to trip it.
		mux.HandleFunc("GET /debug/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("debug/panic: deliberately triggered")
		})
	}

	// Anything the table above did not match. Serving 404s through the mux
	// rather than letting it emit its own keeps the response shape identical to
	// every other error the application produces.
	mux.HandleFunc("/", s.handleNotFound)

	// Order matters: recovery is outermost so that a panic inside the logger or
	// inside a handler is caught by the same block.
	return s.recoverPanic(s.logRequests(mux))
}

// handleHealth answers the container healthcheck. It is the only handler that
// touches the database and does not care what it finds: a portal that can reach
// postgres is healthy, whether or not the schema has been created yet, because
// the seed runs after the listener starts.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	code := http.StatusOK
	if err := s.db.PingContext(r.Context()); err != nil {
		status = "database unreachable"
		code = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%s\n", status)
}

// handleNotFound is the catch-all. It answers in JSON for /api/* and in HTML
// everywhere else, because a JSON client that gets an HTML error page has to
// parse the failure twice.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such endpoint"})
		return
	}
	http.Error(w, "404 page not found", http.StatusNotFound)
}

// User is a portal identity as stored in the database.
type User struct {
	ID       int64
	Username string
	Role     string
	FullName string
	Email    string
}

// Identity is the identity behind a request: the session's claims, enriched
// with the database row when one exists.
//
// The Role always comes from the session token and never from the row. That is
// correct in both postures and it is the reason the forged-token weakness is
// worth showing: a signed token is trustworthy because the server signed it,
// and an unsigned token is not trustworthy at all — but in neither case does
// re-reading the role from the database help, because the client is not
// claiming a username the database disagrees with, it is claiming a role.
type Identity struct {
	Username string
	Role     string
	UserID   int64
	FullName string
	Email    string
	Known    bool
}

// IsAdmin reports whether the caller's session carries the admin role.
func (i Identity) IsAdmin() bool { return i.Role == "admin" }

// identity resolves the session cookie into an Identity. It returns false when
// there is no usable session at all.
func (s *Server) identity(r *http.Request) (Identity, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return Identity{}, false
	}

	session, ok := decodeSession(s.cfg, cookie.Value)
	if !ok {
		return Identity{}, false
	}

	identity := Identity{Username: session.Username, Role: session.Role}

	// The row supplies the numeric id used by the ownership checks, and the
	// display fields. A session naming a user who does not exist is not an
	// error: it is a token the portal did not issue, and the endpoint's own
	// authorization check is what decides whether that matters.
	if user, err := s.userByName(r.Context(), session.Username); err == nil {
		identity.UserID = user.ID
		identity.FullName = user.FullName
		identity.Email = user.Email
		identity.Known = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("looking up session user %q: %v", session.Username, err)
	}

	return identity, true
}

// requireSession wraps a handler that needs any authenticated caller.
func (s *Server) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.identity(r); !ok {
			s.unauthorized(w, r)
			return
		}
		next(w, r)
	}
}

// requireAdmin wraps a handler that needs an administrator.
//
// --- DELIBERATE WEAKNESS: WEAK_BROKEN_ADMIN_AUTHZ ---
//
// With the toggle on, the presence of the header "X-Clayface-Role: admin" is
// accepted as proof of the admin role. The header is client-controlled, so any
// unauthenticated caller can reach the whole panel — the user list, the audit
// log, the diagnostics runner and the configuration dump — by adding one
// header. It demonstrates trusting an authorization decision that arrives from
// the request instead of deriving it from an authenticated session.
//
// With the toggle off the header is ignored entirely and the session must carry
// the admin role. Note that the fall-through path is the same code in both
// postures: what the toggle adds is the bypass, not a different check.
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Weak.BrokenAdminAuthz && r.Header.Get("X-Clayface-Role") == "admin" {
			next(w, r)
			return
		}

		identity, ok := s.identity(r)
		if !ok {
			s.unauthorized(w, r)
			return
		}
		if !identity.IsAdmin() {
			s.forbidden(w, r)
			return
		}

		next(w, r)
	}
}

// authenticate resolves a username and password to a user row.
//
// --- DELIBERATE WEAKNESS: WEAK_SQLI_LOGIN ---
//
// With the toggle on the credential lookup is assembled by string
// concatenation, so the username field is a SQL fragment. The classic
// "' OR '1'='1' -- " payload terminates the string literal, adds a condition
// that is always true, and comments out the password comparison that follows,
// so the query returns the first user in the table and the caller is logged in
// as them. It demonstrates why user input must never be concatenated into a
// statement, and why the fix is parameter binding rather than escaping.
//
// With the toggle off the same lookup is bound as $1 and $2. The payload then
// becomes a literal username that matches nothing.
func (s *Server) authenticate(ctx context.Context, username, password string) (User, error) {
	const columns = "id, username, role, full_name, email"
	digest := sha256Hex(password)

	if s.cfg.Weak.SQLILogin {
		query := "SELECT " + columns + " FROM users WHERE username = '" + username +
			"' AND password_hash = '" + digest + "' LIMIT 1"

		user, err := scanUser(s.db.QueryRowContext(ctx, query))
		if err != nil {
			// The query text travels with the error so that the verbose-errors
			// toggle has something worth printing. It only ever reaches a client
			// in that posture.
			if errors.Is(err, sql.ErrNoRows) {
				return User{}, err
			}
			return User{}, fmt.Errorf("login lookup failed: %w (query: %s)", err, query)
		}
		return user, nil
	}

	query := "SELECT " + columns + " FROM users WHERE username = $1 AND password_hash = $2 LIMIT 1"
	user, err := scanUser(s.db.QueryRowContext(ctx, query, username, digest))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, err
		}
		return User{}, fmt.Errorf("login lookup failed: %w", err)
	}
	return user, nil
}

// userByName loads a user row by username, used to turn a session into an
// Identity.
func (s *Server) userByName(ctx context.Context, username string) (User, error) {
	const query = "SELECT id, username, role, full_name, email FROM users WHERE username = $1"
	return scanUser(s.db.QueryRowContext(ctx, query, username))
}

// scanUser reads one user row. Returning the error rather than logging it lets
// the caller tell "no such user" apart from "the database is gone", which is
// the difference between a 401 and a 500.
func scanUser(row *sql.Row) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Role, &user.FullName, &user.Email)
	return user, err
}

// sha256Hex is the digest the users table stores. It is a bare, unsalted,
// uniterated hash on purpose: it is the baseline the login weakness lands on,
// and it is what makes a dumped users table crackable in seconds.
func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// clientIP is the address the audit log records for a request.
//
// --- DELIBERATE WEAKNESS: WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT ---
//
// With the toggle on, the leftmost X-Forwarded-For value wins when the header
// is present. That header is supplied by the client, so the address recorded
// against a login attempt — or against anything else that gets audited — is
// whatever the attacker chose to write. It demonstrates why a forwarded-for
// header is only trustworthy from a proxy you control, and why the leftmost
// entry in particular is not the client: it is the entry any client can add.
//
// With the toggle off the address comes from the connection, which nginx
// terminates, so it is the proxy's address and cannot be forged from outside.
func (s *Server) clientIP(r *http.Request) string {
	if s.cfg.Weak.VerboseErrorsDebugEndpoint {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			return strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// audit writes one row to the audit log. Failures are logged and swallowed: an
// audit write must never be the reason a request fails, or an attacker could
// take the portal down by making the log table unwritable.
func (s *Server) audit(r *http.Request, username, action, detail string) {
	// context.WithoutCancel so that the insert still lands when the client has
	// already hung up. An attacker who can suppress the audit entry for a
	// failed login by aborting the request has removed the detection.
	ctx := context.WithoutCancel(r.Context())
	const query = "INSERT INTO audit_log (username, action, detail, client_ip) VALUES ($1, $2, $3, $4)"
	if _, err := s.db.ExecContext(ctx, query, username, action, detail, s.clientIP(r)); err != nil {
		log.Printf("audit write failed (action=%s user=%s): %v", action, username, err)
	}
}

// errorDetail renders an error for the client.
//
// --- DELIBERATE WEAKNESS: WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT ---
//
// With the toggle on the caller gets the underlying error verbatim, which for a
// database failure includes the driver's message and — because the vulnerable
// handlers wrap their SQL into the error — the statement itself. Stack traces
// from panics and the effective configuration are rendered by recoverPanic
// below, and the pprof handlers are mounted by NewServer for the same toggle.
// It demonstrates how much an attacker learns from an error page that was
// meant for a developer.
//
// With the toggle off every failure reads the same, and the underlying error
// goes to the log where it belongs.
func (s *Server) errorDetail(err error) string {
	if s.cfg.Weak.VerboseErrorsDebugEndpoint && err != nil {
		return err.Error()
	}
	return "internal server error"
}

// unauthorized answers a request that needs a session it does not have. HTML
// clients get redirected to the login form; API clients get a 401, because
// following a redirect would silently turn a failed API call into a successful
// HTML one.
func (s *Server) unauthorized(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	http.Redirect(w, r, "/portal/login", http.StatusFound)
}

// forbidden answers a request whose caller is authenticated but not allowed.
func (s *Server) forbidden(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.writeJSON(w, http.StatusForbidden, map[string]string{"error": "not permitted"})
		return
	}
	http.Error(w, "403 forbidden", http.StatusForbidden)
}

// writeJSON writes a JSON response. The encoding error is logged rather than
// returned: by the time it happens the header is already on the wire, so there
// is nothing useful left to say to the client.
func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("writing json response: %v", err)
	}
}

// pageCommon is the part of every page's data that the shared layout draws.
type pageCommon struct {
	Title    string
	Notice   string
	Error    string
	Identity *Identity
}

// layout is the page shell. It is small on purpose: this application exists to
// be attacked, and a large front end would only be more code between a reader
// and the weakness they came to look at.
const layout = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Clayface Portal — {{.Title}}</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: system-ui, sans-serif; margin: 0; line-height: 1.5; }
  header, footer { padding: 0.75rem 1.25rem; background: #1f2937; color: #f9fafb; }
  header a { color: #f9fafb; margin-right: 1rem; }
  footer { font-size: 0.8rem; color: #d1d5db; }
  main { padding: 1.25rem; max-width: 60rem; }
  table { border-collapse: collapse; width: 100%; margin: 0.5rem 0 1.5rem; }
  th, td { border: 1px solid #9ca3af; padding: 0.35rem 0.5rem; text-align: left;
           vertical-align: top; font-size: 0.9rem; }
  th { background: #e5e7eb; color: #111827; }
  .error { background: #fee2e2; border: 1px solid #b91c1c; padding: 0.5rem 0.75rem; }
  .notice { background: #dcfce7; border: 1px solid #15803d; padding: 0.5rem 0.75rem; }
  pre { background: #111827; color: #e5e7eb; padding: 0.75rem; overflow-x: auto; }
  code { font-size: 0.9rem; }
</style>
</head>
<body>
<header>
  <strong>Clayface Portal</strong>
  <a href="/portal/">Home</a>
  <a href="/portal/search">Search</a>
  <a href="/portal/profile">Profile</a>
  {{if .Identity}}<a href="/admin/">Admin</a> <a href="/portal/logout">Sign out</a>
  {{else}}<a href="/portal/login">Sign in</a>{{end}}
</header>
<main>
{{if .Error}}<p class="error">{{.Error}}</p>{{end}}
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}
{{template "content" .}}
</main>
<footer>Clayface Portal — deliberately vulnerable lab application. Internal use only.</footer>
</body>
</html>
`

// newPage composes the shared shell with one page's content template. Both are
// parsed from string constants, so a malformed template is a start-up failure
// rather than a failure on the first request that happens to use it.
func newPage(name, content string) *template.Template {
	return template.Must(template.Must(template.New(name).Parse(layout)).Parse(content))
}

// renderPage executes a page template. A failure here is a bug in the template
// rather than anything the caller did, so it is logged and the client gets a
// plain error: there is no useful way for a half-written page to recover.
func (s *Server) renderPage(w http.ResponseWriter, tmpl *template.Template, data any) {
	s.renderPageStatus(w, tmpl, http.StatusOK, data)
}

// renderPageStatus is renderPage for the handlers that need a status other than
// 200 — a rejected login, mostly, where the body is still the form.
func (s *Server) renderPageStatus(w http.ResponseWriter, tmpl *template.Template, status int, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("rendering page %s: %v", tmpl.Name(), err)
	}
}

// recoverPanic turns a panic into a response.
//
// --- DELIBERATE WEAKNESS: WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT ---
//
// With the toggle on, the response body carries the panic value, the full
// goroutine stack trace, and the effective configuration — database host, user
// and password included. It demonstrates that a stack trace is an information
// disclosure: it names internal types, file paths and library versions, and
// here it also hands over the credentials the process is running with.
//
// With the toggle off the panic is logged and the client gets a bare 500.
func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			log.Printf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, recovered, debug.Stack())

			if s.cfg.Weak.VerboseErrorsDebugEndpoint {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "panic: %v\n\n%s\n--- effective configuration ---\n", recovered, debug.Stack())
				for key, value := range s.cfg.Settings() {
					fmt.Fprintf(w, "%s=%s\n", key, value)
				}
				fmt.Fprint(w, "\n--- weakness toggles ---\n")
				for key, value := range s.cfg.WeaknessReport() {
					fmt.Fprintf(w, "%s=%t\n", key, value)
				}
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// logRequests writes one line per request to the container log. It exists so
// that the lab has a second, independent record of what was done to it — the
// audit table is the application's own view, and an attacker who can write to
// the database can edit it.
func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s", r.Method, r.URL.RequestURI(), s.clientIP(r))
	})
}
