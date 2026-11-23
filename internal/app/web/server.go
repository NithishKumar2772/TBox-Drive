package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/app"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

// Server exposes the shared application services to a same-origin browser UI.
// It is intentionally restricted to loopback; remote access needs a separate
// authenticated device-session design and is not enabled by this server.
type Server struct {
	service *app.TeraBoxService
	store *sqlite.Store
	applicationRoot string
	sessionToken string
	assets http.Handler
}

func NewServer(service *app.TeraBoxService, store *sqlite.Store, applicationRoot string) (*Server, error) {
	if service == nil || store == nil { return nil, errors.New("TeraBox service and local store are required") }
	if strings.TrimSpace(applicationRoot) == "" { return nil, errors.New("TeraBox application folder is required") }
	token, err := newSessionToken()
	if err != nil { return nil, fmt.Errorf("create local browser session token: %w", err) }
	return &Server{service: service, store: store, applicationRoot: strings.TrimRight(applicationRoot, "/"), sessionToken: token, assets: http.FileServer(http.FS(Assets))}, nil
}

// Handler returns the HTTP handler. All routes are same-origin and loopback only.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/session", s.createSession)
	mux.HandleFunc("GET /api/status", s.status)
	mux.HandleFunc("GET /api/accounts", s.accounts)
	mux.HandleFunc("POST /api/connect/start", s.connectStart)
	mux.HandleFunc("POST /api/connect/complete", s.connectComplete)
	mux.HandleFunc("POST /api/accounts/disconnect", s.disconnect)
	mux.HandleFunc("GET /api/files", s.files)
	mux.HandleFunc("GET /api/quota", s.quota)
	mux.HandleFunc("POST /api/upload", s.upload)
	mux.HandleFunc("GET /api/download", s.download)
	mux.HandleFunc("GET /api/transfers", s.transfers)
	mux.HandleFunc("POST /api/files/rename", s.rename)
	mux.HandleFunc("POST /api/files/delete", s.deleteFile)
	mux.HandleFunc("POST /api/files/move", s.moveFile)
	mux.HandleFunc("POST /api/files/copy", s.copyFile)
	mux.HandleFunc("/", s.serveAssets)
	return s.securityHeaders(mux)
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) { http.Error(w, "only loopback browser access is allowed", http.StatusForbidden); return }
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			origin := r.Header.Get("Origin")
			if origin == "" || origin != "http://"+r.Host {
				http.Error(w, "cross-origin request rejected", http.StatusForbidden); return
			}
			if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
				http.Error(w, "cross-site request rejected", http.StatusForbidden); return
			}
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/session" {
			cookie, err := r.Cookie("tbox_local_session")
			if err != nil || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(s.sessionToken)) != 1 {
				http.Error(w, "local browser session required; open the launch URL printed by TBox Drive", http.StatusUnauthorized); return
			}
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil { host = h }
	host = strings.Trim(host, "[]")
	return strings.EqualFold(host, "localhost") || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}

func (s *Server) serveAssets(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/assets/") {
		http.NotFound(w, r); return
	}
	if r.URL.Path == "/" { w.Header().Set("Cache-Control", "no-store") }
	s.assets.ServeHTTP(w, r)
}
func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-TBox-Session")
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.sessionToken)) != 1 {
		http.Error(w, "invalid local browser launch token", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "tbox_local_session", Value: s.sessionToken, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60,
	})
	w.WriteHeader(http.StatusNoContent)
}

func newSessionToken() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil { return "", err }
	return hex.EncodeToString(value[:]), nil
}

// LaunchURL returns the private local launch URL. The fragment is
// not sent in HTTP requests or Referer headers; the browser exchanges it for
// an HttpOnly same-site cookie.
func (s *Server) LaunchURL(address string) string {
	return "http://" + address + "/#" + s.sessionToken
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ready": true, "message": "TBox Drive local web service is ready", "applicationRoot": s.applicationRoot})
}
func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.ListAccounts(r.Context()); if err != nil { writeError(w, err); return }
	// Never expose CredentialRef or any provider token to browser JavaScript.
	type publicAccount struct { ID string `json:"id"`; ProviderID string `json:"providerID"`; DisplayName string `json:"displayName"` }
	out := make([]publicAccount, 0, len(items))
	for _, item := range items { out = append(out, publicAccount{item.ID, item.ProviderID, item.DisplayName}) }
	writeJSON(w, http.StatusOK, out)
}
func (s *Server) connectStart(w http.ResponseWriter, r *http.Request) {
	device, err := s.service.BeginConnect(r.Context()); if err != nil { writeError(w, err); return }
	writeJSON(w, http.StatusOK, device)
}
func (s *Server) connectComplete(w http.ResponseWriter, r *http.Request) {
	var device terabox.DeviceAuthorization
	if err := decodeJSON(w, r, &device); err != nil { writeError(w, err); return }
	item, err := s.service.CompleteConnect(r.Context(), device); if err != nil { writeError(w, err); return }
	writeJSON(w, http.StatusOK, map[string]string{"id": item.ID, "providerID": item.ProviderID, "displayName": item.DisplayName})
}
func (s *Server) disconnect(w http.ResponseWriter, r *http.Request) {
	var req struct { AccountID string `json:"accountID"` }
	if err := decodeJSON(w, r, &req); err != nil { writeError(w, err); return }
	if strings.TrimSpace(req.AccountID) == "" { http.Error(w, "accountID is required", http.StatusBadRequest); return }
	if err := s.service.Disconnect(r.Context(), req.AccountID); err != nil { writeError(w, err); return }
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) apiClient(r *http.Request) (*terabox.APIClient, string, error) {
	accountID := strings.TrimSpace(r.URL.Query().Get("account_id"))
	if accountID == "" { return nil, "", errors.New("account_id is required") }
	api, err := s.service.NewAPIClient(r.Context(), accountID)
	return api, accountID, err
}
func (s *Server) files(w http.ResponseWriter, r *http.Request) {
	api, _, err := s.apiClient(r); if err != nil { writeError(w, err); return }
	parent := r.URL.Query().Get("path")
	if parent == "" { parent = s.applicationRoot + "/" }
	if err := s.validateCloudPath(parent); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	page := parsePage(r.URL.Query().Get("page"))
	items, more, err := api.ListDirectory(r.Context(), parent, page, 100); if err != nil { writeError(w, err); return }
	next := ""; if more { next = strconv.Itoa(page + 1) }
	writeJSON(w, http.StatusOK, map[string]any{"files": items, "nextPageToken": next})
}
func (s *Server) quota(w http.ResponseWriter, r *http.Request) {
	api, _, err := s.apiClient(r); if err != nil { writeError(w, err); return }
	result, err := api.StorageQuota(r.Context()); if err != nil { writeError(w, err); return }
	writeJSON(w, http.StatusOK, result)
}
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader(); if err != nil { http.Error(w, "multipart upload is required", http.StatusBadRequest); return }
	var accountID, parentPath, name string
	var temp *os.File
	var fileSize int64
	defer func(){ if temp != nil { _ = temp.Close(); _ = os.Remove(temp.Name()) } }()
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) { break }
		if nextErr != nil { writeError(w, nextErr); return }
		switch part.FormName() {
		case "account_id": b, e := io.ReadAll(io.LimitReader(part, 4096)); if e != nil { writeError(w,e); return }; accountID = strings.TrimSpace(string(b))
		case "parent_path": b, e := io.ReadAll(io.LimitReader(part, 8192)); if e != nil { writeError(w,e); return }; parentPath = strings.TrimSpace(string(b))
		case "name": b, e := io.ReadAll(io.LimitReader(part, 4096)); if e != nil { writeError(w,e); return }; name = strings.TrimSpace(string(b))
		case "file":
			if temp != nil { http.Error(w, "only one file per upload request is supported", http.StatusBadRequest); return }
			if name == "" { name = filepath.Base(part.FileName()) }
			if err := validateFileName(name); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
			tempDir := os.Getenv("TBOX_UPLOAD_TEMP_DIR")
			if tempDir == "" { tempDir = os.TempDir() }
			temp, err = os.CreateTemp(tempDir, "tbox-web-upload-*"); if err != nil { writeError(w,err); return }
			fileSize, err = io.Copy(temp, part); if err != nil { writeError(w,err); return }
		}
		_ = part.Close()
	}
	if temp == nil { http.Error(w, "file field is required", http.StatusBadRequest); return }
	if accountID == "" { http.Error(w, "account_id is required", http.StatusBadRequest); return }
	if parentPath == "" { parentPath = s.applicationRoot }
	if err := s.validateCloudPath(parentPath); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	if err := temp.Sync(); err != nil { writeError(w,err); return }
	if err := temp.Close(); err != nil { writeError(w,err); return }
	api, err := s.service.NewAPIClient(r.Context(), accountID); if err != nil { writeError(w,err); return }
	id, err := randomID(); if err != nil { writeError(w,err); return }
	result, err := api.UploadFile(r.Context(), temp.Name(), path.Join(parentPath,name), terabox.UploadOptions{Repository:s.store, TransferID:id})
	if err != nil { writeError(w,err); return }
	writeJSON(w, http.StatusOK, map[string]any{"file":result.File,"taskID":id,"size":fileSize})
}
func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	api, _, err := s.apiClient(r)
	if err != nil { writeError(w, err); return }
	fileID := strings.TrimSpace(r.URL.Query().Get("file_id"))
	if fileID == "" { http.Error(w, "file_id is required", http.StatusBadRequest); return }
	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "." || name == "/" || name == "" { name = "download" }
	if err := validateFileName(name); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	body, size, err := api.DownloadWithSize(r.Context(), fileID)
	if err != nil { writeError(w, err); return }
	defer body.Close()

	var transferID string
	tracked := size >= 0
	if tracked {
		transferID, err = randomID()
		if err != nil { writeError(w, err); return }
		task := transfer.Task{ID: transferID, ProviderID: "terabox", Direction: "download", Source: fileID, Destination: name, Size: size, Status: transfer.StatusQueued}
		if err := s.store.CreateTransfer(r.Context(), task); err != nil { writeError(w, err); return }
		if err := s.store.UpdateProgress(r.Context(), transferID, 0, transfer.StatusRunning, ""); err != nil {
			_ = s.store.UpdateProgress(context.Background(), transferID, 0, transfer.StatusFailed, "could not start download tracking")
			writeError(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.Header().Set("Cache-Control", "no-store")
	if tracked { w.Header().Set("Content-Length", strconv.FormatInt(size, 10)) }

	var destination io.Writer = w
	var progress *downloadResponseWriter
	if tracked {
		progress = &downloadResponseWriter{writer: w, store: s.store, transferID: transferID, total: size}
		destination = progress
	}
	written, copyErr := io.Copy(destination, body)
	if progress != nil { progress.report() }
	if tracked {
		status, message := transfer.StatusCompleted, ""
		if copyErr != nil {
			status, message = transfer.StatusFailed, copyErr.Error()
			if errors.Is(r.Context().Err(), context.Canceled) { status, message = transfer.StatusCanceled, "download canceled" }
		} else if written != size {
			status, message = transfer.StatusFailed, fmt.Sprintf("download size mismatch: expected %d bytes, received %d", size, written)
		}
		transferred := written
		if transferred > size { transferred = size }
		_ = s.store.UpdateProgress(context.Background(), transferID, transferred, status, message)
	}
}

type downloadResponseWriter struct {
	writer io.Writer
	store *sqlite.Store
	transferID string
	total int64
	transferred int64
	lastReported int64
	lastReport time.Time
}

func (w *downloadResponseWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.transferred += int64(n)
	if n > 0 && (w.transferred-w.lastReported >= 1<<20 || time.Since(w.lastReport) >= 200*time.Millisecond || err != nil) {
		w.report()
	}
	return n, err
}

func (w *downloadResponseWriter) report() {
	if w.transferred == w.lastReported { return }
	transferred := w.transferred
	if transferred > w.total { transferred = w.total }
	_ = w.store.UpdateProgress(context.Background(), w.transferID, transferred, transfer.StatusRunning, "")
	w.lastReported = w.transferred
	w.lastReport = time.Now()
}
func (s *Server) transfers(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListTransfers(r.Context(),100); if err != nil { writeError(w,err); return }
	if items == nil { items = []transfer.Task{} }
	writeJSON(w,http.StatusOK,items)
}
func (s *Server) rename(w http.ResponseWriter, r *http.Request) {
	var req fileOperationRequest
	if err := decodeJSON(w,r,&req); err != nil { writeError(w,err); return }
	if err := s.validateCloudPath(req.Path); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	if err := validateFileName(req.Name); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	api,err := s.service.NewAPIClient(r.Context(),req.AccountID); if err != nil { writeError(w,err); return }
	if err := api.Rename(r.Context(),req.Path,req.Name); err != nil { writeError(w,err); return }; w.WriteHeader(http.StatusNoContent)
}
func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	var req fileOperationRequest
	if err := decodeJSON(w,r,&req); err != nil { writeError(w,err); return }
	if err := s.validateCloudPath(req.Path); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	api,err := s.service.NewAPIClient(r.Context(),req.AccountID); if err != nil { writeError(w,err); return }
	if err := api.DeletePath(r.Context(),req.Path); err != nil { writeError(w,err); return }; w.WriteHeader(http.StatusNoContent)
}
func (s *Server) moveFile(w http.ResponseWriter, r *http.Request) {
	var req fileOperationRequest
	if err := decodeJSON(w,r,&req); err != nil { writeError(w,err); return }
	if err := s.validateCloudPath(req.Path); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	if err := s.validateCloudPath(req.Destination); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	if isSameOrDescendantPath(req.Destination, req.Path) {
		http.Error(w, "destination cannot be the source folder or one of its descendants", http.StatusBadRequest)
		return
	}
	if req.Name != "" {
		if err := validateFileName(req.Name); err != nil { http.Error(w,err.Error(),http.StatusBadRequest); return }
	}
	api,err := s.service.NewAPIClient(r.Context(),req.AccountID); if err != nil { writeError(w,err); return }
	if err := api.Move(r.Context(),req.Path,req.Destination,req.Name); err != nil { writeError(w,err); return }; w.WriteHeader(http.StatusNoContent)
}
func (s *Server) copyFile(w http.ResponseWriter, r *http.Request) {
	var req fileOperationRequest
	if err := decodeJSON(w, r, &req); err != nil { writeError(w, err); return }
	if err := s.validateCloudPath(req.Path); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := s.validateCloudPath(req.Destination); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if isSameOrDescendantPath(req.Destination, req.Path) {
		http.Error(w, "destination cannot be the source folder or one of its descendants", http.StatusBadRequest)
		return
	}
	if req.Name != "" {
		if err := validateFileName(req.Name); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	}
	api, err := s.service.NewAPIClient(r.Context(), req.AccountID)
	if err != nil { writeError(w, err); return }
	if err := api.Copy(r.Context(), req.Path, req.Destination, req.Name); err != nil { writeError(w, err); return }
	w.WriteHeader(http.StatusNoContent)
}
type fileOperationRequest struct {
	AccountID string `json:"accountID"`
	Path string `json:"path"`
	Name string `json:"name"`
	Destination string `json:"destination"`
}
func isSameOrDescendantPath(candidate, parent string) bool {
	cleanCandidate := path.Clean(candidate)
	cleanParent := path.Clean(parent)
	return cleanCandidate == cleanParent || strings.HasPrefix(cleanCandidate, cleanParent+"/")
}

func (s *Server) validateCloudPath(value string) error {
	clean := path.Clean("/" + strings.TrimLeft(strings.TrimSpace(value), "/"))
	root := path.Clean("/" + strings.TrimLeft(s.applicationRoot, "/"))
	if clean != root && !strings.HasPrefix(clean, root+"/") { return errors.New("path must remain inside the TBox Drive application folder") }
	if strings.Contains(value, "\\") { return errors.New("invalid cloud path") }
	return nil
}
func validateFileName(name string) error {
	if strings.TrimSpace(name)=="" || name=="." || name==".." || strings.ContainsAny(name,"/\\") { return errors.New("file name must be a single non-empty path component") }
	return nil
}
func parsePage(value string) int { n,err:=strconv.Atoi(value); if err!=nil || n<1 { return 1 }; return n }
func randomID() (string,error) { var b [16]byte; if _,err:=rand.Read(b[:]); err!=nil{return "",err}; return "web-"+hex.EncodeToString(b[:]),nil }
func decodeJSON(w http.ResponseWriter,r *http.Request,dst any) error {
	r.Body = http.MaxBytesReader(w,r.Body,1<<20)
	defer r.Body.Close()
	decoder:=json.NewDecoder(r.Body); decoder.DisallowUnknownFields()
	if err:=decoder.Decode(dst); err!=nil{return fmt.Errorf("invalid request body: %w",err)}
	if decoder.Decode(&struct{}{}) != io.EOF { return errors.New("request body must contain one JSON value") }
	return nil
}
func writeJSON(w http.ResponseWriter,status int,value any) { w.Header().Set("Content-Type","application/json; charset=utf-8"); w.WriteHeader(status); _=json.NewEncoder(w).Encode(value) }
func writeError(w http.ResponseWriter,err error) {
	status:=http.StatusBadGateway
	if errors.Is(err,context.Canceled) { status=499 } else if errors.Is(err,context.DeadlineExceeded) { status=http.StatusGatewayTimeout }
	writeJSON(w,status,map[string]string{"error":err.Error()})
}

// ListenAndServe starts the local-only browser service.
func (s *Server) ListenAndServe(ctx context.Context, address string) error {
	if address=="" { address="127.0.0.1:8765" }
	host,_,err:=net.SplitHostPort(address); if err!=nil{return fmt.Errorf("invalid web listen address: %w",err)}
	ip:=net.ParseIP(host); if ip==nil || !ip.IsLoopback() { return errors.New("web server may only bind to a loopback IP address") }
	server:=&http.Server{Addr:address,Handler:s.Handler(),ReadHeaderTimeout:5*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:1<<20}
	errCh:=make(chan error,1)
	go func(){errCh<-server.ListenAndServe()}()
	select {
	case <-ctx.Done():
		shutdownCtx,cancel:=context.WithTimeout(context.Background(),5*time.Second); defer cancel()
		if err:=server.Shutdown(shutdownCtx); err!=nil{return fmt.Errorf("shut down web server: %w",err)}
		err:=<-errCh; if errors.Is(err,http.ErrServerClosed){return nil}; return err
	case err:=<-errCh:
		if errors.Is(err,http.ErrServerClosed){return nil}; return err
	}
}

