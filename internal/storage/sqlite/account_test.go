package sqlite

import (
	"context"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/account"
)

func TestAccountRepositoryStoresCredentialReferenceOnly(t *testing.T){
	store,err:=Open(":memory:");if err!=nil{t.Fatal(err)};defer store.Close()
	ctx:=context.Background()
	item:=account.Account{ID:"terabox:42",ProviderID:"terabox",DisplayName:"TeraBox account 42",CredentialRef:"terabox:42"}
	if err:=store.Save(ctx,item);err!=nil{t.Fatal(err)}
	got,err:=store.Get(ctx,item.ID);if err!=nil{t.Fatal(err)}
	if got.CredentialRef!="terabox:42"||got.ProviderID!="terabox"{t.Fatalf("got %#v",got)}
	items,err:=store.List(ctx);if err!=nil{t.Fatal(err)}
	if len(items)!=1{t.Fatalf("want one account, got %d",len(items))}
	if err:=store.Delete(ctx,item.ID);err!=nil{t.Fatal(err)}
	if _,err:=store.Get(ctx,item.ID);err==nil{t.Fatal("expected deleted account")}
}
