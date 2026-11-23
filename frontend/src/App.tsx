import { useCallback, useEffect, useState } from "react";

type AppStatus = { ready: boolean; message: string; applicationRoot?: string };
type Account = { id: string; providerID: string; displayName: string; credentialRef: string };
type CloudFile = { id: string; name: string; path: string; size: number; isDirectory: boolean; modifiedAt?: string };
type DeviceAuthorization = { device_code: string; qrcode_url: string; expires_in: number; interval: number; user_code?: string; verification_url?: string };
type FilePage = { files: CloudFile[]; nextPageToken?: string };
type Quota = { total: number; used: number };
type Transfer = { id: string; provider_id: string; direction: string; source: string; destination: string; size: number; transferred: number; status: string; error?: string };
type UploadResult = { taskID: string; file: CloudFile };
type DriveStatus = { mounted: boolean; mountpoint?: string; provider?: string; mode?: string; message: string };

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
      SelectUploadFiles: () => Promise<string[]>;
      UploadFile: (accountID: string, localPath: string, parentPath: string, name: string, taskID: string) => Promise<UploadResult>;
      DownloadFile: (accountID: string, fileID: string, fileName: string) => Promise<string>;
      ListTransfers: (limit: number) => Promise<Transfer[]>;
      GetVirtualDriveStatus: () => Promise<DriveStatus>;
      MountVirtualDrive: (accountID: string, mountpoint: string) => Promise<DriveStatus>;
      UnmountVirtualDrive: () => Promise<DriveStatus>;
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
  const [activeView, setActiveView] = useState<"files" | "transfers" | "drive">("files");
  const [driveStatus, setDriveStatus] = useState<DriveStatus>({ mounted: false, message: "Virtual drive is not mounted." });
  const [driveLetter, setDriveLetter] = useState("T:");
  const [searchTerm, setSearchTerm] = useState("");
  const [nextPageToken, setNextPageToken] = useState("");
  const [theme, setTheme] = useState<"system" | "light" | "dark">(() => {
    const saved = window.localStorage.getItem("tbox-theme");
    return saved === "light" || saved === "dark" ? saved : "system";
  });

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
      setNextPageToken(page.nextPageToken ?? "");
      const q = await api.GetStorageQuota(id);
      setQuota(q);
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  }, [accountID, currentPath]);

  const loadMoreFiles = async () => {
    const api = desktop();
    if (!api || !accountID || !nextPageToken || busy) return;
    setBusy(true);
    try {
      const page = await api.ListFiles(accountID, currentPath, nextPageToken);
      setFiles(previous => [...previous, ...(page.files ?? [])]);
      setNextPageToken(page.nextPageToken ?? "");
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const refreshTransfers = useCallback(async () => {
    const api = desktop();
    if (!api) return;
    try { setTransfers(await api.ListTransfers(50)); } catch (e) { setMessage(String(e)); }
  }, []);

  useEffect(() => {
    const api = desktop();
    if (!api) { setStatus({ ready: false, message: "Run the TBox Drive desktop build to connect cloud accounts." }); return; }
    api.GetAppStatus().then(setStatus).catch(e => setMessage(String(e)));
    api.GetVirtualDriveStatus().then(setDriveStatus).catch(e => setMessage(String(e)));
    api.ListAccounts().then(list => {
      setAccounts(list);
      if (list.length) setAccountID(list[0].id);
    }).catch(e => setMessage(String(e)));
    const off = window.runtime?.EventsOn("transfer-progress", payload => setProgress({ transferred: payload.transferred, total: payload.total }));
    return () => off?.();
  }, []);

  useEffect(() => { if (accountID) refreshFiles(accountID, currentPath); }, [accountID, currentPath, refreshFiles]);
  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: light)");
    const applyTheme = () => {
      const selected = theme === "system" ? (media.matches ? "light" : "dark") : theme;
      document.documentElement.dataset.theme = selected;
    };
    applyTheme();
    window.localStorage.setItem("tbox-theme", theme);
    media.addEventListener("change", applyTheme);
    return () => media.removeEventListener("change", applyTheme);
  }, [theme]);
  useEffect(() => { if (activeView === "transfers") refreshTransfers(); }, [activeView, refreshTransfers]);

  const mountDrive = async () => {
    const api = desktop();
    if (!api || !accountID) return;
    setBusy(true);
    try {
      const result = await api.MountVirtualDrive(accountID, driveLetter.trim().toUpperCase());
      setDriveStatus(result); setMessage(result.message);
    } catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

  const unmountDrive = async () => {
    const api = desktop();
    if (!api) return;
    setBusy(true);
    try { const result = await api.UnmountVirtualDrive(); setDriveStatus(result); setMessage(result.message); }
    catch (e) { setMessage(String(e)); }
    finally { setBusy(false); }
  };

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
      const localPaths = await api.SelectUploadFiles();
      if (!localPaths.length) return;
      const failed: string[] = [];
      let completed = 0;
      for (const localPath of localPaths) {
        const name = localPath.split(/[\\\\/]/).pop() ?? "upload";
        setMessage(`Uploading ${completed + 1} of ${localPaths.length}: ${name}`);
        try {
          const result = await api.UploadFile(accountID, localPath, currentPath, name, "");
          completed += 1;
          setMessage(`Uploaded ${completed} of ${localPaths.length}: ${result.file.name}`);
        } catch (error) {
          failed.push(`${name}: ${String(error)}`);
        }
      }
      setMessage(failed.length ? `Uploaded ${completed} of ${localPaths.length}. Failed: ${failed.join("; ")}` : `Successfully uploaded ${completed} file(s).`);
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
          <button className={activeView === "drive" ? "nav-item active" : "nav-item"} onClick={() => { setActiveView("drive"); desktop()?.GetVirtualDriveStatus().then(setDriveStatus).catch(e => setMessage(String(e))); }}>Virtual Drive</button>
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
          <div><p className="eyebrow">CLOUD WORKSPACE</p><h1>{activeView === "files" ? "My Files" : activeView === "transfers" ? "Transfers" : "Virtual Drive"}</h1></div>
          <div className="header-actions">
            {accountID && activeView === "files" && <button onClick={upload} disabled={busy}>Upload file</button>}
            <button className="secondary" onClick={() => setTheme(value => value === "system" ? "light" : value === "light" ? "dark" : "system")} aria-label="Change appearance">Theme: {theme}</button>
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
                <input className="search-input" type="search" value={searchTerm} onChange={event => setSearchTerm(event.target.value)} placeholder="Search this folder…" aria-label="Search current folder" />
                {currentPath && <button className="secondary" onClick={parentFolder}>Up one level</button>}
                <button className="secondary" onClick={() => refreshFiles()} disabled={busy}>Refresh</button>
                {accountID && <button className="secondary danger" onClick={disconnect} disabled={busy}>Disconnect</button>}
              </div>
            </div>
            {!accountID ? <div className="empty"><div className="empty-icon">☁</div><h3>No cloud account connected</h3><p>Connect TeraBox to browse and transfer your files.</p></div> :
              <div className="file-list">
                <div className="file-row file-heading"><span>Name</span><span>Size</span><span>Modified</span><span>Action</span></div>
                {files.filter(file => file.name.toLowerCase().includes(searchTerm.trim().toLowerCase())).map(file => <div className="file-row" key={file.id}>
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
                {!files.filter(file => file.name.toLowerCase().includes(searchTerm.trim().toLowerCase())).length && <div className="empty compact"><h3>{busy ? "Loading files…" : searchTerm ? "No matching files" : "This folder is empty"}</h3><p>{searchTerm ? "Try a different search term." : "Files from your TeraBox application folder will appear here."}</p></div>}
                {nextPageToken && <div className="load-more"><button className="secondary" onClick={loadMoreFiles} disabled={busy}>Load more files</button></div>}
              </div>}
          </article>
        </>}

        {activeView === "drive" && <article className="panel drive-panel">
          <div className="panel-head">
            <div><p className="eyebrow">WINDOWS FILE EXPLORER INTEGRATION</p><h2>TeraBox Virtual Drive</h2></div>
            <span className={driveStatus.mounted ? "drive-pill mounted" : "drive-pill"}>{driveStatus.mounted ? "Mounted" : "Not mounted"}</span>
          </div>
          <p className="drive-description">Mount your connected TeraBox application folder as a Windows drive letter. Open files from File Explorer; the prototype downloads them to a local cache on demand and uploads edited files when they are closed.</p>
          <div className="drive-status-grid">
            <div><small>Drive letter</small><strong>{driveStatus.mountpoint || driveLetter}</strong></div>
            <div><small>Provider</small><strong>{driveStatus.provider || "TeraBox"}</strong></div>
            <div><small>Mode</small><strong>{driveStatus.mode || "Hybrid prototype"}</strong></div>
          </div>
          <div className="drive-controls">
            <label htmlFor="drive-letter">Mount as</label>
            <input id="drive-letter" className="search-input drive-letter" value={driveLetter} maxLength={2} onChange={event => setDriveLetter(event.target.value.toUpperCase())} disabled={driveStatus.mounted} />
            {!driveStatus.mounted
              ? <button onClick={mountDrive} disabled={busy || !accountID}>Mount drive</button>
              : <button className="secondary danger" onClick={unmountDrive} disabled={busy}>Unmount drive</button>}
          </div>
          <div className="notice drive-notice"><strong>Requirement</strong><span>Install WinFsp on Windows before mounting. A real TeraBox Open Platform account and application configuration are also required.</span></div>
          <p className="drive-message">{driveStatus.message}</p>
        </article>}

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
