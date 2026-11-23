import { useCallback, useEffect, useState } from "react";

type AppStatus = { ready: boolean; message: string; applicationRoot?: string };
type Account = { id: string; providerID: string; displayName: string; credentialRef: string };
type CloudFile = { id: string; name: string; path: string; size: number; isDirectory: boolean; modifiedAt?: string };
type DeviceAuthorization = { device_code: string; qrcode_url: string; expires_in: number; interval: number; user_code?: string; verification_url?: string };
type FilePage = { files: CloudFile[]; nextPageToken?: string };
type Quota = { total: number; used: number };
type Transfer = { id: string; provider_id: string; direction: string; source: string; destination: string; size: number; transferred: number; status: string; error?: string };
type UploadResult = { taskID: string; file: CloudFile };

declare global {
  interface Window {
    go?: { main?: { DesktopApp?: {
      GetAppStatus: () => Promise<AppStatus>;
      StartTeraBoxConnect: () => Promise<DeviceAuthorization>;
      CompleteTeraBoxConnect: (device: DeviceAuthorization) => Promise<Account>;
      ListAccounts: () => Promise<Account[]>;
      DisconnectAccount: (id: string) => Promise<void>;
      RenameFile: (accountID: string, remotePath: string, newName: string) => Promise<void>;
      DeleteFile: (accountID: string, remotePath: string) => Promise<void>;
      MoveFile: (accountID: string, remotePath: string, destinationDir: string, newName: string) => Promise<void>;
      ListFiles: (accountID: string, parentPath: string, pageToken: string) => Promise<FilePage>;
      GetStorageQuota: (accountID: string) => Promise<Quota>;
      SelectUploadFile: () => Promise<string>;
      UploadFile: (accountID: string, localPath: string, parentPath: string, name: string, taskID: string) => Promise<UploadResult>;
      DownloadFile: (accountID: string, fileID: string, fileName: string) => Promise<string>;
      ListTransfers: (limit: number) => Promise<Transfer[]>;
    } } };
    runtime?: { EventsOn: (event: string, callback: (payload: { transferId: string; transferred: number; total: number }) => void) => (() => void) };
  }
}

const desktop = () => window.go?.main?.DesktopApp;
const formatBytes = (value: number) => {
  if (!value) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
};

export default function App() {
  const [status, setStatus] = useState<AppStatus | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [accountID, setAccountID] = useState("");
  const [files, setFiles] = useState<CloudFile[]>([]);
  const [currentPath, setCurrentPath] = useState("");
  const [quota, setQuota] = useState<Quota | null>(null);
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [device, setDevice] = useState<DeviceAuthorization | null>(null);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [progress, setProgress] = useState<{ transferred: number; total: number } | null>(null);
  const [activeView, setActiveView] = useState<"files" | "transfers">("files");

  const refreshAccounts = useCallback(async () => {
    const api = desktop();
    if (!api) return;
    const list = await api.ListAccounts();
    setAccounts(list);
    if (!accountID && list.length) setAccountID(list[0].id);
  }, [accountID]);

  const refreshFiles = useCallback(async (id = accountID, folder = currentPath) => {
    const api = desktop();
    if (!api || !id) return;
    setBusy(true);
    try {
      const page = await api.ListFiles(id, folder, "");
      setFiles(page.files ?? []);
      const q = await api.GetStorageQuota(id);
      setQuota(q);
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  }, [accountID, currentPath]);

  const refreshTransfers = useCallback(async () => {
    const api = desktop();
    if (!api) return;
    try { setTransfers(await api.ListTransfers(50)); } catch (e) { setMessage(String(e)); }
  }, []);

  useEffect(() => {
    const api = desktop();
    if (!api) { setStatus({ ready: false, message: "Run the TBox Drive desktop build to connect cloud accounts." }); return; }
    api.GetAppStatus().then(setStatus).catch(e => setMessage(String(e)));
    api.ListAccounts().then(list => {
      setAccounts(list);
      if (list.length) setAccountID(list[0].id);
    }).catch(e => setMessage(String(e)));
    const off = window.runtime?.EventsOn("transfer-progress", payload => setProgress({ transferred: payload.transferred, total: payload.total }));
    return () => off?.();
  }, []);

  useEffect(() => { if (accountID) refreshFiles(accountID, currentPath); }, [accountID, currentPath, refreshFiles]);
  useEffect(() => { if (activeView === "transfers") refreshTransfers(); }, [activeView, refreshTransfers]);

  const connect = async () => {
    const api = desktop();
    if (!api) return;
    setBusy(true); setMessage("");
    try {
      const auth = await api.StartTeraBoxConnect();
      setDevice(auth);
      setMessage("Scan the QR code with your TeraBox account. Keep this window open while authorization completes.");
      void api.CompleteTeraBoxConnect(auth).then(async account => {
        setDevice(null); setMessage(`Connected ${account.displayName}`);
        await refreshAccounts(); setAccountID(account.id);
      }).catch(e => { setDevice(null); setMessage(String(e)); });
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const upload = async () => {
    const api = desktop();
    if (!api || !accountID) return;
    setBusy(true); setMessage(""); setProgress(null);
    try {
      const localPath = await api.SelectUploadFile();
      if (!localPath) return;
      const name = localPath.split(/[\\/]/).pop() ?? "upload";
      const result = await api.UploadFile(accountID, localPath, currentPath, name, "");
      setMessage(`Uploaded ${result.file.name}`);
      await refreshFiles(); await refreshTransfers();
    } catch (e) { setMessage(String(e)); await refreshTransfers(); }
    finally { setBusy(false); }
  };

  const rename = async (file: CloudFile) => {
    const api = desktop();
    if (!api || !accountID) return;
    const nextName = window.prompt("Rename item", file.name);
    if (!nextName || nextName.trim() === file.name) return;
    setBusy(true);
    try { await api.RenameFile(accountID, file.path, nextName.trim()); setMessage(`Renamed to ${nextName.trim()}`); await refreshFiles(); }
    catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const remove = async (file: CloudFile) => {
    const api = desktop();
    if (!api || !accountID || !window.confirm(`Delete "${file.name}" from TeraBox?`)) return;
    setBusy(true);
    try { await api.DeleteFile(accountID, file.path); setMessage(`Deleted ${file.name}`); await refreshFiles(); }
    catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const download = async (file: CloudFile) => {
    const api = desktop();
    if (!api || !accountID) return;
    setBusy(true);
    try {
      const saved = await api.DownloadFile(accountID, file.id, file.name);
      if (saved) setMessage(`Downloaded to ${saved}`);
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const disconnect = async () => {
    const api = desktop();
    if (!api || !accountID) return;
    setBusy(true);
    try {
      await api.DisconnectAccount(accountID);
      setFiles([]); setQuota(null); setCurrentPath(""); setAccountID("");
      await refreshAccounts(); setMessage("Account disconnected.");
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const parentFolder = () => {
    if (!currentPath || currentPath === status?.applicationRoot) return;
    const next = currentPath.replace(/\/$/, "").split("/").slice(0, -1).join("/") || status?.applicationRoot || "";
    setCurrentPath(next);
  };

  const qrSource = device?.qrcode_url?.startsWith("data:") || device?.qrcode_url?.startsWith("http")
    ? device.qrcode_url
    : device?.qrcode_url ? `data:image/png;base64,${device.qrcode_url}` : "";

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-mark">T</span><span>TBox Drive</span></div>
        <nav>
          <button className={activeView === "files" ? "nav-item active" : "nav-item"} onClick={() => setActiveView("files")}>My Files</button>
          <button className={activeView === "transfers" ? "nav-item active" : "nav-item"} onClick={() => setActiveView("transfers")}>Transfers</button>
        </nav>
        <div className="provider">
          <span className="dot" /> TeraBox
          <small>{accounts.length ? `${accounts.length} account(s) connected` : "Not connected"}</small>
          {accounts.length > 0 && <select value={accountID} onChange={e => { setAccountID(e.target.value); setCurrentPath(""); }} aria-label="Cloud account">
            {accounts.map(a => <option key={a.id} value={a.id}>{a.displayName}</option>)}
          </select>}
        </div>
      </aside>

      <section className="content">
        <header>
          <div><p className="eyebrow">CLOUD WORKSPACE</p><h1>{activeView === "files" ? "My Files" : "Transfers"}</h1></div>
          <div className="header-actions">
            {accountID && activeView === "files" && <button onClick={upload} disabled={busy}>Upload file</button>}
            {!accounts.length && <button onClick={connect} disabled={busy || !status?.ready}>Connect TeraBox</button>}
          </div>
        </header>

        {status && !status.ready && <div className="notice"><strong>Setup required</strong><span>{status.message}</span></div>}
        {message && <div className="message" role="status">{message}</div>}
        {device && <div className="connect-panel">
          <div><p className="eyebrow">TERABOX AUTHORIZATION</p><h2>Scan to connect your account</h2><p>Approve this device using your TeraBox account. This window will finish connecting automatically.</p>
            {device.user_code && <p className="user-code">Code: {device.user_code}</p>}
            {device.verification_url && <a href={device.verification_url} target="_blank" rel="noreferrer">Open verification page</a>}
          </div>
          {qrSource && <img className="qr" src={qrSource} alt="TeraBox authorization QR code" />}
        </div>}

        {activeView === "files" && <>
          <div className="cards">
            <article><p>Connected accounts</p><strong>{accounts.length}</strong><small>TeraBox cloud accounts</small></article>
            <article><p>Cloud storage used</p><strong>{quota ? formatBytes(quota.used) : "—"}</strong><small>{quota ? `of ${formatBytes(quota.total)} total` : "Connect an account to view quota"}</small></article>
            <article><p>Available storage</p><strong>{quota ? formatBytes(Math.max(0, quota.total - quota.used)) : "—"}</strong><small>Based on current account quota</small></article>
          </div>
          <article className="panel">
            <div className="panel-head">
              <div><p className="eyebrow">CLOUD FILE EXPLORER</p><h2>{currentPath ? currentPath.split("/").pop() : "Application folder"}</h2></div>
              <div className="panel-actions">
                {currentPath && <button className="secondary" onClick={parentFolder}>Up one level</button>}
                <button className="secondary" onClick={() => refreshFiles()} disabled={busy}>Refresh</button>
                {accountID && <button className="secondary danger" onClick={disconnect} disabled={busy}>Disconnect</button>}
              </div>
            </div>
            {!accountID ? <div className="empty"><div className="empty-icon">☁</div><h3>No cloud account connected</h3><p>Connect TeraBox to browse and transfer your files.</p></div> :
              <div className="file-list">
                <div className="file-row file-heading"><span>Name</span><span>Size</span><span>Modified</span><span>Action</span></div>
                {files.map(file => <div className="file-row" key={file.id}>
                  <button className="file-name" onClick={() => file.isDirectory && setCurrentPath(file.path)} disabled={!file.isDirectory}>
                    <span className="file-icon">{file.isDirectory ? "▰" : "▤"}</span>{file.name}
                  </button>
                  <span>{file.isDirectory ? "—" : formatBytes(file.size)}</span>
                  <span>{file.modifiedAt ? new Date(file.modifiedAt).toLocaleDateString() : "—"}</span>
                  <span className="row-actions">
                    {!file.isDirectory && <button className="text-button" onClick={() => download(file)} disabled={busy}>Download</button>}
                    <button className="text-button" onClick={() => rename(file)} disabled={busy}>Rename</button>
                    <button className="text-button danger-text" onClick={() => remove(file)} disabled={busy}>Delete</button>
                  </span>
                </div>)}
                {!files.length && <div className="empty compact"><h3>{busy ? "Loading files…" : "This folder is empty"}</h3><p>Files from your TeraBox application folder will appear here.</p></div>}
              </div>}
          </article>
        </>}

        {activeView === "transfers" && <article className="panel">
          <div className="panel-head"><div><p className="eyebrow">TRANSFER CENTER</p><h2>Recent activity</h2></div><button className="secondary" onClick={refreshTransfers}>Refresh</button></div>
          {transfers.length ? <div className="file-list">
            <div className="file-row file-heading"><span>File</span><span>Progress</span><span>Status</span><span>Direction</span></div>
            {transfers.map(t => <div className="file-row" key={t.id}>
              <span>{t.destination.split("/").pop() || t.source.split(/[\\/]/).pop()}</span>
              <span>{formatBytes(t.transferred)} / {formatBytes(t.size)}</span>
              <span className={t.status === "completed" ? "status-good" : "status-muted"}>{t.status}</span>
              <span>{t.direction}</span>
            </div>)}
          </div> : <div className="empty"><div className="empty-icon">⇅</div><h3>No transfers yet</h3><p>Uploads and downloads will appear here.</p></div>}
          {progress && <div className="progress-wrap"><div className="progress-label"><span>Current transfer</span><span>{formatBytes(progress.transferred)} / {formatBytes(progress.total)}</span></div><div className="progress"><span style={{ width: `${progress.total ? Math.min(100, progress.transferred / progress.total * 100) : 0}%` }} /></div></div>}
        </article>}
      </section>
    </main>
  );
}
