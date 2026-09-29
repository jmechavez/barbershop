package http

import (
	"net/http"

	"barbershop/internal/app"
)

const sessionCookieName = "session"

// setSessionCookie writes the session cookie on the response.
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(app.SessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   false, // set true when behind HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie tells the browser to drop the cookie.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// sessionToken reads the session token from the request, or "" if absent.
func sessionToken(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
