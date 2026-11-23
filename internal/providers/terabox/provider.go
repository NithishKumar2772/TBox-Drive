package terabox

import (
	"context"
	"errors"
	"io"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

var ErrNotImplemented = errors.New("TeraBox integration is not implemented")
type Adapter struct{}
func New() *Adapter { return &Adapter{} }
func (*Adapter) ID() string { return "terabox" }
func (*Adapter) Name() string { return "TeraBox" }
// Capabilities remain false until each operation is verified and integration-tested.
func (*Adapter) Capabilities() provider.Capabilities { return provider.Capabilities{} }
func (*Adapter) List(context.Context,string,string)([]provider.File,string,error){ return nil,"",ErrNotImplemented }
func (*Adapter) Download(context.Context,string)(io.ReadCloser,error){ return nil,ErrNotImplemented }
func (*Adapter) Upload(context.Context,provider.UploadRequest)(provider.UploadResult,error){ return provider.UploadResult{},ErrNotImplemented }
func (*Adapter) Delete(context.Context,string)error{return ErrNotImplemented}
