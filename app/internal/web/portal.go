package web

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clayface/app/internal/config"
)

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

const portalIndexContent = `{{define "content"}}
<h1>Customer portal</h1>
<p>Order tracking, invoicing and document exchange for Clayface customers.</p>

<h2>Document library</h2>
<table>
  <tr><th>Title</th><th>Classification</th><th>File</th><th></th></tr>
  {{range .Documents}}
  <tr>
    <td>{{.Title}}</td>
    <td>{{.Classification}}</td>
    <td><code>{{.Filename}}</code></td>
    <td><a href="/portal/download?file={{.Filename}}">download</a></td>
  </tr>
  {{else}}
  <tr><td colspan="4">No documents.</td></tr>
  {{end}}
</table>
{{end}}`

const loginContent = `{{define "content"}}
<h1>Sign in</h1>
<form method="post" action="/portal/login">
  <p><label>Username<br><input name="username" autocomplete="username"></label></p>
  <p><label>Password<br><input name="password" type="password" autocomplete="current-password"></label></p>
  <p><button type="submit">Sign in</button></p>
</form>
<p><a href="/portal/">Back to the portal</a></p>
{{end}}`

const profileContent = `{{define "content"}}
<h1>Your profile</h1>
<table>
  <tr><th>Username</th><td>{{.Identity.Username}}</td></tr>
  <tr><th>Full name</th><td>{{.Identity.FullName}}</td></tr>
  <tr><th>Email</th><td>{{.Identity.Email}}</td></tr>
  <tr><th>Role</th><td>{{.Identity.Role}}</td></tr>
  <tr><th>Session issued</th><td>{{.Issued}}</td></tr>
</table>

<h2>Your orders</h2>
<table>
  <tr><th>Ref</th><th>Description</th><th>Amount</th><th>Status</th></tr>
  {{range .Orders}}
  <tr>
    <td>{{.OrderRef}}</td>
    <td>{{.Description}}</td>
    <td>{{.Amount}}</td>
    <td>{{.Status}}</td>
  </tr>
  {{else}}
  <tr><td colspan="4">No orders on this account.</td></tr>
  {{end}}
</table>
<p><a href="/portal/search">Search the document library</a></p>
{{end}}`

const searchContent = `{{define "content"}}
<h1>Search</h1>
<form method="get" action="/portal/search">
  <p><label>Query <input name="q" value="{{.Query}}" size="40"></label>
     <button type="submit">Search</button></p>
</form>
{{if .Searched}}
<h2>Results for <code>{{.Query}}</code></h2>
<table>
  <tr><th>ID</th><th>Title</th><th>Body</th></tr>
  {{range .Hits}}
  <tr><td>{{.ID}}</td><td>{{.Title}}</td><td>{{.Body}}</td></tr>
  {{else}}
  <tr><td colspan="3">Nothing matched.</td></tr>
  {{end}}
</table>
{{end}}
{{end}}`

var (
	portalIndexPage = newPage("portal-index", portalIndexContent)
	loginPage       = newPage("portal-login", loginContent)
	profilePage     = newPage("portal-profile", profileContent)
	searchPage      = newPage("portal-search", searchContent)
)

// ---------------------------------------------------------------------------
// View models
// ---------------------------------------------------------------------------

type documentRow struct {
	ID             int64
	Title          string
	Filename       string
	Classification string
	CreatedAt      time.Time
}

type orderRow struct {
	ID          int64
	OrderRef    string
	Description string
	Amount      string
	Status      string
}

type searchHit struct {
	ID    int64
	Title string
	Body  string
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handlePortalIndex is the portal's front page: the document library, plus a
// link to sign in. It is unauthenticated, which is how a visitor discovers the
// filenames the download endpoint takes.
func (s *Server) handlePortalIndex(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	documents, err := s.listDocuments(r.Context())
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}

	data := struct {
		pageCommon
		Documents []documentRow
	}{
		pageCommon: pageCommon{Title: "Home", Identity: identityOrNil(identity)},
		Documents:  documents,
	}
	s.renderPage(w, portalIndexPage, data)
}

// handlePortalLoginForm renders the sign-in form.
func (s *Server) handlePortalLoginForm(w http.ResponseWriter, r *http.Request) {
	identity, ok := s.identity(r)

	data := struct {
		pageCommon
	}{}
	if ok {
		data.Notice = "Already signed in as " + identity.Username + "."
	}
	data.Title = "Sign in"
	s.renderPage(w, loginPage, data)
}

// handlePortalLogin checks the submitted credentials.
//
// The two postures meet in Server.authenticate; what happens here is the same
// either way, which is what makes the toggle a property of the lookup rather
// than of the handler.
func (s *Server) handlePortalLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := s.authenticate(r.Context(), username, password)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		s.audit(r, username, "login.failed", "invalid credentials")
		s.renderLoginFailure(w, "Incorrect username or password.")
		return
	case err != nil:
		// A database failure is not a credential failure, and saying so out
		// loud is the verbose-errors weakness. The audit entry records that
		// something failed without duplicating the error into the table.
		s.audit(r, username, "login.failed", "credential lookup error")
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, s.cfg, newSession(user.Username, user.Role))
	s.audit(r, user.Username, "login.success", "portal session opened")
	http.Redirect(w, r, "/portal/profile", http.StatusSeeOther)
}

// renderLoginFailure redraws the form with a message. The status is 401 so that
// a scripted client can tell a rejected login from a served page.
func (s *Server) renderLoginFailure(w http.ResponseWriter, message string) {
	data := struct {
		pageCommon
	}{pageCommon: pageCommon{Title: "Sign in", Error: message}}
	s.renderPageStatus(w, loginPage, http.StatusUnauthorized, data)
}

// handlePortalLogout clears the cookie. There is no server-side session to
// destroy, so this is the whole of signing out.
func (s *Server) handlePortalLogout(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)
	if identity.Username != "" {
		s.audit(r, identity.Username, "logout", "session cookie cleared")
	}
	clearSessionCookie(w, s.cfg)
	http.Redirect(w, r, "/portal/", http.StatusSeeOther)
}

// handlePortalProfile shows the signed-in user and the orders on their
// account. It is wrapped in requireSession by the routing table.
func (s *Server) handlePortalProfile(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	orders, err := s.ordersForUser(r.Context(), identity.UserID)
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}

	data := struct {
		pageCommon
		Issued string
		Orders []orderRow
	}{
		pageCommon: pageCommon{Title: "Profile", Identity: identityOrNil(identity)},
		Issued:     identityIssued(r, s.cfg),
		Orders:     orders,
	}
	s.renderPage(w, profilePage, data)
}

// handlePortalSearch searches the document register.
func (s *Server) handlePortalSearch(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")

	var hits []searchHit
	if term != "" {
		var err error
		hits, err = s.searchDocuments(r.Context(), term)
		if err != nil {
			s.audit(r, callerName(r, s), "document.search.error", term)
			http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
			return
		}
		s.audit(r, callerName(r, s), "document.search", term)
	}

	data := struct {
		pageCommon
		Query    string
		Searched bool
		Hits     []searchHit
	}{
		pageCommon: pageCommon{Title: "Search", Identity: identityFromRequest(r, s)},
		Query:      term,
		Searched:   term != "",
		Hits:       hits,
	}
	s.renderPage(w, searchPage, data)
}

// handlePortalDownload serves a document from the on-disk store.
//
// --- DELIBERATE WEAKNESS: WEAK_PATH_TRAVERSAL_DOWNLOAD ---
//
// With the toggle on, the requested name is joined straight onto the documents
// directory with no cleaning and no check, so "../" sequences walk out of it.
// "?file=../../etc/passwd" reads the container's password file, and a longer
// chain reaches anything else the process can read — including the application
// configuration and the postgres client credentials if they are on disk. It
// demonstrates why a path assembled from user input must be constrained to the
// directory it is supposed to be in, and why the check has to happen after the
// path is resolved rather than before.
//
// Note that this is not the same as an unrestricted file read: the endpoint is
// still relative to the process's own filesystem, which is the container's.
// The escape is out of the documents directory, not out of the container.
//
// With the toggle off any name containing a separator or ".." is rejected
// outright, and the resolved path is then confirmed to still be inside the
// documents directory. Both checks are kept because either one alone would be
// a single point of failure.
func (s *Server) handlePortalDownload(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("file")
	if name == "" {
		http.Error(w, "the file parameter is required", http.StatusBadRequest)
		return
	}

	documentsDir := s.cfg.DocumentsDir()
	var path string

	if s.cfg.Weak.PathTraversalDownload {
		path = filepath.Join(documentsDir, name)
	} else {
		if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
			s.audit(r, callerName(r, s), "document.download.rejected", name)
			http.Error(w, "invalid file name", http.StatusBadRequest)
			return
		}

		path = filepath.Join(documentsDir, name)
		relative, err := filepath.Rel(documentsDir, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			s.audit(r, callerName(r, s), "document.download.rejected", name)
			http.Error(w, "invalid file name", http.StatusBadRequest)
			return
		}
	}

	file, err := os.Open(path)
	if err != nil {
		s.audit(r, callerName(r, s), "document.download.missing", name)
		http.Error(w, s.errorDetail(err), http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.Error(w, "not a file", http.StatusNotFound)
		return
	}

	s.audit(r, callerName(r, s), "document.download", name)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%q", filepath.Base(name)))
	http.ServeContent(w, r, filepath.Base(name), info.ModTime(), file)
}

// ---------------------------------------------------------------------------
// Queries shared by the portal and the API
// ---------------------------------------------------------------------------

// listDocuments reads the document register.
func (s *Server) listDocuments(ctx context.Context) ([]documentRow, error) {
	const query = `SELECT id, title, filename, classification, created_at
	               FROM documents ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing documents: %w", err)
	}
	defer rows.Close()

	var documents []documentRow
	for rows.Next() {
		var doc documentRow
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Filename, &doc.Classification, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("reading document row: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, rows.Err()
}

// searchDocuments searches the document register by title.
//
// --- DELIBERATE WEAKNESS: WEAK_SQLI_SEARCH ---
//
// With the toggle on the search term is concatenated into the LIKE pattern, so
// the term is SQL. A UNION SELECT appends rows of the attacker's choosing to
// the result: "zz%' UNION SELECT id, service_account, api_key FROM api_keys -- "
// returns one row per credential in the api_keys table, because that table has
// three columns of compatible types. It demonstrates injection through a
// read-only endpoint — no login, no write, and the entire credential table
// comes out — and why "it is only a search box" is not a security argument.
//
// With the toggle off the term is bound as a parameter, so the same payload is
// searched for as a literal string and matches nothing.
func (s *Server) searchDocuments(ctx context.Context, term string) ([]searchHit, error) {
	const columns = "id, title, body"

	var (
		query string
		args  []any
	)

	if s.cfg.Weak.SQLISearch {
		query = "SELECT " + columns + " FROM documents WHERE title LIKE '%" + term + "%' ORDER BY id"
	} else {
		query = "SELECT " + columns + " FROM documents WHERE title LIKE '%' || $1 || '%' ORDER BY id"
		args = []any{term}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		// The statement travels with the error so that a student in the
		// verbose-errors posture sees the SQL they just broke, which is how the
		// injection is normally confirmed by hand.
		return nil, fmt.Errorf("document search failed: %w (query: %s)", err, query)
	}
	defer rows.Close()

	var hits []searchHit
	for rows.Next() {
		var hit searchHit
		if err := rows.Scan(&hit.ID, &hit.Title, &hit.Body); err != nil {
			return nil, fmt.Errorf("reading search row: %w", err)
		}
		hits = append(hits, hit)
	}

	return hits, rows.Err()
}

// ordersForUser reads the orders belonging to one user's customers. This is the
// query the IDOR weakness fails to apply to a single row.
func (s *Server) ordersForUser(ctx context.Context, userID int64) ([]orderRow, error) {
	const query = `SELECT o.id, o.order_ref, o.description, o.status, o.amount_cents, o.currency
	               FROM orders o
	               JOIN customers c ON c.id = o.customer_id
	               WHERE c.user_id = $1
	               ORDER BY o.id`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}
	defer rows.Close()

	var orders []orderRow
	for rows.Next() {
		var (
			order    orderRow
			cents    int64
			currency string
		)
		if err := rows.Scan(&order.ID, &order.OrderRef, &order.Description, &order.Status, &cents, &currency); err != nil {
			return nil, fmt.Errorf("reading order row: %w", err)
		}
		order.Amount = formatAmount(cents, currency)
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

// formatAmount renders a minor-unit amount for display. Money is stored in
// cents so that nothing in this application has to do floating point.
func formatAmount(cents int64, currency string) string {
	return fmt.Sprintf("%s %d.%02d", currency, cents/100, cents%100)
}

// identityOrNil converts the "no session" case into a nil pointer, which is
// what the shared layout tests for when it decides whether to draw the sign-out
// link.
func identityOrNil(identity Identity) *Identity {
	if identity.Username == "" {
		return nil
	}
	return &identity
}

// identityFromRequest is identityOrNil for handlers that only need to draw the
// header.
func identityFromRequest(r *http.Request, s *Server) *Identity {
	identity, _ := s.identity(r)
	return identityOrNil(identity)
}

// callerName is the username to attribute an action to, falling back to a
// marker for the unauthenticated case so that the audit log distinguishes "a
// signed-in user did this" from "somebody with no session did this".
func callerName(r *http.Request, s *Server) string {
	if identity, ok := s.identity(r); ok {
		return identity.Username
	}
	return "anonymous"
}

// identityIssued renders the session's issue time for the profile page. It is
// re-decoded from the cookie rather than carried on Identity, because only the
// session-facing code should care that the token has a timestamp at all.
func identityIssued(r *http.Request, cfg *config.Config) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "unknown"
	}
	session, ok := decodeSession(cfg, cookie.Value)
	if !ok {
		return "unknown"
	}
	return time.Unix(session.IssuedAt, 0).UTC().Format(time.RFC3339)
}
