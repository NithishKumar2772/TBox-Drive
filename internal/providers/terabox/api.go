package terabox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

type APIConfig struct {
	AccessToken string
	TokenInfo TokenInfo
	ApplicationRoot string
	HTTPClient *http.Client
}

type APIClient struct {
	accessToken, apiBase, uploadBase, applicationRoot string
	http, downloadHTTP *http.Client
}

type Quota struct { Total uint64 `json:"total"`; Used uint64 `json:"used"` }
type apiFile struct {
	FSID uint64 `json:"fs_id"`
	Path string `json:"path"`
	Name string `json:"server_filename"`
	Size uint64 `json:"size"`
	IsDir int `json:"isdir"`
	Modified uint64 `json:"server_mtime"`
}
type listResponse struct {
	Errno int `json:"errno"`
	ShowMsg string `json:"show_msg"`
	Info []apiFile `json:"info"`
	List []apiFile `json:"list"`
	HasMore int `json:"has_more"`
}
type quotaResponse struct {
	Errno int `json:"errno"`
	ShowMsg string `json:"show_msg"`
	Total uint64 `json:"total"`
	Used uint64 `json:"used"`
}

func NewAPIClient(cfg APIConfig) (*APIClient, error) {
	if strings.TrimSpace(cfg.AccessToken) == "" {
		return nil, errors.New("TeraBox access token is required")
	}
	apiBase, err := serviceBase(cfg.TokenInfo.APIDomain, serviceDomainAPI)
	if err != nil {
		return nil, fmt.Errorf("invalid TeraBox API domain: %w", err)
	}
	uploadBase, err := serviceBase(cfg.TokenInfo.UploadDomain, serviceDomainUpload)
	if err != nil {
		return nil, fmt.Errorf("invalid TeraBox upload domain: %w", err)
	}
	root, err := normalizeAbsoluteCloudPath(cfg.ApplicationRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid TeraBox application root: %w", err)
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	// Never follow redirects for requests carrying OAuth tokens. A redirect
	// could otherwise forward query-string credentials to an untrusted host.
	safeClient := *client
	safeClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	downloadClient := safeClient
	downloadClient.Timeout = 0

	return &APIClient{
		accessToken: cfg.AccessToken,
		apiBase: apiBase,
		uploadBase: uploadBase,
		applicationRoot: root,
		http: &safeClient,
		downloadHTTP: &downloadClient,
	}, nil
}


func NewAPIClientWithTestDomains(cfg APIConfig) (*APIClient,error) {
	if cfg.HTTPClient == nil { return nil, errors.New("test HTTP client is required") }
	if strings.TrimSpace(cfg.AccessToken) == "" { return nil, errors.New("TeraBox access token is required") }
	parseLoopback := func(raw string) (string,error) {
		u,err:=url.Parse(raw)
		if err!=nil || u==nil || u.Host=="" || u.User!=nil || u.RawQuery!="" || u.Fragment!="" || (u.Scheme!="http" && u.Scheme!="https") { return "", errors.New("invalid loopback test origin") }
		host:=u.Hostname()
		if host!="127.0.0.1" && host!="localhost" && host!="::1" { return "", errors.New("test domains must use loopback hosts") }
		return strings.TrimRight(u.String(),"/"),nil
	}
	apiBase,err:=parseLoopback(cfg.TokenInfo.APIDomain); if err!=nil{return nil,err}
	uploadBase,err:=parseLoopback(cfg.TokenInfo.UploadDomain); if err!=nil{return nil,err}
	root,err:=normalizeAbsoluteCloudPath(cfg.ApplicationRoot); if err!=nil{return nil,err}
	safe:=*cfg.HTTPClient
	safe.CheckRedirect=func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}
	download:=safe;download.Timeout=0
	return &APIClient{accessToken:cfg.AccessToken,apiBase:apiBase,uploadBase:uploadBase,applicationRoot:root,http:&safe,downloadHTTP:&download},nil
}

type serviceDomainKind int
const (
	serviceDomainAPI serviceDomainKind = iota
	serviceDomainUpload
)

func serviceBase(raw string, kind serviceDomainKind) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" { return "", errors.New("domain is empty") }
	if !strings.Contains(raw, "://") { raw = "https://" + raw }
	u, err := url.Parse(raw)
	if err != nil { return "", err }
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if u.Scheme != "https" || u.Hostname() == "" || u.Port() != "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("service domain must be a trusted HTTPS origin")
	}
	if !isTrustedServiceHost(host, kind) {
		return "", fmt.Errorf("untrusted TeraBox service hostname %q", host)
	}
	return "https://" + host, nil
}

func isTrustedServiceHost(host string, kind serviceDomainKind) bool {
	if host == "terabox.com" || strings.HasSuffix(host, ".terabox.com") ||
		host == "terabox.app" || strings.HasSuffix(host, ".terabox.app") {
		return true
	}
	// TeraBox test fixtures use loopback HTTP, but that exception is handled
	// only by test clients injected explicitly into APIConfig.
	return false
}

func (c *APIClient) ListDirectory(ctx context.Context, dir string, page, pageSize int) ([]provider.File, bool, error) {
	if page < 1 { return nil, false, errors.New("page must be at least 1") }
	if pageSize < 1 || pageSize > 10000 { return nil, false, errors.New("page size must be between 1 and 10000") }
	if err := c.validateRemotePath(dir); err != nil { return nil, false, err }
	q := url.Values{
		"access_tokens": {c.accessToken},
		"dir": {dir},
		"page": {strconv.Itoa(page)},
		"num": {strconv.Itoa(pageSize)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/openapi/api/list?"+q.Encode(), nil)
	if err != nil { return nil, false, err }
	var result listResponse
	if err := c.getJSON(req, &result); err != nil { return nil, false, err }
	items := result.Info
	if items == nil { items = result.List }
	files := make([]provider.File, 0, len(items))
	for _, item := range items {
		name := item.Name
		if name == "" { name = pathBase(item.Path) }
		files = append(files, provider.File{
			ID: strconv.FormatUint(item.FSID, 10), Name: name, Path: item.Path,
			Size: int64(item.Size), IsDirectory: item.IsDir == 1,
			ModifiedAt: time.Unix(int64(item.Modified), 0),
		})
	}
	return files, result.HasMore == 1, nil
}

func (c *APIClient) StorageQuota(ctx context.Context) (Quota, error) {
	q := url.Values{"access_tokens": {c.accessToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/openapi/api/quota?"+q.Encode(), nil)
	if err != nil { return Quota{}, err }
	var result quotaResponse
	if err := c.getJSON(req, &result); err != nil { return Quota{}, err }
	return Quota{Total: result.Total, Used: result.Used}, nil
}

func normalizeAbsoluteCloudPath(value string) (string, error) {
	if strings.TrimSpace(value) == "" { return "", errors.New("path is empty") }
	if !strings.HasPrefix(value, "/") || strings.Contains(value, "\\") {
		return "", errors.New("path must be an absolute cloud path")
	}
	parts := strings.Split(value, "/")
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			return "", errors.New("path cannot contain traversal segments")
		default:
			clean = append(clean, part)
		}
	}
	if len(clean) == 0 { return "/", nil }
	return "/" + strings.Join(clean, "/"), nil
}

func (c *APIClient) validateRemotePath(remotePath string) error {
	clean, err := normalizeAbsoluteCloudPath(remotePath)
	if err != nil { return err }
	if c.applicationRoot != "" && !isWithinCloudRoot(c.applicationRoot, clean) {
		return errors.New("path is outside the TeraBox application folder")
	}
	return nil
}

func isWithinCloudRoot(root, candidate string) bool {
	root = strings.TrimRight(root, "/")
	candidate = strings.TrimRight(candidate, "/")
	return candidate == root || strings.HasPrefix(candidate, root+"/")
}

func (c *APIClient) getJSON(req *http.Request, dst any) error {
	resp, err := c.http.Do(req)
	if err != nil { return fmt.Errorf("TeraBox API request failed: %w", err) }
	defer resp.Body.Close()
	const max = 4 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil { return fmt.Errorf("read TeraBox API response: %w", err) }
	if len(body) > max { return errors.New("TeraBox API response exceeded 4 MiB") }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("TeraBox API returned HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, dst); err != nil { return fmt.Errorf("decode TeraBox API response: %w", err) }
	var status struct {
		Errno int `json:"errno"`
		ShowMsg string `json:"show_msg"`
	}
	if err := json.Unmarshal(body, &status); err == nil && status.Errno != 0 {
		msg := strings.TrimSpace(status.ShowMsg)
		if msg == "" { msg = "request rejected" }
		return fmt.Errorf("TeraBox API error %d: %s", status.Errno, msg)
	}
	return nil
}

func pathBase(p string) string {
	p = strings.TrimRight(p, "/")
	if p == "" { return "/" }
	i := strings.LastIndex(p, "/")
	return p[i+1:]
}
