package tls_client

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

type CookieJarOption func(config *cookieJarConfig)

type cookieJarConfig struct {
	logger       Logger
	skipExisting bool
	debug        bool
	allowEmpty   bool
}

// WithSkipExisting is used to skip existing cookies in the jar (default: false).
// This is useful when you want to add new cookies without overwriting existing ones.
func WithSkipExisting() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.skipExisting = true
	}
}

// WithAllowEmpty is used to allow empty cookies in the jar (default: false).
// Note: this is not recommended as empty cookies are usually not valid.
func WithAllowEmpty() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.allowEmpty = true
	}
}

func WithDebugLogger() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.debug = true
	}
}

func WithLogger(logger Logger) CookieJarOption {
	return func(config *cookieJarConfig) {
		config.logger = logger
	}
}

// CookieJar is the interface that wraps the basic CookieJar methods.
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
func NewCookieJar(options ...CookieJarOption) CookieJar {
	realJar, _ := cookiejar.New(nil)

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
		jar:     realJar,
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
				jar.config.logger.Debug("cookie %s already exists in jar. skipping", cookie.Name)
				continue
			}

			jar.config.logger.Debug("adding cookie %s to jar", cookie.Name)
			newCookies = append(newCookies, cookie)
		}
		cookies = append(existing, newCookies...)
	} else {
		var existingCookies []*http.Cookie
		for _, cookie := range existing {
			if findCookieByName(cookies, cookie.Name) != nil {
				jar.config.logger.Debug("cookie %s already exists in jar. skipping", cookie.Name)
				continue
			}

			jar.config.logger.Debug("adding cookie %s to jar", cookie.Name)
			existingCookies = append(existingCookies, cookie)
		}
		cookies = append(existingCookies, cookies...)
	}

	jar.jar.SetCookies(u, cookies)
	jar.cookies[hostKey] = cookies
}

// Cookies returns the cookies for the given url.
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
	seen := make(map[string]bool)
	var filteredCookies []*http.Cookie

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

// nonEmpty filters out empty cookies if allowEmptyCookies is false.
func (jar *cookieJar) nonEmpty(cookies []*http.Cookie) []*http.Cookie {
	if jar.config.allowEmpty {
		return cookies
	}

	var filteredCookies []*http.Cookie
	for _, cookie := range cookies {
		if cookie.Value == "" {
			jar.config.logger.Debug("cookie %s is empty and will be filtered out", cookie.Name)
			continue
		}

		filteredCookies = append(filteredCookies, cookie)
	}

	return filteredCookies
}

// notExpired filters out expired cookies.
func (jar *cookieJar) notExpired(cookies []*http.Cookie) []*http.Cookie {
	var filteredCookies []*http.Cookie

	for _, c := range cookies {
		// we misuse the max age here for "deletion" reasons. To be 100% correct a MaxAge equals 0 should also be deleted but we do not do it for now.
		if c.MaxAge <= CookieExpired {
			jar.config.logger.Debug("cookie %s in jar max age set to 0 or below. will be excluded from request", c.Name)
			continue
		}

		// TODO: this is currently commented out as the cookie parser does not parse the expire correctly out of the Set-Cookie header.
		/*if c.Expires.Before(now) {
			jar.config.logger.Debug("cookie %s in jar expired. will be excluded from request", c.Name)
			continue
		}*/

		filteredCookies = append(filteredCookies, c)
	}

	return filteredCookies
}

// findCookieByName returns the cookie with the given name if it exists and is not expired.
func findCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}
