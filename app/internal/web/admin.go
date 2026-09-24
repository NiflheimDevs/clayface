package web

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"time"
)

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

const adminIndexContent = `{{define "content"}}
<h1>Administration</h1>
<p>Internal panel. Not linked from the customer-facing portal navigation.</p>

<h2>Row counts</h2>
<table>
  <tr><th>Table</th><th>Rows</th></tr>
  {{range .Counts}}
  <tr><td><code>{{.Table}}</code></td><td>{{.Rows}}</td></tr>
  {{end}}
</table>

<h2>Pages</h2>
<ul>
  <li><a href="/admin/users">Users</a></li>
  <li><a href="/admin/audit">Audit log</a></li>
  <li><a href="/admin/diagnostics">Diagnostics</a></li>
  <li><a href="/admin/config">Configuration</a></li>
</ul>

<h2>Weakness toggles</h2>
<table>
  <tr><th>Toggle</th><th>State</th></tr>
  {{range .Weaknesses}}
  <tr><td><code>{{.Name}}</code></td><td>{{if .Enabled}}vulnerable{{else}}hardened{{end}}</td></tr>
  {{end}}
</table>
{{end}}`

const adminUsersContent = `{{define "content"}}
<h1>Users</h1>
<table>
  <tr><th>ID</th><th>Username</th><th>Role</th><th>Full name</th><th>Email</th><th>Department</th></tr>
  {{range .Users}}
  <tr>
    <td>{{.ID}}</td>
    <td>{{.Username}}</td>
    <td>{{.Role}}</td>
    <td>{{.FullName}}</td>
    <td>{{.Email}}</td>
    <td>{{.Department}}</td>
  </tr>
  {{end}}
</table>
<p>Credentials are not shown here. The application service account's credential
is held in <code>api_keys</code>, alongside the directory bind credential.</p>
{{end}}`

const adminAuditContent = `{{define "content"}}
<h1>Audit log</h1>
<p>Most recent {{.Limit}} entries, newest first. The client address is whatever
the request claimed it was.</p>
<table>
  <tr><th>When</th><th>User</th><th>Action</th><th>Detail</th><th>Client</th></tr>
  {{range .Entries}}
  <tr>
    <td>{{.OccurredAt}}</td>
    <td>{{.Username}}</td>
    <td>{{.Action}}</td>
    <td>{{.Detail}}</td>
    <td>{{.ClientIP}}</td>
  </tr>
  {{end}}
</table>
{{end}}`

const adminDiagnosticsContent = `{{define "content"}}
<h1>Diagnostics</h1>
<p>Reachability check for an internal host. Runs a single ICMP echo request and
returns the output of the command.</p>
<form method="get" action="/admin/diagnostics">
  <p><label>Host <input name="host" value="{{.Host}}" size="40"></label>
     <button type="submit">Run</button></p>
</form>
{{if .Ran}}
<h2>Output</h2>
<pre>{{.Output}}</pre>
{{if .CommandError}}<p class="error">command exited with: {{.CommandError}}</p>{{end}}
{{end}}
{{end}}`

const adminConfigContent = `{{define "content"}}
<h1>Configuration</h1>
<p>Effective configuration of the running process, read from the environment at
start-up.</p>
<table>
  <tr><th>Setting</th><th>Value</th></tr>
  {{range .Settings}}
  <tr><td><code>{{.Key}}</code></td><td><code>{{.Value}}</code></td></tr>
  {{end}}
</table>
<h2>Weakness toggles</h2>
<table>
  <tr><th>Toggle</th><th>State</th></tr>
  {{range .Weaknesses}}
  <tr><td><code>{{.Name}}</code></td><td>{{if .Enabled}}vulnerable{{else}}hardened{{end}}</td></tr>
  {{end}}
</table>
{{end}}`

var (
	adminIndexPage       = newPage("admin-index", adminIndexContent)
	adminUsersPage       = newPage("admin-users", adminUsersContent)
	adminAuditPage       = newPage("admin-audit", adminAuditContent)
	adminDiagnosticsPage = newPage("admin-diagnostics", adminDiagnosticsContent)
	adminConfigPage      = newPage("admin-config", adminConfigContent)
)

// ---------------------------------------------------------------------------
// View models
// ---------------------------------------------------------------------------

type toggleState struct {
	Name    string
	Enabled bool
}

type settingRow struct {
	Key   string
	Value string
}

type countRow struct {
	Table string
	Rows  int64
}

type auditRow struct {
	OccurredAt string
	Username   string
	Action     string
	Detail     string
	ClientIP   string
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleAdminIndex is the panel's front page: row counts, navigation, and the
// current posture. Showing the toggles here is deliberate — during a lab
// exercise the first question is always "is the weakness I am looking for
// actually switched on in this deployment".
func (s *Server) handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	counts, err := s.tableCounts(r.Context())
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}

	data := struct {
		pageCommon
		Counts     []countRow
		Weaknesses []toggleState
	}{
		pageCommon: pageCommon{Title: "Administration", Identity: identityOrNil(identity)},
		Counts:     counts,
		Weaknesses: toggleStates(s),
	}
	s.renderPage(w, adminIndexPage, data)
}

// handleAdminUsers lists every portal account, including the service accounts.
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	const query = `SELECT id, username, role, full_name, email, department FROM users ORDER BY id`
	rows, err := s.db.QueryContext(r.Context(), query)
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type userRow struct {
		ID         int64
		Username   string
		Role       string
		FullName   string
		Email      string
		Department string
	}

	var users []userRow
	for rows.Next() {
		var user userRow
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.FullName,
			&user.Email, &user.Department); err != nil {
			http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	s.audit(r, identity.Username, "admin.users.view", fmt.Sprintf("%d rows", len(users)))

	data := struct {
		pageCommon
		Users []userRow
	}{
		pageCommon: pageCommon{Title: "Users", Identity: identityOrNil(identity)},
		Users:      users,
	}
	s.renderPage(w, adminUsersPage, data)
}

// handleAdminAudit shows the audit trail. Reading the log is itself an admin
// action, and the fact that it is audited is the only reason an attacker cannot
// browse it silently — unless they are using the header bypass, which is
// recorded as anonymous instead.
func (s *Server) handleAdminAudit(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	const limit = 100
	const query = `SELECT occurred_at, username, action, detail, client_ip
	               FROM audit_log ORDER BY occurred_at DESC LIMIT $1`
	rows, err := s.db.QueryContext(r.Context(), query, limit)
	if err != nil {
		http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []auditRow
	for rows.Next() {
		var (
			entry      auditRow
			occurredAt time.Time
		)
		if err := rows.Scan(&occurredAt, &entry.Username, &entry.Action, &entry.Detail, &entry.ClientIP); err != nil {
			http.Error(w, s.errorDetail(err), http.StatusInternalServerError)
			return
		}
		entry.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
		entries = append(entries, entry)
	}

	s.audit(r, identity.Username, "admin.audit.view", fmt.Sprintf("%d rows", len(entries)))

	data := struct {
		pageCommon
		Limit   int
		Entries []auditRow
	}{
		pageCommon: pageCommon{Title: "Audit log", Identity: identityOrNil(identity)},
		Limit:      limit,
		Entries:    entries,
	}
	s.renderPage(w, adminAuditPage, data)
}

// hostnamePattern is the hardened form's idea of a hostname.
var hostnamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.\-]{0,252})$`)

// handleAdminDiagnostics runs a reachability check against a host the operator
// names.
//
// --- DELIBERATE WEAKNESS: WEAK_COMMAND_INJECTION_DIAGNOSTICS ---
//
// With the toggle on, the host parameter is interpolated into a command string
// that is handed to `sh -c`, so shell metacharacters in the parameter are
// interpreted by the shell rather than treated as part of an argument.
// "?host=127.0.0.1; id" runs `id` after the ping, and its output comes back in
// the response. Because the shell is involved, the injected command runs with
// the container process's identity and can read everything that process can:
// the application configuration, the environment (which holds the database
// password and the session signing key), and the document store.
//
// The demonstration does not depend on ping succeeding. A container usually
// cannot send ICMP without CAP_NET_RAW, so the ping half may well report a
// permission error — the injected command's output is the finding, and it is
// there either way.
//
// With the toggle off the command is run without a shell at all, with the host
// as a separate argument, and the host must first match a hostname pattern. No
// shell means no metacharacters to interpret, which is the fix: the parameter
// was never the problem, passing it through a shell was.
func (s *Server) handleAdminDiagnostics(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)
	host := r.URL.Query().Get("host")

	var (
		output    string
		runErr    error
		rejection string
	)

	if host != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var command *exec.Cmd
		if s.cfg.Weak.CommandInjectionDiagnostics {
			command = exec.CommandContext(ctx, "sh", "-c", "ping -c1 "+host)
		} else if hostnamePattern.MatchString(host) {
			command = exec.CommandContext(ctx, "ping", "-c1", host)
		} else {
			rejection = "the host must be a hostname or an IP address literal"
		}

		if command != nil {
			raw, err := command.CombinedOutput()
			output = string(raw)
			runErr = err
			s.audit(r, identity.Username, "admin.diagnostics", host)
		}
	}

	data := struct {
		pageCommon
		Host         string
		Ran          bool
		Output       string
		CommandError string
	}{
		pageCommon: pageCommon{Title: "Diagnostics", Identity: identityOrNil(identity), Error: rejection},
		Host:       host,
		Ran:        host != "" && rejection == "",
		Output:     output,
	}
	if runErr != nil && rejection == "" {
		data.CommandError = runErr.Error()
	}
	s.renderPage(w, adminDiagnosticsPage, data)
}

// handleAdminConfig prints the effective configuration.
//
// --- DELIBERATE WEAKNESS: WEAK_HARDCODED_DB_CREDENTIALS ---
//
// With the toggle on this page prints the live database password and states
// that it came from a compiled-in literal, which is the difference between
// "there is a password" and "there is a password that is the same on every
// deployment of this image". It demonstrates why a configuration dump is a
// credential disclosure even when the configuration is only read, and why the
// debug surface needs the same authentication and authorization as the data it
// describes.
//
// With the toggle off config.Settings redacts the value and reports the source
// as redacted instead. The page still exists — the fix is the redaction, not
// the removal, because an operator still needs to see the rest of the settings.
func (s *Server) handleAdminConfig(w http.ResponseWriter, r *http.Request) {
	identity, _ := s.identity(r)

	settings := s.cfg.Settings()
	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]settingRow, 0, len(settings))
	for _, key := range keys {
		rows = append(rows, settingRow{Key: key, Value: settings[key]})
	}

	data := struct {
		pageCommon
		Settings   []settingRow
		Weaknesses []toggleState
	}{
		pageCommon: pageCommon{Title: "Configuration", Identity: identityOrNil(identity)},
		Settings:   rows,
		Weaknesses: toggleStates(s),
	}
	s.renderPage(w, adminConfigPage, data)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// tableCounts gathers the row counts the admin front page shows.
func (s *Server) tableCounts(ctx context.Context) ([]countRow, error) {
	tables := []string{"users", "customers", "orders", "invoices", "documents", "api_keys", "audit_log"}

	counts := make([]countRow, 0, len(tables))
	for _, table := range tables {
		// The table name comes from the fixed list above, never from a request,
		// so interpolating it is not the same thing as the weaknesses in this
		// package — there is no input here for anyone to influence.
		var count int64
		if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			return nil, fmt.Errorf("counting %s: %w", table, err)
		}
		counts = append(counts, countRow{Table: table, Rows: count})
	}

	return counts, nil
}

// toggleStates renders the weakness report as a sorted slice, so that the two
// admin pages that show it agree on the order.
func toggleStates(s *Server) []toggleState {
	report := s.cfg.WeaknessReport()
	names := make([]string, 0, len(report))
	for name := range report {
		names = append(names, name)
	}
	sort.Strings(names)

	states := make([]toggleState, 0, len(names))
	for _, name := range names {
		states = append(states, toggleState{Name: name, Enabled: report[name]})
	}
	return states
}
