package httpkit

import (
	"net/url"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/stretchr/testify/assert"
)

const (
	Alive   = 1
	Expired = -1
)

var urlObject = &url.URL{Scheme: "http", Host: "test.com", Path: "/test"}

func TestCookieJar_GivenEmptyJar_WhenSetCookies_ThenCookiesAreAdded(t *testing.T) {
	jar := NewCookieJar()

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Alive}})

	assert.Equal(t, 1, len(jar.Cookies(urlObject)), "Expected cookie to be added")
	assert.Equal(t, "first", jar.Cookie(urlObject, "1").Value, "Expected value to be set")
}

func TestCookieJar_GivenEmptyJar_WhenSetDuplicateCookies_ThenLastCookiesAreAdded(t *testing.T) {
	jar := NewCookieJar()

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Alive}})
	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "second", MaxAge: Alive}})

	assert.Equal(t, 1, len(jar.Cookies(urlObject)), "Only one cookie should be present after unique filtering")
	assert.Equal(t, "second", jar.Cookie(urlObject, "1").Value, "Expected value to be set to the last one")
}

func TestCookieJar_GivenExpiredCookie_WhenSetCookies_ThenExpiredCookieIsExcluded(t *testing.T) {
	jar := NewCookieJar()

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Expired}})

	assert.Equal(t, 0, len(jar.Cookies(urlObject)), "Expected expired cookie to be excluded")
}

func TestCookieJar_GivenSkipExisting_WhenSetExistingCookies_ThenExistingCookiesAreSkipped(t *testing.T) {
	jar := NewCookieJar(WithSkipExisting())

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Alive}})
	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "second", MaxAge: Alive}})

	assert.Equal(t, 1, len(jar.Cookies(urlObject)), "Only one cookie should be present after skipping existing")
	assert.Equal(t, "first", jar.Cookie(urlObject, "1").Value, "Expected existing value to be kept")
}

func TestCookieJar_GivenAllowEmpty_WhenSetEmptyCookies_ThenEmptyCookiesAreAllowed(t *testing.T) {
	jar := NewCookieJar(WithAllowEmpty())

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "", MaxAge: Alive}})

	assert.Equal(t, 1, len(jar.Cookies(urlObject)), "Expected empty cookie to be accepted")
}

func TestCookieJar_GivenAliveCookie_WhenSetExpiredCookies_ThenOldCookiesIsExcluded(t *testing.T) {
	jar := NewCookieJar()
	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Alive}})

	jar.SetCookies(urlObject, []*http.Cookie{{Name: "1", Value: "first", MaxAge: Expired}})

	assert.Equal(t, 0, len(jar.Cookies(urlObject)), "Expected expired cookie to be excluded")
}
