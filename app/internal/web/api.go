package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// View models
// ---------------------------------------------------------------------------

// orderDetail is one order as the API returns it. The account owner travels
// with the row on purpose: it is how a client can tell that the row it just
// fetched belongs to somebody else.
type orderDetail struct {
	ID           int64  `json:"id"`
	OrderRef     string `json:"order_ref"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	AmountCents  int64  `json:"amount_cents"`
	Currency     string `json:"currency"`
	CustomerID   int64  `json:"customer_id"`
	Customer     string `json:"customer"`
	AccountOwner string `json:"account_owner"`
}

// invoiceDetail is one invoice as the API returns it.
type invoiceDetail struct {
	ID           int64  `json:"id"`
	InvoiceRef   string `json:"invoice_ref"`
	OrderID      int64  `json:"order_id"`
	AmountCents  int64  `json:"amount_cents"`
	Currency     string `json:"currency"`
	IssuedOn     string `json:"issued_on"`
	DueOn        string `json:"due_on"`
	Paid         bool   `json:"paid"`
	Notes        string `json:"notes"`
	AccountOwner string `json:"account_owner"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleAPILogin authenticates and returns the session token in the response
// body as well as in a cookie.
//
// Returning the token in the body is what makes the forgeable-token weakness
// easy to demonstrate: a client can take the token it was issued, decode it,
// change the role field, and present the result. That is a two-line exercise
// rather than a cookie-editor exercise.
func (s *Server) handleAPILogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		// Bounded read: this endpoint is unauthenticated, so the body is
		// attacker-controlled and an unbounded decode would be a memory
		// exhaustion lever.
		body := io.LimitReader(r.Body, 64*1024)
		if err := json.NewDecoder(body).Decode(&credentials); err != nil {
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed JSON body"})
			return
		}
	} else {
		credentials.Username = r.FormValue("username")
		credentials.Password = r.FormValue("password")
	}

	user, err := s.authenticate(r.Context(), credentials.Username, credentials.Password)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		s.audit(r, credentials.Username, "api.login.failed", "invalid credentials")
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	case err != nil:
		s.audit(r, credentials.Username, "api.login.failed", "credential lookup error")
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	session := newSession(user.Username, user.Role)
	setSessionCookie(w, s.cfg, session)
	s.audit(r, user.Username, "api.login.success", "api session opened")

	s.writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"session":       encodeSession(s.cfg, session),
		"user": map[string]any{
			"username":  user.Username,
			"role":      user.Role,
			"full_name": user.FullName,
			"email":     user.Email,
		},
	})
}

// handleAPIProfile returns the caller's own identity.
func (s *Server) handleAPIProfile(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)
	s.writeJSON(w, http.StatusOK, map[string]any{
		"username":  identity.Username,
		"role":      identity.Role,
		"full_name": identity.FullName,
		"email":     identity.Email,
	})
}

// handleAPIOrders lists the orders on the caller's own account. Unlike the
// single-order endpoint below, this one is scoped by the session in both
// postures — which is exactly what makes the single-order weakness visible when
// the two are compared.
func (s *Server) handleAPIOrders(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	orders, err := s.ordersForUser(r.Context(), identity.UserID)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

// handleAPIOrder returns one order by id.
//
// --- DELIBERATE WEAKNESS: WEAK_IDOR_DOCUMENTS ---
//
// With the toggle on, the row is selected by primary key alone: no comparison
// against the caller's own account, and no session required either. Any caller
// who can count from one to the number of orders in the database can read the
// whole book — customer names, order values, statuses — one incrementing id at
// a time. It demonstrates an authorization check that was never written, which
// is the most common form of this bug: the query is correct, it just answers a
// question nobody should have been allowed to ask.
//
// With the toggle off the same query carries an ownership predicate that joins
// through customers to the session's user id, and the endpoint requires a
// session at all. A row belonging to somebody else is a 403; a row that does
// not exist is a 404.
func (s *Server) handleAPIOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	identity, hasSession := s.identity(r)
	if !s.cfg.Weak.IDORDocuments && !hasSession {
		s.unauthorized(w, r)
		return
	}

	order, err := s.fetchOrder(r.Context(), id, identity.UserID, !s.cfg.Weak.IDORDocuments)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such order"})
		return
	case err != nil:
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	s.audit(r, identity.Username, "api.order.view", fmt.Sprintf("order %d (%s)", order.ID, order.AccountOwner))
	s.writeJSON(w, http.StatusOK, order)
}

// handleAPIInvoice returns one invoice by id, with the same weakness and the
// same fix as handleAPIOrder. The two endpoints are separate because the lab
// uses them to show that the bug repeats wherever a row is fetched by id.
func (s *Server) handleAPIInvoice(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	identity, hasSession := s.identity(r)
	if !s.cfg.Weak.IDORDocuments && !hasSession {
		s.unauthorized(w, r)
		return
	}

	invoice, err := s.fetchInvoice(r.Context(), id, identity.UserID, !s.cfg.Weak.IDORDocuments)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such invoice"})
		return
	case err != nil:
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	s.audit(r, identity.Username, "api.invoice.view", fmt.Sprintf("invoice %d (%s)", invoice.ID, invoice.AccountOwner))
	s.writeJSON(w, http.StatusOK, invoice)
}

// handleAPISearch is the JSON form of the portal search. It shares the query
// builder, so it carries the same title. It is grouped under the API because
// that is how the customer's own integrations reach it.
func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")
	if term == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the q parameter is required"})
		return
	}

	hits, err := s.searchDocuments(r.Context(), term)
	if err != nil {
		s.audit(r, callerName(r, s), "api.search.error", term)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	// A nil slice marshals to "null"; an empty one marshals to "[]", which is
	// what a client that iterates the result wants.
	if hits == nil {
		hits = []searchHit{}
	}

	s.audit(r, callerName(r, s), "api.search", term)
	s.writeJSON(w, http.StatusOK, map[string]any{"query": term, "count": len(hits), "results": hits})
}

// handleAPIKeys lists the API keys held by the internal services.
//
// The endpoint is not itself a weakness: it requires a session, and it is
// exactly the kind of listing a service inventory would have. What makes it
// worth reading is the data behind it — see the
// WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB toggle, which decides whether one of
// these rows holds a usable password or a placeholder.
func (s *Server) handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	const query = `SELECT id, key_name, service_account, credential_type, api_key, notes, created_at
	               FROM api_keys ORDER BY id`

	rows, err := s.db.QueryContext(r.Context(), query)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}
	defer rows.Close()

	type apiKey struct {
		ID             int64  `json:"id"`
		KeyName        string `json:"key_name"`
		ServiceAccount string `json:"service_account"`
		CredentialType string `json:"credential_type"`
		APIKey         string `json:"api_key"`
		Notes          string `json:"notes"`
		CreatedAt      string `json:"created_at"`
	}

	keys := []apiKey{}
	for rows.Next() {
		var (
			key       apiKey
			createdAt time.Time
		)
		if err := rows.Scan(&key.ID, &key.KeyName, &key.ServiceAccount, &key.CredentialType,
			&key.APIKey, &key.Notes, &createdAt); err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
			return
		}
		key.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": s.errorDetail(err)})
		return
	}

	identity, _ := s.identity(r)
	s.audit(r, identity.Username, "api.keys.list", fmt.Sprintf("%d rows", len(keys)))
	s.writeJSON(w, http.StatusOK, map[string]any{"keys": keys})
}

// handleAPIFetch retrieves a URL server-side and returns the body — the
// feature the customer integrations use to pull a supplier's price feed.
//
// --- DELIBERATE WEAKNESS: WEAK_SSRF_URL_FETCH ---
//
// With the toggle on, any URL the caller supplies is fetched from inside the
// container network. That is server-side request forgery: the caller cannot
// reach the internal services directly, but the portal can, so the portal
// becomes the proxy. Reaching http://postgres:5432/ or a cloud metadata
// address at 169.254.169.254/ are the usual first moves, and the response body
// is returned verbatim, so a service that answers with data answers the
// attacker. It demonstrates why an outbound fetch built from user input needs
// an allow-list, not a deny-list.
//
// http://10.0.0.1/ is NOT one of those moves, and this container is why. The
// portal runs on app01 in the DMZ, and 10.0.0.1 is the edge router's address
// on the LAN leg — a different interface, so the fetch does not route there at
// all. The reachable management address from here is 10.0.10.1, and the DMZ
// ruleset (networks.dmz.allow in lab.yaml) deliberately does not open :443 on
// it: the segment gets LDAP, DNS and DHCP and nothing else. So a caller who
// swaps the URL for either address gets a timeout, not an admin console, and
// that refusal is the boundary working rather than the weakness failing.
//
// With the toggle off the scheme must be https and the host must not be a
// loopback, private, link-local or unspecified address literal, and redirects
// are re-checked against the same rules so that a permitted host cannot bounce
// the request to a forbidden one.
//
// The hardened form is a real check but not a complete one, and it is worth
// saying so: a hostname is validated as a literal, so a name that resolves into
// the private range still passes. Closing that needs resolution-time checks or
// an egress policy, which is a network control rather than an application one.
func (s *Server) handleAPIFetch(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the url parameter is required"})
		return
	}

	if !s.cfg.Weak.SSRFURLFetch {
		if err := validateFetchTarget(target); err != nil {
			s.audit(r, callerName(r, s), "api.fetch.rejected", target)
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	if !s.cfg.Weak.SSRFURLFetch {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("too many redirects")
			}
			return validateFetchTarget(req.URL.String())
		}
	}

	response, err := client.Get(target)
	if err != nil {
		s.audit(r, callerName(r, s), "api.fetch.error", target)
		s.writeJSON(w, http.StatusBadGateway, map[string]string{"error": s.errorDetail(err)})
		return
	}
	defer response.Body.Close()

	// Bounded at one megabyte: the point of the exercise is to read what an
	// internal service answers with, not to stream a large body through the
	// portal.
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		s.writeJSON(w, http.StatusBadGateway, map[string]string{"error": s.errorDetail(err)})
		return
	}

	s.audit(r, callerName(r, s), "api.fetch", fmt.Sprintf("%s -> %s", target, response.Status))
	s.writeJSON(w, http.StatusOK, map[string]any{
		"url":          target,
		"status":       response.StatusCode,
		"content_type": response.Header.Get("Content-Type"),
		"body":         string(body),
	})
}

// handleAPIDebugConfig returns the effective configuration as JSON.
//
// --- DELIBERATE WEAKNESS: WEAK_HARDCODED_DB_CREDENTIALS ---
//
// The redaction is decided inside config.Settings, which is the single place
// both this endpoint and /admin/config read from, so the two cannot disagree
// about whether the password is being printed.
//
// The endpoint also reports the weakness toggles and the toolchain version.
// Neither is a secret, and both are useful in the lab: the toggle report is how
// a student confirms which posture the deployment is in, and the toolchain
// version is the sort of detail a debug endpoint leaks in the real world and
// that a version-to-CVE lookup turns into an exploit.
func (s *Server) handleAPIDebugConfig(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"config":    s.cfg.Settings(),
		"weakness":  s.cfg.WeaknessReport(),
		"goVersion": runtime.Version(),
	})
}

// ---------------------------------------------------------------------------
// Queries and helpers
// ---------------------------------------------------------------------------

// fetchOrder reads one order. When enforceOwnership is false the row is
// selected by primary key alone, which is the IDOR weakness; when it is true
// the query also requires the order's customer to belong to the session's user.
//
// The order id is bound as a parameter in both postures on purpose. The
// injection toggles and this one are meant to be separable: an id that cannot
// inject keeps this endpoint's finding a pure authorization bug, which is what
// the validation playbook asserts.
func (s *Server) fetchOrder(ctx context.Context, id, userID int64, enforceOwnership bool) (orderDetail, error) {
	query := `SELECT o.id, o.order_ref, o.description, o.status, o.amount_cents, o.currency,
	                 c.id, c.company, u.username
	          FROM orders o
	          JOIN customers c ON c.id = o.customer_id
	          JOIN users u ON u.id = c.user_id
	          WHERE o.id = $1`
	args := []any{id}

	if enforceOwnership {
		query += " AND c.user_id = $2"
		args = append(args, userID)
	}

	var order orderDetail
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&order.ID, &order.OrderRef, &order.Description, &order.Status,
		&order.AmountCents, &order.Currency, &order.CustomerID, &order.Customer, &order.AccountOwner)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return orderDetail{}, err
		}
		return orderDetail{}, fmt.Errorf("fetching order %d: %w", id, err)
	}

	return order, nil
}

// fetchInvoice reads one invoice, with the same ownership rule as fetchOrder.
func (s *Server) fetchInvoice(ctx context.Context, id, userID int64, enforceOwnership bool) (invoiceDetail, error) {
	query := `SELECT i.id, i.invoice_ref, i.order_id, i.amount_cents, i.currency,
	                 i.issued_on, i.due_on, i.paid, i.notes, u.username
	          FROM invoices i
	          JOIN orders o ON o.id = i.order_id
	          JOIN customers c ON c.id = o.customer_id
	          JOIN users u ON u.id = c.user_id
	          WHERE i.id = $1`
	args := []any{id}

	if enforceOwnership {
		query += " AND c.user_id = $2"
		args = append(args, userID)
	}

	var (
		invoice  invoiceDetail
		issuedOn time.Time
		dueOn    time.Time
	)
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&invoice.ID, &invoice.InvoiceRef, &invoice.OrderID, &invoice.AmountCents, &invoice.Currency,
		&issuedOn, &dueOn, &invoice.Paid, &invoice.Notes, &invoice.AccountOwner)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return invoiceDetail{}, err
		}
		return invoiceDetail{}, fmt.Errorf("fetching invoice %d: %w", id, err)
	}

	invoice.IssuedOn = issuedOn.Format("2006-01-02")
	invoice.DueOn = dueOn.Format("2006-01-02")

	return invoice, nil
}

// pathID reads and validates the {id} path segment, answering the request
// itself if it is not a number. Returning a bool keeps the call sites from
// having to repeat the error response.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "the id must be a positive integer", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// validateFetchTarget is the hardened half of the SSRF toggle. It rejects
// anything that is not https and any host that is an address literal in a range
// the container can reach but the caller cannot.
func validateFetchTarget(target string) error {
	parsed, err := url.Parse(target)
	if err != nil {
		return errors.New("unparseable url")
	}

	if parsed.Scheme != "https" {
		return errors.New("only https targets are allowed")
	}

	host := parsed.Hostname()
	if host == "" {
		return errors.New("no host in url")
	}

	// A name, rather than an address, can still resolve into the private range.
	// Blocking the obvious names is worth doing and is not sufficient; see the
	// note on this toggle in the README.
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return errors.New("target is not routable")
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return errors.New("target is not routable")
		}
	}

	return nil
}
