package httpkit

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"maps"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
)

const CookieExpired = -1

// CookieJarOption is a function that modifies the configuration of a cookieJar.
type CookieJarOption func(config *cookieJarConfig)

type cookieJarConfig struct {
	logger       Logger
	skipExisting bool
	debug        bool
	allowEmpty   bool
}

// WithSkipExisting returns a CookieJarOption that skips existing cookies in the jar (default: false).
// This is useful when you want to add new cookies without overwriting existing ones.
func WithSkipExisting() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.skipExisting = true
	}
}

// WithAllowEmpty returns a CookieJarOption that allows empty cookies in the jar (default: false).
// Note: this is not recommended as empty cookies are usually not valid.
func WithAllowEmpty() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.allowEmpty = true
	}
}

// WithDebugLogger returns a CookieJarOption that enables debug logging for the cookie jar.
func WithDebugLogger() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.debug = true
	}
}

// WithLogger returns a CookieJarOption that sets a custom logger for the cookie jar.
func WithLogger(logger Logger) CookieJarOption {
	return func(config *cookieJarConfig) {
		config.logger = logger
	}
}

// CookieJar is the interface that wraps the basic CookieJar methods, including additional helpers.
type CookieJar interface {
	http.CookieJar
	CookiesMap() map[string][]*http.Cookie
	Cookie(u *url.URL, name string) *http.Cookie
}

type cookieJar struct {
	jar     *cookiejar.Jar
	config  *cookieJarConfig
	cookies map[string][]*http.Cookie
	sync.RWMutex
}

// NewCookieJar creates a new empty cookie jar with the given options.
// Returns nil if the underlying cookiejar.Jar cannot be created.
func NewCookieJar(options ...CookieJarOption) CookieJar {
	underlyingJar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	config := &cookieJarConfig{}

	for _, opt := range options {
		opt(config)
	}

	if config.logger == nil {
		config.logger = NewNoopLogger()
	}

	if config.debug {
		config.logger = NewDebugLogger(config.logger)
	}

	return &cookieJar{
		jar:     underlyingJar,
		config:  config,
		cookies: make(map[string][]*http.Cookie),
	}
}

// SetCookies sets the cookies for the given URL according to the rules defined in the config.
func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	defer jar.Unlock()

	cookies = jar.nonEmpty(cookies)
	cookies = jar.unique(cookies)

	hostKey := jar.buildCookieHostKey(u)
	existing := jar.cookies[hostKey]

	if jar.config.skipExisting {
		var newCookies []*http.Cookie
		for _, cookie := range cookies {
			if findCookieByName(existing, cookie.Name) != nil {
				jar.config.logger.Debug("[SetCookies] Cookie '%s' already exists in jar. Skipping.", cookie.Name)
				continue
			}
			jar.config.logger.Debug("[SetCookies] Adding new cookie '%s' to jar.", cookie.Name)
			newCookies = append(newCookies, cookie)
		}
		cookies = append(existing, newCookies...)
	} else {
		var existingCookies []*http.Cookie
		for _, cookie := range existing {
			if findCookieByName(cookies, cookie.Name) != nil {
				jar.config.logger.Debug("[SetCookies] Cookie '%s' already exists in jar. Skipping.", cookie.Name)
				continue
			}
			jar.config.logger.Debug("[SetCookies] Adding existing cookie '%s' to jar.", cookie.Name)
			existingCookies = append(existingCookies, cookie)
		}
		cookies = append(existingCookies, cookies...)
	}

	jar.jar.SetCookies(u, cookies)
	jar.cookies[hostKey] = cookies
}

// Cookies returns the cookies for the given url, filtering out expired cookies.
func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	hostKey := jar.buildCookieHostKey(u)
	cookies := jar.cookies[hostKey]

	return jar.notExpired(cookies)
}

// CookiesMap returns all cookies in the jar, grouped by host key.
func (jar *cookieJar) CookiesMap() map[string][]*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	copied := make(map[string][]*http.Cookie)
	maps.Copy(copied, jar.cookies)

	return copied
}

// Cookie returns the cookie with the given name for the given url or nil if not found.
func (jar *cookieJar) Cookie(u *url.URL, name string) *http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	hostKey := jar.buildCookieHostKey(u)
	cookies := jar.cookies[hostKey]

	for _, cookie := range cookies {
		if cookie.Name == name && cookie.MaxAge > CookieExpired {
			return cookie
		}
	}

	return nil
}

// buildCookieHostKey builds a host key for the cookie jar based on the URL.
// It uses the last two parts of the host (e.g. "example.com" or "sub.example.com") as the key.
func (jar *cookieJar) buildCookieHostKey(u *url.URL) string {
	hostParts := strings.Split(u.Host, ".")

	if len(hostParts) >= 2 {
		return fmt.Sprintf("%s.%s", hostParts[len(hostParts)-2], hostParts[len(hostParts)-1])
	}

	return u.Host
}

// unique filters out duplicate cookies by name and keeps the last one.
func (jar *cookieJar) unique(cookies []*http.Cookie) []*http.Cookie {
	seen := make(map[string]bool, len(cookies))
	filteredCookies := make([]*http.Cookie, 0, len(cookies))

	for i := len(cookies) - 1; i >= 0; i-- {
		c := cookies[i]
		if seen[c.Name] {
			continue
		}
		filteredCookies = append(filteredCookies, c)
		seen[c.Name] = true
	}

	return filteredCookies
}

// nonEmpty filters out empty cookies if allowEmpty is false.
func (jar *cookieJar) nonEmpty(cookies []*http.Cookie) []*http.Cookie {
	if jar.config.allowEmpty {
		return cookies
	}

	filteredCookies := cookies[:0]
	for _, cookie := range cookies {
		if cookie.Value == "" {
			jar.config.logger.Debug("[nonEmpty] Cookie '%s' is empty and will be filtered out.", cookie.Name)
			continue
		}
		filteredCookies = append(filteredCookies, cookie)
	}

	return filteredCookies
}

// notExpired filters out expired cookies.
func (jar *cookieJar) notExpired(cookies []*http.Cookie) []*http.Cookie {
	filteredCookies := cookies[:0]
	for _, cookie := range cookies {
		if cookie.MaxAge <= CookieExpired {
			jar.config.logger.Debug("[notExpired] Cookie '%s' in jar has max age <= 0. Will be excluded from request.", cookie.Name)
			continue
		}

		// TODO: The cookie parser does not parse the Expires field correctly from the Set-Cookie header.
		// Once fixed, consider also filtering cookies by expiration date here.
		// if cookie.Expires.Before(now) {
		// 	jar.config.logger.Debug("[notExpired] Cookie '%s' in jar expired. Will be excluded from request.", cookie.Name)
		// 	continue
		// }

		filteredCookies = append(filteredCookies, cookie)
	}
	return filteredCookies
}

// findCookieByName returns the cookie with the given name if it exists in the slice.
func findCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}
