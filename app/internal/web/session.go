package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clayface/app/internal/config"
)

// sessionCookieName is the cookie the portal keeps its identity in. It is a
// single cookie holding an opaque string; everything else about the session is
// derived from it, on every request, with no server-side session table. That
// choice is what makes the forgeable-token weakness worth demonstrating: there
// is no server state to invalidate a forged token against.
const sessionCookieName = "clayface_session"

// Session is the identity carried inside the cookie: who the caller says they
// are and what role they claim. It is deliberately small, because in the
// vulnerable posture the client controls all of it.
type Session struct {
	Username string
	Role     string
	IssuedAt int64
}

// encodeSession turns a Session into the cookie value.
//
// --- DELIBERATE WEAKNESS: WEAK_FORGEABLE_SESSION_TOKEN ---
//
// With the toggle on the cookie is the base64url encoding of
// "username|role|issuedAt" and nothing else: no signature, no message
// authentication code, no server-side lookup. Anyone who can read their own
// cookie can write a different one, and because nothing verifies it, a client
// can mint "adm-hermione|admin|<now>" and be treated as an administrator. The
// weakness demonstrates why a session token must be integrity-protected by the
// server rather than merely encoded by it.
//
// With the toggle off the same payload is signed with HMAC-SHA256 keyed on
// SESSION_SECRET, and decodeSession verifies that signature before trusting
// anything inside. The payload format does not change; the difference is
// entirely in whether the server can tell a token it issued from a token
// somebody else wrote.
func encodeSession(cfg *config.Config, s Session) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(
		s.Username + "|" + s.Role + "|" + strconv.FormatInt(s.IssuedAt, 10)))

	if cfg.Weak.ForgeableSessionToken {
		return payload
	}

	return payload + "." + signPayload(cfg, payload)
}

// decodeSession parses a cookie value back into a Session, rejecting anything
// that fails verification.
func decodeSession(cfg *config.Config, raw string) (Session, bool) {
	if raw == "" {
		return Session{}, false
	}

	if cfg.Weak.ForgeableSessionToken {
		// --- DELIBERATE WEAKNESS: WEAK_FORGEABLE_SESSION_TOKEN ---
		//
		// The token is accepted for what it says. There is no key to check it
		// against and therefore no way to tell a token the portal issued from
		// one the client typed out, which is the whole of the weakness.
		return parsePayload(raw)
	}

	payload, signature, found := strings.Cut(raw, ".")
	if !found {
		return Session{}, false
	}
	// hmac.Equal rather than ==: a byte-by-byte comparison that returns early
	// leaks how much of a guessed signature was correct, which is enough to
	// forge one byte at a time.
	if !hmac.Equal([]byte(signPayload(cfg, payload)), []byte(signature)) {
		return Session{}, false
	}

	return parsePayload(payload)
}

// signPayload is the HMAC-SHA256 of the encoded payload, hex encoded. It is
// only ever called in the hardened posture — encodeSession skips it and
// decodeSession never reaches it while the toggle is on.
func signPayload(cfg *config.Config, payload string) string {
	mac := hmac.New(sha256.New, []byte(cfg.SessionSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// parsePayload decodes the "username|role|issuedAt" body shared by both
// postures. It checks shape only: three fields, a usable timestamp, a non-empty
// username. No field is checked against the database here — see Server.identity
// for why the role in particular must not be.
func parsePayload(payload string) (Session, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return Session{}, false
	}

	fields := strings.Split(string(decoded), "|")
	if len(fields) != 3 {
		return Session{}, false
	}

	issued, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return Session{}, false
	}
	if fields[0] == "" {
		return Session{}, false
	}

	return Session{Username: fields[0], Role: fields[1], IssuedAt: issued}, true
}

// setSessionCookie writes the session cookie.
//
// The cookie attributes differ between the postures for the same reason the
// token itself does: the vulnerable one is a token the client is trusted to
// hold and to read, and the hardened one is not.
func setSessionCookie(w http.ResponseWriter, cfg *config.Config, s Session) {
	cookie := &http.Cookie{
		Name:  sessionCookieName,
		Value: encodeSession(cfg, s),
		Path:  "/",
	}

	if !cfg.Weak.ForgeableSessionToken {
		// HttpOnly keeps the cookie away from JavaScript, so a cross-site
		// scripting bug elsewhere on the portal cannot read the session.
		// SameSite=Lax stops the cookie riding along on a cross-site form
		// submission. Secure keeps it off plaintext HTTP.
		cookie.HttpOnly = true
		cookie.Secure = true
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

// clearSessionCookie expires the cookie. The attributes have to match the ones
// it was set with, or the browser keeps the original alongside the empty one.
func clearSessionCookie(w http.ResponseWriter, cfg *config.Config) {
	cookie := &http.Cookie{
		Name:    sessionCookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}

	if !cfg.Weak.ForgeableSessionToken {
		cookie.HttpOnly = true
		cookie.Secure = true
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

// newSession builds a session for a freshly authenticated user.
func newSession(username, role string) Session {
	return Session{Username: username, Role: role, IssuedAt: time.Now().Unix()}
}
