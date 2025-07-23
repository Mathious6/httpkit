package tests

import (
	"strings"
	"testing"

	"github.com/Mathious6/httpkit"
	"github.com/Mathious6/httpkit/profiles"
	http "github.com/bogdanfinn/fhttp"
	"github.com/stretchr/testify/assert"
)

func TestWeb(t *testing.T) {
	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Firefox_110),
	}

	client, err := httpkit.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://registrierung.web.de/account/email-registration", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	_, err = client.Do(req)
	assert.NoError(t, err)
}
