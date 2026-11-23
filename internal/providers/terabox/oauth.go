package terabox

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const oauthBaseURL = "https://www.terabox.com"

type OAuthConfig struct {
	ClientID      string
	ClientSecret  string
	PrivateSecret string
	BaseURL       string
	HTTPClient    *http.Client
}

type OAuthClient struct {
	cfg  OAuthConfig
	http *http.Client
}

type OAuthError struct {
	Code    int
	Message string
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("TeraBox OAuth error %d: %s", e.Code, e.Message)
}

type DeviceAuthorization struct {
	DeviceCode      string `json:"device_code"`
	QRCodeURL       string `json:"qrcode_url"`
	ExpiresIn       int64  `json:"expires_in"`
	Interval        int64  `json:"interval"`
	UserCode        string `json:"user_code,omitempty"`
	VerificationURL string `json:"verification_url,omitempty"`
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type,omitempty"`
}

type TokenInfo struct {
	ClientID    string `json:"client_id"`
	APIDomain   string `json:"api_domain"`
	UploadDomain string `json:"upload_domain"`
	ExpiresIn   int64  `json:"expires_in"`
	CreateTime  int64  `json:"create_time"`
	UserID      int64  `json:"user_id"`
}

type oauthEnvelope[T any] struct {
	Data    T      `json:"data"`
	Errno   int    `json:"errno"`
	ShowMsg string `json:"show_msg"`
}

func NewOAuthClient(cfg OAuthConfig) (*OAuthClient, error) {
	if strings.TrimSpace(cfg.ClientID) == "" ||
		strings.TrimSpace(cfg.ClientSecret) == "" ||
		strings.TrimSpace(cfg.PrivateSecret) == "" {
		return nil, errors.New("TeraBox OAuth client ID, client secret and private secret are required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = oauthBaseURL
	}

	u, err := url.Parse(cfg.BaseURL)
	localHTTP := u != nil && u.Scheme == "http" &&
		(u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1")
	if err != nil || u == nil || u.Host == "" || (u.Scheme != "https" && !localHTTP) {
		return nil, errors.New("OAuth base URL must use HTTPS (HTTP is allowed only for loopback tests)")
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &OAuthClient{cfg: cfg, http: client}, nil
}

func (c *OAuthClient) signature(timestamp int64) string {
	sum := md5.Sum([]byte(c.cfg.ClientID + "_" +
		strconv.FormatInt(timestamp, 10) + "_" +
		c.cfg.ClientSecret + "_" + c.cfg.PrivateSecret))
	return hex.EncodeToString(sum[:])
}

func (c *OAuthClient) BeginDeviceAuthorization(ctx context.Context) (DeviceAuthorization, error) {
	q := url.Values{"client_id": {c.cfg.ClientID}}
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/oauth/devicecode?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DeviceAuthorization{}, err
	}

	var env oauthEnvelope[DeviceAuthorization]
	if err := c.do(req, &env); err != nil {
		return DeviceAuthorization{}, err
	}
	if env.Data.DeviceCode == "" || env.Data.ExpiresIn <= 0 || env.Data.Interval <= 0 {
		return DeviceAuthorization{}, errors.New("TeraBox returned incomplete device authorization data")
	}
	return env.Data, nil
}

// ExchangeCode exchanges an approved device code or authorization code for tokens.
func (c *OAuthClient) ExchangeCode(ctx context.Context, grantType, code string) (OAuthToken, error) {
	if grantType != "authorization_code" && grantType != "device_code" {
		return OAuthToken{}, errors.New("unsupported TeraBox OAuth grant type")
	}
	if strings.TrimSpace(code) == "" {
		return OAuthToken{}, errors.New("authorization code is required")
	}
	v := c.signedValues()
	v.Set("grant_type", grantType)
	v.Set("code", code)
	return c.postToken(ctx, "/oauth/gettoken", v)
}

// WaitForDeviceAuthorization polls at the interval supplied by TeraBox until
// approval, cancellation, or expiration. Errno 400001 represents pending approval.
func (c *OAuthClient) WaitForDeviceAuthorization(ctx context.Context, device DeviceAuthorization) (OAuthToken, error) {
	if device.DeviceCode == "" || device.ExpiresIn <= 0 || device.Interval <= 0 {
		return OAuthToken{}, errors.New("invalid TeraBox device authorization")
	}

	deadline := time.NewTimer(time.Duration(device.ExpiresIn) * time.Second)
	defer deadline.Stop()

	for {
		poll := time.NewTimer(time.Duration(device.Interval) * time.Second)
		select {
		case <-ctx.Done():
			poll.Stop()
			return OAuthToken{}, ctx.Err()
		case <-deadline.C:
			poll.Stop()
			return OAuthToken{}, errors.New("TeraBox device authorization expired")
		case <-poll.C:
		}

		token, err := c.ExchangeCode(ctx, "device_code", device.DeviceCode)
		if err == nil {
			return token, nil
		}
		var apiErr *OAuthError
		if !errors.As(err, &apiErr) || apiErr.Code != 400001 {
			return OAuthToken{}, err
		}
	}
}

func (c *OAuthClient) Refresh(ctx context.Context, refreshToken string) (OAuthToken, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return OAuthToken{}, errors.New("refresh token is required")
	}
	v := c.signedValues()
	v.Set("refresh_token", refreshToken)
	// Refresh tokens are single-use; persist the returned replacement atomically.
	return c.postToken(ctx, "/oauth/refreshtoken", v)
}

func (c *OAuthClient) TokenInformation(ctx context.Context, accessToken string) (TokenInfo, error) {
	if strings.TrimSpace(accessToken) == "" {
		return TokenInfo{}, errors.New("access token is required")
	}
	v := url.Values{"access_token": {accessToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.cfg.BaseURL, "/")+"/oauth/tokeninfo", strings.NewReader(v.Encode()))
	if err != nil {
		return TokenInfo{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var env oauthEnvelope[TokenInfo]
	if err := c.do(req, &env); err != nil {
		return TokenInfo{}, err
	}
	if env.Data.APIDomain == "" {
		return TokenInfo{}, errors.New("TeraBox token information omitted API domain")
	}
	return env.Data, nil
}

func (c *OAuthClient) signedValues() url.Values {
	now := time.Now().Unix()
	return url.Values{
		"client_id":     {c.cfg.ClientID},
		"client_secret": {c.cfg.ClientSecret},
		"timestamp":     {strconv.FormatInt(now, 10)},
		"sign":          {c.signature(now)},
	}
}

func (c *OAuthClient) postToken(ctx context.Context, path string, values url.Values) (OAuthToken, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, items := range values {
		for _, value := range items {
			if err := writer.WriteField(key, value); err != nil {
				return OAuthToken{}, err
			}
		}
	}
	if err := writer.Close(); err != nil {
		return OAuthToken{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.cfg.BaseURL, "/")+path, &body)
	if err != nil {
		return OAuthToken{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var env oauthEnvelope[OAuthToken]
	if err := c.do(req, &env); err != nil {
		return OAuthToken{}, err
	}
	if env.Data.AccessToken == "" || env.Data.RefreshToken == "" || env.Data.ExpiresIn <= 0 {
		return OAuthToken{}, errors.New("TeraBox returned incomplete token data")
	}
	return env.Data, nil
}

func (c *OAuthClient) do(req *http.Request, destination any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("TeraBox OAuth request failed: %w", err)
	}
	defer resp.Body.Close()

	const maxResponse = 1 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil {
		return fmt.Errorf("read TeraBox OAuth response: %w", err)
	}
	if len(data) > maxResponse {
		return errors.New("TeraBox OAuth response exceeded 1 MiB")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("TeraBox OAuth returned HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("decode TeraBox OAuth response: %w", err)
	}

	var status struct {
		Errno   int    `json:"errno"`
		ShowMsg string `json:"show_msg"`
	}
	if err := json.Unmarshal(data, &status); err == nil && status.Errno != 0 {
		message := strings.TrimSpace(status.ShowMsg)
		if message == "" {
			message = "request rejected"
		}
		return &OAuthError{Code: status.Errno, Message: message}
	}
	return nil
}
