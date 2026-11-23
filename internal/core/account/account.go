package account

import (
	"context"
	"time"
)

type Account struct {
	ID string
	ProviderID string
	DisplayName string
	CredentialRef string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	Save(context.Context,Account) error
	Get(context.Context,string)(Account,error)
	List(context.Context)([]Account,error)
	Delete(context.Context,string) error
}
