package account

import (
	"context"
	"time"
)

type Account struct {
	ID string `json:"id"`
	ProviderID string `json:"providerID"`
	DisplayName string `json:"displayName"`
	CredentialRef string `json:"credentialRef,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Repository interface {
	Save(context.Context,Account) error
	Get(context.Context,string)(Account,error)
	List(context.Context)([]Account,error)
	Delete(context.Context,string) error
}
