package provider

type Capabilities struct {
	MultipartUpload bool
	ResumableUpload bool
	ParallelUpload bool
	StreamingDownload bool
	ServerSideCopy bool
	FolderManagement bool
	RemoteChecksum bool
}
