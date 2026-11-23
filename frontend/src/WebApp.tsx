import { useCallback, useEffect, useRef, useState } from "react";

type Account = { id: string; providerID: string; displayName: string };
type CloudFile = { id: string; name: string; path: string; size: number; isDirectory: boolean; modifiedAt?: string };
type DeviceAuthorization = { device_code: string; qrcode_url: string; expires_in: number; interval: number; user_code?: string; verification_url?: string };
type Transfer = { id: string; direction: string; source: string; destination: string; size: number; transferred: number; status: string; error?: string };
type FilePage = { files: CloudFile[]; nextPageToken?: string };
type Quota = { total: number; used: number };

async function api<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, { credentials: "same-origin", ...init });
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try { const body = await response.json(); message = body.error || message; } catch { /* non-JSON error response */ }
    throw new Error(message);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
const json = (value: unknown): RequestInit => ({ method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(value) });
const bytes = (value: number) => {
  if (!value) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
};

export default function WebApp() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [accountID, setAccountID] = useState("");
  const [root, setRoot] = useState("");
  const [files, setFiles] = useState<CloudFile[]>([]);
  const [nextPageToken, setNextPageToken] = useState("");
  const [uploadProgress, setUploadProgress] = useState<{ name: string; loaded: number; total: number; stage: string } | null>(null);
  const [currentPath, setCurrentPath] = useState("");
  const [quota, setQuota] = useState<Quota | null>(null);
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [device, setDevice] = useState<DeviceAuthorization | null>(null);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [search, setSearch] = useState("");
  const [view, setView] = useState<"files" | "transfers" | "settings">("files");
  const fileInput = useRef<HTMLInputElement>(null);
  const connectAbort = useRef<AbortController | null>(null);
  const uploadRequest = useRef<XMLHttpRequest | null>(null);
  const fileRequestGeneration = useRef(0);

  const refreshAccounts = useCallback(async () => {
    const result = await api<Account[]>("/api/accounts");
    setAccounts(result);
    setAccountID(current => current && result.some(item => item.id === current) ? current : result[0]?.id ?? "");
  }, []);
  const refreshFiles = useCallback(async (id = accountID, folder = currentPath) => {
    const generation = ++fileRequestGeneration.current;
    if (!id) { setFiles([]); setNextPageToken(""); setQuota(null); return; }
    const query = new URLSearchParams({ account_id: id, path: folder || root, page: "1" });
    const [page, storage] = await Promise.all([api<FilePage>(`/api/files?${query}`), api<Quota>(`/api/quota?account_id=${encodeURIComponent(id)}`)]);
    if (generation !== fileRequestGeneration.current) return;
    setFiles(page.files || []);
    setNextPageToken(page.nextPageToken || "");
    setQuota(storage);
  }, [accountID, currentPath, root]);
  const refreshTransfers = useCallback(async () => setTransfers(await api<Transfer[]>("/api/transfers")), []);

  const loadMoreFiles = async () => {
    if (!accountID || !nextPageToken || busy) return;
    const generation = fileRequestGeneration.current;
    const requestedAccount = accountID;
    const requestedPath = currentPath;
    const requestedToken = nextPageToken;
    setBusy(true);
    try {
      const query = new URLSearchParams({ account_id: requestedAccount, path: requestedPath || root, page: requestedToken });
      const page = await api<FilePage>(`/api/files?${query}`);
      if (generation !== fileRequestGeneration.current || requestedAccount !== accountID || requestedPath !== currentPath) return;
      setFiles(previous => {
        const known = new Set(previous.map(file => file.id || file.path));
        return [...previous, ...(page.files || []).filter(file => !known.has(file.id || file.path))];
      });
      setNextPageToken(page.nextPageToken || "");
    } catch (error) {
      setMessage(String(error));
    } finally {
      if (generation === fileRequestGeneration.current) setBusy(false);
    }
  };

  useEffect(() => {
    const bootstrap = async () => {
      const launchToken = window.location.hash.slice(1);
      if (launchToken) {
        const response = await fetch("/api/session", {
          method: "POST",
          credentials: "same-origin",
          headers: { "X-TBox-Session": launchToken },
        });
        if (!response.ok) throw new Error("Invalid local launch URL. Restart TBox Drive and use the URL printed in its console.");
        window.history.replaceState(null, "", window.location.pathname + window.location.search);
      }
      const status = await api<{ applicationRoot: string }>("/api/status");
      setRoot(status.applicationRoot);
      await Promise.all([refreshAccounts(), refreshTransfers()]);
    };
    void bootstrap().catch(error => setMessage(`Local browser session unavailable. Use the private launch URL printed by TBox Drive. ${String(error)}`));
  }, [refreshAccounts, refreshTransfers]);
  useEffect(() => { if (accountID) void refreshFiles(accountID, currentPath).catch(error => setMessage(String(error))); }, [accountID, currentPath, refreshFiles]);
  useEffect(() => { const timer = window.setInterval(() => { void refreshTransfers().catch(() => undefined); }, 5000); return () => window.clearInterval(timer); }, [refreshTransfers]);
  const uploadActive = uploadProgress !== null;
  useEffect(() => {
    if (!uploadActive) return;
    const timer = window.setInterval(() => { void refreshTransfers().catch(() => undefined); }, 1000);
    return () => window.clearInterval(timer);
  }, [uploadActive, refreshTransfers]);
  useEffect(() => () => {
    connectAbort.current?.abort();
    uploadRequest.current?.abort();
  }, []);

  const connect = async () => {
    setBusy(true); setMessage("");
    try {
      const auth = await api<DeviceAuthorization>("/api/connect/start", json({}));
      setDevice(auth);
      setMessage("Complete authorization in TeraBox. Keep this page open while the connection is verified.");
      const controller = new AbortController();
      connectAbort.current = controller;
      const account = await api<Account>("/api/connect/complete", { ...json(auth), signal: controller.signal });
      setDevice(null); setMessage(`Connected ${account.displayName}`);
      await refreshAccounts();
      setAccountID(account.id);
    } catch (error) {
      if (!(error instanceof DOMException && error.name === "AbortError")) setMessage(String(error));
      setDevice(null);
    } finally { setBusy(false); }
  };

  const uploadOne = (file: File): Promise<void> => new Promise((resolve, reject) => {
    const data = new FormData();
    data.append("account_id", accountID);
    data.append("parent_path", currentPath || root);
    data.append("name", file.name);
    data.append("file", file, file.name);
    const request = new XMLHttpRequest();
    uploadRequest.current = request;
    request.open("POST", "/api/upload");
    request.withCredentials = true;
    request.upload.onprogress = event => {
      if (event.lengthComputable) {
        setUploadProgress({ name: file.name, loaded: event.loaded, total: event.total, stage: "Sending to local TBox Drive service" });
      }
    };
    request.upload.onload = () => {
      setUploadProgress({ name: file.name, loaded: file.size, total: file.size, stage: "File received locally; TeraBox upload is being finalized" });
    };
    request.onload = () => {
      uploadRequest.current = null;
      if (request.status >= 200 && request.status < 300) resolve();
      else {
        let reason = `Upload failed (HTTP ${request.status})`;
        try { reason = JSON.parse(request.responseText).error || reason; } catch { /* keep HTTP status */ }
        reject(new Error(reason));
      }
    };
    request.onerror = () => { uploadRequest.current = null; reject(new Error("Network error while sending file to the local service")); };
    request.onabort = () => { uploadRequest.current = null; reject(new Error("Upload cancelled")); };
    request.send(data);
  });

  const uploadFiles = async (selected: FileList | null) => {
    if (!selected?.length || !accountID || busy) return;
    setBusy(true);
    setMessage("");
    setUploadProgress(null);
    let complete = 0;
    const failures: string[] = [];
    const filesToUpload = Array.from(selected);
    try {
      for (const file of filesToUpload) {
        setUploadProgress({ name: file.name, loaded: 0, total: file.size, stage: "Preparing upload" });
        try {
          await uploadOne(file);
          complete += 1;
          setMessage(`Completed ${complete} of ${filesToUpload.length}: ${file.name}`);
          await refreshTransfers();
        } catch (error) {
          failures.push(`${file.name}: ${String(error)}`);
          if (String(error).includes("Upload cancelled")) break;
        }
      }
      setMessage(failures.length ? `Completed ${complete} of ${filesToUpload.length}. ${failures.join("; ")}` : `Successfully uploaded ${complete} file(s).`);
    } finally {
      uploadRequest.current = null;
      setUploadProgress(null);
      if (fileInput.current) fileInput.current.value = "";
      try { await Promise.all([refreshFiles(), refreshTransfers()]); } catch (error) { setMessage(String(error)); }
      setBusy(false);
    }
  };

  const cancelUpload = () => uploadRequest.current?.abort();

  const download = async (file: CloudFile) => {
    const query = new URLSearchParams({ account_id: accountID, file_id: file.id, name: file.name });
    const anchor = document.createElement("a");
    anchor.href = `/api/download?${query}`;
    anchor.download = file.name;
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    setMessage(`Download started: ${file.name}`);
  };
  const rename = async (file: CloudFile) => {
    const name = window.prompt("Rename item", file.name);
    if (!name?.trim() || name.trim() === file.name) return;
    setBusy(true);
    try { await api("/api/files/rename", json({ accountID, path: file.path, name: name.trim() })); await refreshFiles(); setMessage("File renamed."); }
    catch (error) { setMessage(String(error)); }
    finally { setBusy(false); }
  };
  const move = async (file: CloudFile) => {
    const destination = window.prompt("Move to folder path (inside your TBox Drive application folder)", root);
    if (!destination?.trim()) return;
    const normalized = destination.trim().replace(/\\/g, "/").replace(/\/$/, "");
    const source = file.path.replace(/\/$/, "");
    const sourceParent = source.slice(0, source.lastIndexOf("/")) || root;
    if (normalized === sourceParent) {
      setMessage("This item is already in that folder.");
      return;
    }
    if (normalized === source || normalized.startsWith(source + "/")) {
      setMessage("A folder cannot be moved into itself or one of its descendants.");
      return;
    }
    setBusy(true);
    try {
      await api("/api/files/move", json({ accountID, path: file.path, destination: normalized }));
      await refreshFiles();
      setMessage(`Moved ${file.name}.`);
    } catch (error) { setMessage(String(error)); }
    finally { setBusy(false); }
  };

  const remove = async (file: CloudFile) => {
    if (!window.confirm(`Delete "${file.name}" from TeraBox?`)) return;
    setBusy(true);
    try { await api("/api/files/delete", json({ accountID, path: file.path })); await refreshFiles(); setMessage("File deleted."); }
    catch (error) { setMessage(String(error)); }
    finally { setBusy(false); }
  };
  const disconnect = async () => {
    if (!accountID || !window.confirm("Disconnect this TeraBox account from TBox Drive?")) return;
    setBusy(true);
    try { await api("/api/accounts/disconnect", json({ accountID })); await refreshAccounts(); setCurrentPath(""); setMessage("Account disconnected."); }
    catch (error) { setMessage(String(error)); }
    finally { setBusy(false); }
  };
  const parent = () => {
    if (!currentPath || currentPath === root) return;
    const parts = currentPath.replace(/\/$/, "").split("/");
    parts.pop();
    const next = parts.join("/") || root;
    setCurrentPath(next);
  };
  const filtered = files.filter(file => file.name.toLowerCase().includes(search.trim().toLowerCase()));

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-mark">T</span><span>TBox Drive</span></div>
        <nav>
          <button className={view === "files" ? "nav-item active" : "nav-item"} onClick={() => setView("files")}>My Files</button>
          <button className={view === "transfers" ? "nav-item active" : "nav-item"} onClick={() => setView("transfers")}>Transfers</button>
          <button className={view === "settings" ? "nav-item active" : "nav-item"} onClick={() => setView("settings")}>Settings</button>
        </nav>
        <div className="provider"><span className="dot" /> TeraBox <small>{accounts.length ? "Connected account" : "Not connected"}</small>
          <select value={accountID} onChange={event => { setAccountID(event.target.value); setCurrentPath(""); }} disabled={!accounts.length}>
            {accounts.length ? accounts.map(item => <option key={item.id} value={item.id}>{item.displayName}</option>) : <option value="">No account connected</option>}
          </select>
        </div>
      </aside>
      <section className="content">
        <header><div><p className="eyebrow">SELF-HOSTED CLOUD STORAGE</p><h1>{view === "files" ? "My Files" : view === "transfers" ? "Transfers" : "Settings"}</h1></div>
          {view === "files" && <div className="header-actions"><input ref={fileInput} type="file" multiple hidden onChange={event => void uploadFiles(event.target.files)} /><button disabled={!accountID || busy} onClick={() => fileInput.current?.click()}>Upload files</button></div>}
        </header>
        {!accounts.length && <article className="connect-panel"><div><p className="eyebrow">CONNECT YOUR CLOUD</p><h2>Connect a TeraBox account</h2><p>Authorize using TeraBox device login. Your account tokens stay in the operating-system credential vault and are never sent to browser JavaScript.</p><button disabled={busy} onClick={() => void connect()}>Connect TeraBox</button></div>
          {device && <div><p className="user-code">Code: {device.user_code || "Use the QR code"}</p>{device.qrcode_url && <img className="qr" src={device.qrcode_url.startsWith("data:") || device.qrcode_url.startsWith("http") ? device.qrcode_url : `data:image/png;base64,${device.qrcode_url}`} alt="TeraBox authorization QR code" />}{device.verification_url && <p><a href={device.verification_url} target="_blank" rel="noreferrer">Open TeraBox authorization</a></p>}</div>}
        </article>}
        {message && <div className="message" role="status">{message}</div>}
        {view === "files" && <>
          <div className="cards">
            <article><p>Cloud storage used</p><strong>{quota ? bytes(quota.used) : "—"}</strong><small>{quota ? `of ${bytes(quota.total)}` : "Connect an account to view quota"}</small></article>
            <article><p>Files in this folder</p><strong>{files.length}</strong><small>Current directory</small></article>
            <article><p>Recent transfers</p><strong>{transfers.length}</strong><small>Saved transfer records</small></article>
          </div>
          {accounts.length > 0 && <article className="panel">
            <div className="panel-head"><div><p className="eyebrow">TERABOX STORAGE</p><h2>{currentPath || root || "Loading folder…"}</h2></div><div className="panel-actions"><button className="secondary" onClick={parent} disabled={!currentPath || currentPath === root}>Up</button><button className="secondary" onClick={() => void refreshFiles()} disabled={busy}>Refresh</button></div></div>
            <input className="search-input" placeholder="Search this folder" value={search} onChange={event => setSearch(event.target.value)} />
            {filtered.length ? <div className="file-list">
              <div className="file-row file-heading"><span>Name</span><span>Size</span><span>Type</span><span>Actions</span></div>
              {filtered.map(file => <div className="file-row" key={file.id}>
                <span><button className="file-name" disabled={!file.isDirectory} onClick={() => setCurrentPath(file.path)}><span className="file-icon">{file.isDirectory ? "▰" : "▤"}</span>{file.name}</button></span>
                <span>{file.isDirectory ? "—" : bytes(file.size)}</span><span>{file.isDirectory ? "Folder" : "File"}</span>
                <span className="row-actions">{!file.isDirectory && <button className="text-button" disabled={busy} onClick={() => void download(file)}>Download</button>}<button className="text-button" disabled={busy} onClick={() => void rename(file)}>Rename</button><button className="text-button" disabled={busy} onClick={() => void move(file)}>Move</button><button className="text-button danger-text" disabled={busy} onClick={() => void remove(file)}>Delete</button></span>
              </div>)}
            </div> : <div className="empty compact"><h3>{busy ? "Loading…" : "This folder is empty"}</h3><p>Upload files or open another folder.</p></div>}
            {nextPageToken && <div className="load-more"><button className="secondary" disabled={busy} onClick={() => void loadMoreFiles()}>{busy ? "Loading…" : "Load more files"}</button></div>}
            {uploadProgress && <div className="upload-progress"><div className="progress-label"><span>{uploadProgress.name}</span><span>{uploadProgress.total ? Math.round(uploadProgress.loaded / uploadProgress.total * 100) : 0}%</span></div><div className="progress"><span style={{ width: `${uploadProgress.total ? Math.min(100, uploadProgress.loaded / uploadProgress.total * 100) : 0}%` }} /></div><small>{uploadProgress.stage} · {bytes(uploadProgress.loaded)} / {bytes(uploadProgress.total)}</small><button className="secondary danger" disabled={!uploadRequest.current} onClick={cancelUpload}>Cancel current upload</button></div>}
          </article>}
        </>}
        {view === "transfers" && <article className="panel"><div className="panel-head"><div><p className="eyebrow">TRANSFER CENTER</p><h2>Recent activity</h2></div><button className="secondary" onClick={() => void refreshTransfers()}>Refresh</button></div>
          {transfers.length ? <div className="file-list"><div className="file-row file-heading"><span>File</span><span>Progress</span><span>Status</span><span>Direction</span></div>{transfers.map(item => <div className="file-row" key={item.id}><span>{item.destination.split("/").pop() || item.source.split(/[\\/]/).pop()}</span><span>{bytes(item.transferred)} / {bytes(item.size)}{item.size > 0 && <div className="transfer-mini-progress"><span style={{ width: `${Math.min(100, item.transferred / item.size * 100)}%` }} /></div>}</span><span>{item.status}{item.error ? <small className="transfer-error">{item.error}</small> : null}</span><span>{item.direction}</span></div>)}</div> : <div className="empty"><h3>No transfers yet</h3><p>Uploads and downloads will appear here.</p></div>}
        </article>}
        {view === "settings" && <article className="panel settings-panel"><p className="eyebrow">LOCAL SECURITY</p><h2>Self-hosted mode</h2><div className="security-callout"><strong>Loopback-only access</strong><span>The web service listens on 127.0.0.1. Remote/LAN access is disabled. TeraBox credentials remain in your operating-system credential vault.</span></div><div className="settings-row"><div><strong>Connected account</strong><small>{accounts[0]?.displayName || "No account connected"}</small></div>{accountID && <button className="secondary danger" disabled={busy} onClick={() => void disconnect()}>Disconnect</button>}</div><div className="settings-row"><div><strong>Application folder</strong><small>{root || "Not configured"}</small></div></div></article>}
      </section>
    </main>
  );
}
