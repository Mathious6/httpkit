package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mathious6/httpkit/profiles"

	"github.com/Mathious6/httpkit"
	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

func main() {
	sslPinning()
	requestToppsAsChrome107Client()
	postAsTlsClient()
	requestWithFollowRedirectSwitch()
	requestWithCustomClient()
	requestWithJa3CustomClientWithTwoGreaseExtensions()
	rotateProxiesOnClient()
	downloadImageWithTlsClient()
	testPskExtension()
	testALPSExtension()
}

type TlsBrowserleaksResponse struct {
	UserAgent  string `json:"user_agent"`
	Ja3Hash    string `json:"ja3_hash"`
	Ja3Text    string `json:"ja3_text"`
	Ja3NHash   string `json:"ja3n_hash"`
	Ja3NText   string `json:"ja3n_text"`
	Ja4        string `json:"ja4"`
	Ja4R       string `json:"ja4_r"`
	Ja4O       string `json:"ja4_o"`
	Ja4Ro      string `json:"ja4_ro"`
	AkamaiHash string `json:"akamai_hash"`
	AkamaiText string `json:"akamai_text"`
	TLS        struct {
		CipherSuite []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"cipher_suite"`
		ConnectionVersion []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"connection_version"`
		RecordVersion []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"record_version"`
		HandshakeVersion []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"handshake_version"`
		CipherSuites []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"cipher_suites"`
		Extensions []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
			Data  struct {
				SupportedVersions []struct {
					Name  string `json:"name"`
					Value int    `json:"value"`
				} `json:"supported_versions"`
			} `json:"data,omitempty"`
		} `json:"extensions"`
	} `json:"tls"`
	HTTP2 []struct {
		Type                string   `json:"type"`
		Length              int      `json:"length"`
		Settings            []string `json:"settings,omitempty"`
		WindowSizeIncrement int      `json:"window_size_increment,omitempty"`
		StreamID            int      `json:"stream_id,omitempty"`
		Headers             []string `json:"headers,omitempty"`
		Flags               []string `json:"flags,omitempty"`
		Priority            struct {
			Weight    int `json:"weight"`
			DepID     int `json:"dep_id"`
			Exclusive int `json:"exclusive"`
		} `json:"priority,omitempty"`
	} `json:"http2"`
}

type TlsApiResponse struct {
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Method      string `json:"method"`
	TLS         struct {
		TLSVersionRecord     string   `json:"tls_version_record"`
		TLSVersionNegotiated string   `json:"tls_version_negotiated"`
		Ja3                  string   `json:"ja3"`
		Ja3Hash              string   `json:"ja3_hash"`
		ClientRandom         string   `json:"client_random"`
		SessionID            string   `json:"session_id"`
		Ciphers              []string `json:"ciphers"`
		Extensions           []struct {
			EllipticCurvesPointFormats interface{} `json:"elliptic_curves_point_formats,omitempty"`
			Name                       string      `json:"name"`
			ServerName                 string      `json:"server_name,omitempty"`
			Data                       string      `json:"data,omitempty"`
			PskKeyExchangeMode         string      `json:"PSK_Key_Exchange_Mode,omitempty"`
			SupportedGroups            []string    `json:"supported_groups,omitempty"`
			Protocols                  []string    `json:"protocols,omitempty"`
			SignatureAlgorithms        []string    `json:"signature_algorithms,omitempty"`
			SharedKeys                 []struct {
				TLSGrease0X7A7A string `json:"TLS_GREASE (0x7a7a),omitempty"`
				X2551929        string `json:"X25519 (29),omitempty"`
			} `json:"shared_keys,omitempty"`
			Versions      []string `json:"versions,omitempty"`
			Algorithms    []string `json:"algorithms,omitempty"`
			StatusRequest struct {
				CertificateStatusType   string `json:"certificate_status_type"`
				ResponderIDListLength   int    `json:"responder_id_list_length"`
				RequestExtensionsLength int    `json:"request_extensions_length"`
			} `json:"status_request,omitempty"`
			PaddingDataLength int `json:"padding_data_length,omitempty"`
		} `json:"extensions"`
	} `json:"tls"`
	HTTP2 struct {
		AkamaiFingerprint     string `json:"akamai_fingerprint"`
		AkamaiFingerprintHash string `json:"akamai_fingerprint_hash"`
		SentFrames            []struct {
			FrameType string   `json:"frame_type"`
			Settings  []string `json:"settings,omitempty"`
			Headers   []string `json:"headers,omitempty"`
			Flags     []string `json:"flags,omitempty"`
			Priority  struct {
				Weight    int `json:"weight"`
				DependsOn int `json:"depends_on"`
				Exclusive int `json:"exclusive"`
			} `json:"priority,omitempty"`
			Length    int `json:"length"`
			Increment int `json:"increment,omitempty"`
			StreamID  int `json:"stream_id,omitempty"`
		} `json:"sent_frames"`
	} `json:"http2"`
	HTTP1 struct {
		Headers []string `json:"headers"`
	} `json:"http1"`
}

func sslPinning() {
	jar := httpkit.NewCookieJar()

	//	I generated the pins by running the following command:
	//	➜ hpkp-pins -server=bstn.com:443

	pins := map[string][]string{
		"bstn.com": {
			"NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=",
			"4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=",
			"x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4=",
		},
	}

	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(60),
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithRandomTLSExtensionOrder(),
		httpkit.WithCookieJar(jar),
		httpkit.WithCertificatePinning(pins, httpkit.DefaultBadPinHandler),
		httpkit.WithCharlesProxy("127.0.0.1", "8888"),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	u := "https://bstn.com"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":             {"*/*"},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"sec-ch-ua":          {`"Google Chrome";v="107", "Chromium";v="107", "Not=A?Brand";v="24"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"macOS"`},
		"sec-fetch-dest":     {"empty"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	err = resp.Body.Close()

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("GET %s : %d\n", u, resp.StatusCode)
}

func requestToppsAsChrome107Client() {
	jar := httpkit.NewCookieJar()

	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithDebug(),
		// httpkit.WithProxyUrl("http://user:pass@host:port"),
		// httpkit.WithNotFollowRedirects(),
		// httpkit.WithInsecureSkipVerify(),
		httpkit.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.topps.com/", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	// this will fail as topps does now enforce solving cloudflare. only tls is not enough anymore
	log.Printf("requesting topps as chrome107 => status code: %d\n", resp.StatusCode)

	u, err := url.Parse("https://www.topps.com/")
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("tls client cookies for url %s : %v\n", u.String(), client.GetCookies(u))
}

func postAsTlsClient() {
	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithClientProfile(profiles.Chrome_107),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	postData := url.Values{}
	postData.Add("foo", "bar")
	postData.Add("baz", "foo")

	req, err := http.NewRequest(http.MethodPost, "https://eonk4gg5hquk0g6.m.pipedream.net", strings.NewReader(postData.Encode()))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"content-type":    {"application/x-www-form-urlencoded"},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"content-type",
			"accept-language",
			"user-agent",
			"content-length",
			"host",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("POST Request status code: %d\n", resp.StatusCode)
}

func requestWithFollowRedirectSwitch() {
	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithNotFollowRedirects(),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://currys.co.uk/products/sony-playstation-5-digital-edition-825-gb-10205198.html", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting currys.co.uk without automatic redirect follow => status code: %d (Redirect Not Folloed)\n", resp.StatusCode)

	client.SetFollowRedirect(true)

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting currys.co.uk with automatic redirect follow => status code: %d (Redirect Followed)\n", resp.StatusCode)
}

func downloadImageWithTlsClient() {
	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithNotFollowRedirects(),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://avatars.githubusercontent.com/u/17678241?v=4", nil)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	log.Printf("requesting image => status code: %d\n", resp.StatusCode)

	ex, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}

	exPath := filepath.Dir(ex)

	fileName := fmt.Sprintf("%s/%s", exPath, "example-test.jpg")

	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("wrote file to: %s\n", fileName)
}

func rotateProxiesOnClient() {
	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(30),
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithProxyUrl("http://user:pass@host:port"), // you can also use socks5://user:pass@host:port or socks5h://user:pass@host:port
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 1 => ip: %s", tlsApiResponse.IP))

	// you need to put in here a valid proxy to make the example work
	err = client.SetProxy("http://user:pass@host:port")
	if err != nil {
		log.Println(err)
		return
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse = TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 2 => ip: %s", tlsApiResponse.IP))
}

func requestWithCustomClient() {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	pseudoHeaderOrder := []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	connectionFlow := uint32(15663105)

	specFactory := func() (tls.ClientHelloSpec, error) {
		return tls.ClientHelloSpec{
			CipherSuites: []uint16{
				tls.GREASE_PLACEHOLDER,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			CompressionMethods: []uint8{
				tls.CompressionNone,
			},
			Extensions: []tls.TLSExtension{
				&tls.UtlsGREASEExtension{},
				&tls.SNIExtension{},
				&tls.ExtendedMasterSecretExtension{},
				&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
				&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
					tls.CurveID(tls.GREASE_PLACEHOLDER),
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				}},
				&tls.SupportedPointsExtension{SupportedPoints: []byte{
					0,
				}},
				&tls.SessionTicketExtension{},
				&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&tls.StatusRequestExtension{},
				&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.ECDSAWithP256AndSHA256,
					tls.PSSWithSHA256,
					tls.PKCS1WithSHA256,
					tls.ECDSAWithP384AndSHA384,
					tls.PSSWithSHA384,
					tls.PKCS1WithSHA384,
					tls.PSSWithSHA512,
					tls.PKCS1WithSHA512,
				}},
				&tls.SCTExtension{},
				&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
					{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: tls.X25519},
				}},
				&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
					tls.PskModeDHE,
				}},
				&tls.SupportedVersionsExtension{Versions: []uint16{
					tls.VersionTLS13,
					tls.VersionTLS12,
					tls.VersionTLS11,
					tls.VersionTLS10,
				}},
				&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
					tls.CertCompressionBrotli,
				}},
				&tls.ApplicationSettingsExtension{
					SupportedProtocols: []string{"h2"},
				},
				&tls.UtlsGREASEExtension{},
				&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
			},
		}, nil
	}

	customClientProfile := profiles.NewClientProfile(tls.ClientHelloID{
		Client:      "MyCustomProfile",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}, settings, settingsOrder, pseudoHeaderOrder, connectionFlow, nil, nil)

	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(60),
		httpkit.WithClientProfile(customClientProfile), // use custom profile here
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.topps.com/", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting topps as customClient1 => status code: %d\n", resp.StatusCode)
}

func requestWithJa3CustomClientWithTwoGreaseExtensions() {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingEnablePush:        0,
		http2.SettingInitialWindowSize: 6291456,
		http2.SettingMaxHeaderListSize: 262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	pseudoHeaderOrder := []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	connectionFlow := uint32(15663105)

	ja3String := "771,2570-4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,2570-18-5-27-11-0-10-35-16-65037-51-13-23-43-17513-65281-45-2570,2570-25497-29-23-24,0"

	supportedSignatureAlgorithms := []string{
		"ECDSAWithP256AndSHA256",
		"PSSWithSHA256",
		"PKCS1WithSHA256",
		"ECDSAWithP384AndSHA384",
		"PSSWithSHA384",
		"PKCS1WithSHA384",
		"PSSWithSHA512",
		"PKCS1WithSHA512",
	}
	var supportedDelegatedCredentialsAlgorithms []string
	supportedVersions := []string{"GREASE", "1.3", "1.2"}
	keyShareCurves := []string{"GREASE", "X25519Kyber768", "X25519"}
	supportedProtocolsALPN := []string{"h2", "http/1.1"}
	supportedProtocolsALPS := []string{"h2"}
	echCandidateCipherSuites := []httpkit.CandidateCipherSuites{
		{
			KdfId:  "HKDF_SHA256",
			AeadId: "AEAD_AES_128_GCM",
		},
		{
			KdfId:  "HKDF_SHA256",
			AeadId: "AEAD_CHACHA20_POLY1305",
		},
	}
	candidatePayloads := []uint16{128, 160, 192, 224}
	certCompressionAlgos := []string{"brotli"}

	specFactory, err := httpkit.GetSpecFactoryFromJa3String(ja3String, supportedSignatureAlgorithms, supportedDelegatedCredentialsAlgorithms, supportedVersions, keyShareCurves, supportedProtocolsALPN, supportedProtocolsALPS, echCandidateCipherSuites, candidatePayloads, certCompressionAlgos, 0)

	customClientProfile := profiles.NewClientProfile(tls.ClientHelloID{
		Client:      "MyCustomProfile",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}, settings, settingsOrder, pseudoHeaderOrder, connectionFlow, nil, nil)

	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(60),
		httpkit.WithClientProfile(customClientProfile), // use custom profile here
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.browserleaks.com/tls", nil)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%s", string(readBytes))
	log.Printf("status code: %d\n", resp.StatusCode)
}

func testPskExtension() {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	pseudoHeaderOrder := []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	connectionFlow := uint32(15663105)

	specFactory := func() (tls.ClientHelloSpec, error) {
		return tls.ClientHelloSpec{
			CipherSuites: []uint16{
				tls.GREASE_PLACEHOLDER,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			CompressionMethods: []uint8{
				tls.CompressionNone,
			},
			Extensions: []tls.TLSExtension{
				&tls.UtlsGREASEExtension{},
				&tls.PSKKeyExchangeModesExtension{[]uint8{
					tls.PskModeDHE,
				}},
				&tls.KeyShareExtension{[]tls.KeyShare{
					{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: tls.X25519},
				}},
				&tls.ApplicationSettingsExtension{
					SupportedProtocols: []string{"h2"},
				},
				&tls.SupportedVersionsExtension{[]uint16{
					tls.GREASE_PLACEHOLDER,
					tls.VersionTLS13,
					tls.VersionTLS12,
				}},
				&tls.SNIExtension{},
				&tls.SupportedPointsExtension{SupportedPoints: []byte{
					tls.PointFormatUncompressed,
				}},
				&tls.StatusRequestExtension{},
				&tls.ExtendedMasterSecretExtension{},
				&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&tls.SupportedCurvesExtension{[]tls.CurveID{
					tls.CurveID(tls.GREASE_PLACEHOLDER),
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				}},
				&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
				&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
					tls.CertCompressionBrotli,
				}},
				&tls.SCTExtension{},
				&tls.SessionTicketExtension{},
				&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.ECDSAWithP256AndSHA256,
					tls.PSSWithSHA256,
					tls.PKCS1WithSHA256,
					tls.ECDSAWithP384AndSHA384,
					tls.PSSWithSHA384,
					tls.PKCS1WithSHA384,
					tls.PSSWithSHA512,
					tls.PKCS1WithSHA512,
				}},
				&tls.UtlsGREASEExtension{},
				&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				&tls.UtlsPreSharedKeyExtension{OmitEmptyPsk: true},
			},
		}, nil
	}

	customClientProfile := profiles.NewClientProfile(tls.ClientHelloID{
		Client:      "MyCustomProfileWithPSK",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}, settings, settingsOrder, pseudoHeaderOrder, connectionFlow, nil, nil)

	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(60),
		httpkit.WithClientProfile(customClientProfile),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	if strings.Contains(tlsApiResponse.TLS.Ja3, "-41,") {
		log.Println("profile includes PSK extension (41)")
	} else {
		log.Println("profile does not include PSK extension (41)")
	}

	// Now we are doing the second request that the session resumption kicks in
	req, err = http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"user-agent",
		},
	}

	secondResp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer secondResp.Body.Close()

	secondRespReadBytes, err := io.ReadAll(secondResp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	secondTlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(secondRespReadBytes, &secondTlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	if strings.Contains(secondTlsApiResponse.TLS.Ja3, "-41,") {
		log.Println("profile includes PSK extension (41)")
	} else {
		log.Println("profile does not include PSK extension (41)")
	}
}

func testALPSExtension() {
	options := []httpkit.HttpClientOption{
		httpkit.WithTimeoutSeconds(60),
		httpkit.WithClientProfile(profiles.Chrome_133),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.browserleaks.com", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse := TlsBrowserleaksResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	if strings.Contains(tlsApiResponse.Ja3Text, "17613") && !strings.Contains(tlsApiResponse.Ja3Text, "17513") {
		log.Println("profile includes new ALPS extension (17613) and not old one (17513)")
	} else {
		log.Println("profile does not include new ALPS extension (17613)")
	}
}
