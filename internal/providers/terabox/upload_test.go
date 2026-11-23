package terabox

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

func TestPlanTeraBoxChunks(t *testing.T){
	chunks,err:=PlanTeraBoxChunks(10<<20,8<<20);if err!=nil{t.Fatal(err)}
	if len(chunks)!=2{t.Fatalf("want 2 chunks, got %d",len(chunks))}
	var total int64;for _,c:=range chunks{if c.Size<=4<<20{t.Fatalf("part is not larger than 4 MiB: %d",c.Size)};total+=c.Size}
	if total!=10<<20{t.Fatalf("got total %d",total)}
}
func TestPlanTeraBoxChunksAvoidsSmallTail(t *testing.T){chunks,err:=PlanTeraBoxChunks((8<<20)+1,8<<20);if err!=nil{t.Fatal(err)};if len(chunks)!=1||chunks[0].Size!=(8<<20)+1{t.Fatalf("unexpected chunks: %#v",chunks)}}
func TestPlanTeraBoxChunksAboveFourGiBStaysOneLogicalFile(t *testing.T) {
	const size = int64(4<<30) + (37 << 20) + 123
	chunks, err := PlanTeraBoxChunks(size, DefaultUploadChunkSize)
	if err != nil { t.Fatal(err) }
	if len(chunks) < 2 { t.Fatalf("expected automatic multipart upload for >4 GiB, got %d part(s)", len(chunks)) }
	var offset int64
	for i, part := range chunks {
		if part.Index != i { t.Fatalf("part %d has index %d", i, part.Index) }
		if part.Offset != offset { t.Fatalf("part %d starts at %d, want contiguous offset %d", i, part.Offset, offset) }
		if part.Size <= minMultiPartSize { t.Fatalf("part %d is below TeraBox multipart minimum: %d", i, part.Size) }
		offset += part.Size
	}
	if offset != size { t.Fatalf("parts cover %d bytes, want %d", offset, size) }
	// UploadFile sends one block list to precreate and one final create call,
	// so these internal parts are assembled as one remote file.
}

func TestPlanTeraBoxChunksRejectsEmpty(t *testing.T){if _,err:=PlanTeraBoxChunks(0,8<<20);err==nil{t.Fatal("expected empty file rejection")}}

func TestUploadFilePersistsTransferAndParts(t *testing.T){
	var mu sync.Mutex
	uploaded:=map[int]string{}
	var blockList string
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		switch r.URL.Path{
		case "/openapi/api/precreate":
			if err:=r.ParseForm();err!=nil{t.Error(err)}
			if r.Form.Get("autoinit")!="1"{t.Error("autoinit missing")}
			blockList=r.Form.Get("block_list")
			fmt.Fprint(w,`{"errno":0,"uploadid":"upload-123","return_type":1,"block_list":[0,1]}`)
		case "/rest/2.0/pcs/superfile2":
			part,_:=strconv.Atoi(r.URL.Query().Get("partseq"))
			if r.URL.Query().Get("method")!="upload"||r.URL.Query().Get("uploadid")!="upload-123"{t.Errorf("bad upload query: %v",r.URL.Query())}
			f,_,err:=r.FormFile("file");if err!=nil{t.Error(err);http.Error(w,"bad multipart",400);return}
			data,err:=io.ReadAll(f);f.Close();if err!=nil{t.Error(err)}
			sum:=md5.Sum(data);digest:=hex.EncodeToString(sum[:])
			mu.Lock();uploaded[part]=digest;mu.Unlock()
			fmt.Fprintf(w,`{"errno":0,"md5":%q,"uploadid":"upload-123","partseq":%d}`,digest,part)
		case "/openapi/api/create":
			if err:=r.ParseForm();err!=nil{t.Error(err)}
			if r.Form.Get("uploadid")!="upload-123"||r.Form.Get("size")!=strconv.Itoa(10<<20)||r.Form.Get("rtype")!="1"{t.Errorf("bad create form: %v",r.Form)}
			fmt.Fprint(w,`{"errno":0,"fs_id":987,"path":"/From: Other Applications/TBox-Drive-123/big.bin","server_filename":"big.bin","size":10485760,"isdir":0}`)
		default:http.NotFound(w,r)
		}
	}))
	defer server.Close()
	local:=path.Join(t.TempDir(),"big.bin");data:=make([]byte,10<<20);for i:=range data{data[i]=byte(i%251)}
	if err:=os.WriteFile(local,data,0600);err!=nil{t.Fatal(err)}
	api,err:=NewAPIClient(APIConfig{AccessToken:"test",TokenInfo:TokenInfo{APIDomain:server.URL,UploadDomain:server.URL},ApplicationRoot:"/From: Other Applications/TBox-Drive-123"});if err!=nil{t.Fatal(err)}
	store,err:=sqlite.Open(":memory:");if err!=nil{t.Fatal(err)};defer store.Close()
	var progress int64
	result,err:=api.UploadFile(context.Background(),local,"/From: Other Applications/TBox-Drive-123/big.bin",UploadOptions{ChunkSize:8<<20,OnProgress:func(n,total int64){progress=n},Repository:store,TransferID:"task-upload-1"})
	if err!=nil{t.Fatal(err)}
	if result.TaskID!="task-upload-1"{t.Fatalf("task ID %q",result.TaskID)}
	if result.File.ID!="987"||result.File.Name!="big.bin"||result.File.Size!=10<<20{t.Fatalf("unexpected result: %#v",result.File)}
	if len(uploaded)!=2||progress!=10<<20{t.Fatalf("uploaded=%v progress=%d",uploaded,progress)}
	if !strings.HasPrefix(blockList,"[")||!strings.HasSuffix(blockList,"]"){t.Fatalf("invalid block list %q",blockList)}
	task,err:=store.GetTransfer(context.Background(),"task-upload-1");if err!=nil{t.Fatal(err)}
	if task.Status!=transfer.StatusCompleted||task.Transferred!=10<<20{t.Fatalf("persisted task: %#v",task)}
	parts,err:=store.ListParts(context.Background(),"task-upload-1");if err!=nil{t.Fatal(err)}
	if len(parts)!=2||parts[0].Status!="completed"||parts[1].Status!="completed"{t.Fatalf("persisted parts: %#v",parts)}
}
