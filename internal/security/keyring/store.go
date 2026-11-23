package keyring

import (
	"context"
	"errors"
	"strings"

	"github.com/zalando/go-keyring"
)

// Store uses the operating system credential vault (Windows Credential
// Manager, macOS Keychain, or a Linux Secret Service provider).
type Store struct { service string }

func New(service string)(*Store,error){
	if strings.TrimSpace(service)==""{return nil,errors.New("credential service name is required")}
	return &Store{service:service},nil
}
func(s *Store)Get(ctx context.Context,reference string)([]byte,error){
	if err:=ctx.Err();err!=nil{return nil,err};if reference==""{return nil,errors.New("credential reference is required")}
	value,err:=keyring.Get(s.service,reference);if err!=nil{return nil,err};return []byte(value),nil
}
func(s *Store)Set(ctx context.Context,reference string,value []byte)error{
	if err:=ctx.Err();err!=nil{return err};if reference==""{return errors.New("credential reference is required")};if len(value)==0{return errors.New("credential value cannot be empty")}
	return keyring.Set(s.service,reference,string(value))
}
func(s *Store)Delete(ctx context.Context,reference string)error{
	if err:=ctx.Err();err!=nil{return err};if reference==""{return errors.New("credential reference is required")}
	return keyring.Delete(s.service,reference)
}
