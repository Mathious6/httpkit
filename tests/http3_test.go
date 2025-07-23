package tests

import (
	"io"
	"strings"
	"testing"

	"github.com/Mathious6/httpkit"
	"github.com/Mathious6/httpkit/profiles"
	http "github.com/bogdanfinn/fhttp"
)

func TestHTTP3(t *testing.T) {
	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_133),
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithDebug(),
	}

	client, err := httpkit.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://http3.is/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "it does support HTTP/3!") {
		t.Fatal("Response did not contain HTTP3 result")
	}
}
