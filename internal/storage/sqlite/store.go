package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	_ "modernc.org/sqlite"
)

type Store struct { db *sql.DB }
type Part = transfer.Part

var _ transfer.Repository = (*Store)(nil)

func Open(path string)(*Store,error){
	if path==""{return nil,errors.New("database path is required")}
	if path!=":memory:"{dir:=filepath.Dir(path);if err:=os.MkdirAll(dir,0700);err!=nil{return nil,fmt.Errorf("create database directory: %w",err)}}
	db,err:=sql.Open("sqlite",path);if err!=nil{return nil,fmt.Errorf("open SQLite database: %w",err)}
	db.SetMaxOpenConns(1);s:=&Store{db:db}
	if err:=db.Ping();err!=nil{db.Close();return nil,fmt.Errorf("ping SQLite database: %w",err)}
	for _,pragma:=range []string{"PRAGMA foreign_keys = ON","PRAGMA journal_mode = WAL","PRAGMA busy_timeout = 5000"}{if _,err:=db.Exec(pragma);err!=nil{db.Close();return nil,fmt.Errorf("configure SQLite: %w",err)}}
	if err:=s.Migrate(context.Background());err!=nil{db.Close();return nil,err};return s,nil
}
func(s *Store)Close()error{return s.db.Close()}
func(s *Store)Migrate(ctx context.Context)error{
	const schema=`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS accounts(id TEXT PRIMARY KEY,provider_id TEXT NOT NULL,display_name TEXT NOT NULL DEFAULT '',credential_ref TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS transfers(id TEXT PRIMARY KEY,provider_id TEXT NOT NULL,direction TEXT NOT NULL,source TEXT NOT NULL,destination TEXT NOT NULL,size INTEGER NOT NULL CHECK(size>=0),transferred INTEGER NOT NULL DEFAULT 0 CHECK(transferred>=0 AND transferred<=size),status TEXT NOT NULL,error TEXT NOT NULL DEFAULT '',created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_transfers_status_updated ON transfers(status,updated_at);
CREATE TABLE IF NOT EXISTS transfer_parts(transfer_id TEXT NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,part_index INTEGER NOT NULL CHECK(part_index>=0),byte_offset INTEGER NOT NULL CHECK(byte_offset>=0),byte_size INTEGER NOT NULL CHECK(byte_size>=0),checksum TEXT NOT NULL DEFAULT '',remote_part_id TEXT NOT NULL DEFAULT '',status TEXT NOT NULL,updated_at TEXT NOT NULL,PRIMARY KEY(transfer_id,part_index));
CREATE TABLE IF NOT EXISTS provider_sessions(transfer_id TEXT PRIMARY KEY REFERENCES transfers(id) ON DELETE CASCADE,session_ref TEXT NOT NULL,expires_at TEXT,updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS transfer_events(id INTEGER PRIMARY KEY AUTOINCREMENT,transfer_id TEXT NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,event_type TEXT NOT NULL,created_at TEXT NOT NULL);`
	tx,err:=s.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
	if _,err:=tx.ExecContext(ctx,schema);err!=nil{return fmt.Errorf("apply SQLite schema: %w",err)}
	if _,err:=tx.ExecContext(ctx,"INSERT OR IGNORE INTO schema_migrations(version,applied_at) VALUES(1,?)",time.Now().UTC().Format(time.RFC3339Nano));err!=nil{return err};return tx.Commit()
}
func(s *Store)CreateTransfer(ctx context.Context,t transfer.Task)error{
	if t.ID==""||t.ProviderID==""||t.Direction==""{return errors.New("transfer ID, provider and direction are required")};if t.Size<0||t.Transferred<0||t.Transferred>t.Size{return errors.New("invalid transfer byte counts")}
	now:=time.Now().UTC();if t.CreatedAt.IsZero(){t.CreatedAt=now};if t.UpdatedAt.IsZero(){t.UpdatedAt=now};if t.Status==""{t.Status=transfer.StatusQueued}
	_,err:=s.db.ExecContext(ctx,`INSERT INTO transfers(id,provider_id,direction,source,destination,size,transferred,status,error,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,t.ID,t.ProviderID,t.Direction,t.Source,t.Destination,t.Size,t.Transferred,string(t.Status),t.Error,t.CreatedAt.UTC().Format(time.RFC3339Nano),t.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err!=nil{return fmt.Errorf("create transfer: %w",err)};return nil
}
func(s *Store)GetTransfer(ctx context.Context,id string)(transfer.Task,error){
	var t transfer.Task;var status,created,updated string
	err:=s.db.QueryRowContext(ctx,`SELECT id,provider_id,direction,source,destination,size,transferred,status,error,created_at,updated_at FROM transfers WHERE id=?`,id).Scan(&t.ID,&t.ProviderID,&t.Direction,&t.Source,&t.Destination,&t.Size,&t.Transferred,&status,&t.Error,&created,&updated)
	if err!=nil{return transfer.Task{},err};t.Status=transfer.Status(status);t.CreatedAt,_=time.Parse(time.RFC3339Nano,created);t.UpdatedAt,_=time.Parse(time.RFC3339Nano,updated);return t,nil
}
func(s *Store)ListTransfers(ctx context.Context,limit int)([]transfer.Task,error){
	if limit<=0||limit>1000{return nil,errors.New("transfer list limit must be between 1 and 1000")}
	rows,err:=s.db.QueryContext(ctx,`SELECT id,provider_id,direction,source,destination,size,transferred,status,error,created_at,updated_at FROM transfers ORDER BY updated_at DESC LIMIT ?`,limit);if err!=nil{return nil,err};defer rows.Close();items:=make([]transfer.Task,0)
	for rows.Next(){var t transfer.Task;var status,created,updated string;if err:=rows.Scan(&t.ID,&t.ProviderID,&t.Direction,&t.Source,&t.Destination,&t.Size,&t.Transferred,&status,&t.Error,&created,&updated);err!=nil{return nil,err};t.Status=transfer.Status(status);t.CreatedAt,_=time.Parse(time.RFC3339Nano,created);t.UpdatedAt,_=time.Parse(time.RFC3339Nano,updated);items=append(items,t)};return items,rows.Err()
}
func(s *Store)UpdateProgress(ctx context.Context,id string,transferred int64,status transfer.Status,errText string)error{
	if transferred<0{return errors.New("transferred bytes cannot be negative")}
	result,err:=s.db.ExecContext(ctx,`UPDATE transfers SET transferred=?,status=?,error=?,updated_at=? WHERE id=? AND size>=?`,transferred,string(status),errText,time.Now().UTC().Format(time.RFC3339Nano),id,transferred);if err!=nil{return fmt.Errorf("update transfer progress: %w",err)}
	n,err:=result.RowsAffected();if err!=nil{return err};if n==0{return errors.New("transfer not found or progress exceeds total size")};return nil
}
func(s *Store)SavePart(ctx context.Context,p transfer.Part)error{
	if p.TransferID==""||p.Index<0||p.Offset<0||p.Size<0||p.Status==""{return errors.New("invalid transfer part")};if p.UpdatedAt.IsZero(){p.UpdatedAt=time.Now().UTC()}
	_,err:=s.db.ExecContext(ctx,`INSERT INTO transfer_parts(transfer_id,part_index,byte_offset,byte_size,checksum,remote_part_id,status,updated_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(transfer_id,part_index) DO UPDATE SET byte_offset=excluded.byte_offset,byte_size=excluded.byte_size,checksum=excluded.checksum,remote_part_id=excluded.remote_part_id,status=excluded.status,updated_at=excluded.updated_at`,p.TransferID,p.Index,p.Offset,p.Size,p.Checksum,p.RemotePartID,p.Status,p.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err!=nil{return fmt.Errorf("save transfer part: %w",err)};return nil
}
func(s *Store)ListParts(ctx context.Context,transferID string)([]transfer.Part,error){
	rows,err:=s.db.QueryContext(ctx,`SELECT transfer_id,part_index,byte_offset,byte_size,checksum,remote_part_id,status,updated_at FROM transfer_parts WHERE transfer_id=? ORDER BY part_index`,transferID);if err!=nil{return nil,err};defer rows.Close();parts:=make([]transfer.Part,0)
	for rows.Next(){var p transfer.Part;var updated string;if err:=rows.Scan(&p.TransferID,&p.Index,&p.Offset,&p.Size,&p.Checksum,&p.RemotePartID,&p.Status,&updated);err!=nil{return nil,err};p.UpdatedAt,_=time.Parse(time.RFC3339Nano,updated);parts=append(parts,p)};return parts,rows.Err()
}
func(s *Store)SaveProviderSession(ctx context.Context,transferID,sessionRef string,expiresAt *time.Time)error{
	if transferID==""||sessionRef==""{return errors.New("transfer ID and provider session reference are required")};var expiry any;if expiresAt!=nil{expiry=expiresAt.UTC().Format(time.RFC3339Nano)}
	_,err:=s.db.ExecContext(ctx,`INSERT INTO provider_sessions(transfer_id,session_ref,expires_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(transfer_id) DO UPDATE SET session_ref=excluded.session_ref,expires_at=excluded.expires_at,updated_at=excluded.updated_at`,transferID,sessionRef,expiry,time.Now().UTC().Format(time.RFC3339Nano));return err
}
