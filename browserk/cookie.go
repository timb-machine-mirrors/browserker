package browserk

import "time"

// Cookie properties
type Cookie struct {
	Name         string    `json:"name"`               // Cookie name.
	Value        string    `json:"value"`              // Cookie value.
	Domain       string    `json:"domain"`             // Cookie domain.
	Path         string    `json:"path"`               // Cookie path.
	Expires      float64   `json:"expires"`            // Cookie expiration date as the number of seconds since the UNIX epoch.
	Size         int       `json:"size"`               // Cookie size.
	HTTPOnly     bool      `json:"httpOnly"`           // True if cookie is http-only.
	Secure       bool      `json:"secure"`             // True if cookie is secure.
	Session      bool      `json:"session"`            // True in case of session cookie.
	SameSite     string    `json:"sameSite,omitempty"` // Cookie SameSite type. enum values: Strict, Lax, None
	Priority     string    `json:"priority"`           // Cookie Priority enum values: Low, Medium, High
	ObservedTime time.Time `json:"time_observed"`      // When the cookie was observed being set
}

// CookieAfterTime returns cookies that were observed after t
func CookieAfterTime(c []*Cookie, t time.Time) []*Cookie {
	cookies := make([]*Cookie, 0)
	for _, v := range c {
		if t.After(v.ObservedTime) {
			cookies = append(cookies, v)
		}
	}
	return cookies
}

// DiffCookies (should return new only)
func DiffCookies(oldCookies, newCookies []*Cookie) []*Cookie {
	if oldCookies == nil {
		return newCookies
	}

	if newCookies == nil {
		return nil
	}
	// TODO do comparison
	return newCookies
}
