package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/account"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
)

type CredentialLocker interface {
	AcquireCredentialLock(context.Context, string, time.Duration) (func() error, error)
}
type SecretStore interface {
	Get(context.Context,string)([]byte,error)
	Set(context.Context,string,[]byte) error
	Delete(context.Context,string) error
}
type storedTeraBoxCredential struct {
	Token terabox.OAuthToken `json:"token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
	TokenInfo terabox.TokenInfo `json:"token_info"`
	TokenInfoFetchedAt time.Time `json:"token_info_fetched_at"`
}
type TeraBoxService struct {
	oauth *terabox.OAuthClient
	secrets SecretStore
	accounts account.Repository
	transfers transfer.Repository
	applicationRoot string
	now func()time.Time
	refreshMu sync.Mutex
	credentialLocker CredentialLocker
}
func NewTeraBoxService(oauth *terabox.OAuthClient,secrets SecretStore,accounts account.Repository,transfers transfer.Repository,applicationRoot string)(*TeraBoxService,error){
	if oauth==nil||secrets==nil||accounts==nil{return nil,errors.New("OAuth client, secret store and account repository are required")}
	if applicationRoot==""{return nil,errors.New("TeraBox application folder path is required")}
	locker, _ := accounts.(CredentialLocker)
	return &TeraBoxService{oauth:oauth,secrets:secrets,accounts:accounts,transfers:transfers,applicationRoot:applicationRoot,now:time.Now,credentialLocker:locker},nil
}
func(s *TeraBoxService)BeginConnect(ctx context.Context)(terabox.DeviceAuthorization,error){return s.oauth.BeginDeviceAuthorization(ctx)}
func(s *TeraBoxService)CompleteConnect(ctx context.Context,device terabox.DeviceAuthorization)(account.Account,error){
	token,err:=s.oauth.WaitForDeviceAuthorization(ctx,device);if err!=nil{return account.Account{},err}
	info,err:=s.oauth.TokenInformation(ctx,token.AccessToken);if err!=nil{return account.Account{},fmt.Errorf("load TeraBox account information: %w",err)}
	if info.UserID==0{return account.Account{},errors.New("TeraBox token information did not include a user ID")}
	id:=strconv.FormatInt(info.UserID,10);ref:="terabox:"+id;now:=s.now().UTC()
	credential:=storedTeraBoxCredential{Token:token,AccessExpiresAt:now.Add(time.Duration(token.ExpiresIn)*time.Second),TokenInfo:info,TokenInfoFetchedAt:now}
	encoded,err:=json.Marshal(credential);if err!=nil{return account.Account{},err}
	if err:=s.secrets.Set(ctx,ref,encoded);err!=nil{return account.Account{},fmt.Errorf("store TeraBox credentials in secure store: %w",err)}
	item:=account.Account{ID:"terabox:"+id,ProviderID:"terabox",DisplayName:"TeraBox account "+id,CredentialRef:ref,CreatedAt:now,UpdatedAt:now}
	if err:=s.accounts.Save(ctx,item);err!=nil{_ = s.secrets.Delete(ctx,ref);return account.Account{},fmt.Errorf("save TeraBox account: %w",err)}
	return item,nil
}
func(s *TeraBoxService)ListAccounts(ctx context.Context)([]account.Account,error){return s.accounts.List(ctx)}
func(s *TeraBoxService)NewAPIClient(ctx context.Context,accountID string)(*terabox.APIClient,error){
	// TeraBox refresh tokens are single-use. Serialize refresh and credential
	// persistence so concurrent API calls cannot consume the same token twice.
	s.refreshMu.Lock();defer s.refreshMu.Unlock()
	release,lockErr:=s.acquireCredentialLock(ctx,accountID);if lockErr!=nil{return nil,fmt.Errorf("lock TeraBox credentials: %w",lockErr)};defer func(){_ = release()}()
	item,err:=s.accounts.Get(ctx,accountID);if err!=nil{return nil,err};if item.ProviderID!="terabox"{return nil,errors.New("account is not a TeraBox account")}
	raw,err:=s.secrets.Get(ctx,item.CredentialRef);if err!=nil{return nil,fmt.Errorf("load TeraBox credentials: %w",err)}
	var credential storedTeraBoxCredential;if err:=json.Unmarshal(raw,&credential);err!=nil{return nil,errors.New("stored TeraBox credential data is invalid")}
	now:=s.now().UTC()
	if !now.Before(credential.AccessExpiresAt.Add(-5*time.Minute)){
		token,err:=s.oauth.Refresh(ctx,credential.Token.RefreshToken);if err!=nil{return nil,fmt.Errorf("refresh TeraBox access token: %w",err)}
		credential.Token=token;credential.AccessExpiresAt=now.Add(time.Duration(token.ExpiresIn)*time.Second)
		// Refresh tokens are single-use. Persist the replacement immediately,
		// before any later network call (such as token-info refresh) can fail.
		rotated,marshalErr:=json.Marshal(credential)
		if marshalErr!=nil{return nil,fmt.Errorf("encode rotated TeraBox credentials: %w",marshalErr)}
		if saveErr:=s.secrets.Set(ctx,item.CredentialRef,rotated);saveErr!=nil{return nil,fmt.Errorf("persist rotated TeraBox credentials: %w",saveErr)}
	}
	if now.Sub(credential.TokenInfoFetchedAt)>time.Hour{
		info,err:=s.oauth.TokenInformation(ctx,credential.Token.AccessToken);if err!=nil{return nil,fmt.Errorf("refresh TeraBox API domains: %w",err)}
		credential.TokenInfo=info;credential.TokenInfoFetchedAt=now
	}
	encoded,err:=json.Marshal(credential);if err!=nil{return nil,err}
	if err:=s.secrets.Set(ctx,item.CredentialRef,encoded);err!=nil{return nil,fmt.Errorf("update TeraBox credentials: %w",err)}
	return terabox.NewAPIClient(terabox.APIConfig{AccessToken:credential.Token.AccessToken,TokenInfo:credential.TokenInfo,ApplicationRoot:s.applicationRoot})
}
func(s *TeraBoxService)ConnectedProvider(ctx context.Context,accountID string)(*terabox.Adapter,error){
	api,err:=s.NewAPIClient(ctx,accountID);if err!=nil{return nil,err};return terabox.NewConnectedWithRepository(api,s.transfers),nil
}
func(s *TeraBoxService)Disconnect(ctx context.Context,accountID string)error{
	// Serialize disconnect with refresh so a concurrent refresh cannot write
	// credentials back after the account has been removed.
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	release,lockErr:=s.acquireCredentialLock(ctx,accountID);if lockErr!=nil{return fmt.Errorf("lock TeraBox credentials: %w",lockErr)};defer func(){_ = release()}()
	item,err:=s.accounts.Get(ctx,accountID);if err!=nil{return err}
	// Remove the account record first. A database failure must not destroy
	// the only reference to credentials that are still stored in the vault.
	if err:=s.accounts.Delete(ctx,accountID);err!=nil{return fmt.Errorf("delete TeraBox account record: %w",err)}
	if err:=s.secrets.Delete(ctx,item.CredentialRef);err!=nil{
		if restoreErr:=s.accounts.Save(ctx,item);restoreErr!=nil{
			return fmt.Errorf("delete TeraBox credentials: %v; restore account record: %w",err,restoreErr)
		}
		return fmt.Errorf("delete TeraBox credentials: %w",err)
	}
	return nil
}

func (s *TeraBoxService) acquireCredentialLock(ctx context.Context, accountID string) (func() error, error) {
	if s.credentialLocker == nil { return func() error { return nil }, nil }
	return s.credentialLocker.AcquireCredentialLock(ctx, accountID, 5*time.Minute)
}
