package transfer

import (
 "fmt"
 "io"
 "math"
 "os"
)

type Chunk struct { Index int `json:"index"`; Offset int64 `json:"offset"`; Size int64 `json:"size"` }

func PlanChunks(fileSize, chunkSize int64) ([]Chunk,error) { return PlanChunksLimited(fileSize,chunkSize,1_000_000) }
func PlanChunksLimited(fileSize,chunkSize int64,maxChunks int)([]Chunk,error) {
 if fileSize<0{return nil,fmt.Errorf("file size cannot be negative")}; if chunkSize<=0{return nil,fmt.Errorf("chunk size must be positive")}; if maxChunks<0{return nil,fmt.Errorf("max chunks cannot be negative")}; if fileSize==0{return []Chunk{},nil}
 count:=1+(fileSize-1)/chunkSize; if count>int64(maxChunks)||count>int64(math.MaxInt){return nil,fmt.Errorf("chunk count %d exceeds limit %d",count,maxChunks)}
 out:=make([]Chunk,0,int(count)); for offset,index:=int64(0),0;offset<fileSize;index++ { size:=chunkSize;if remain:=fileSize-offset;remain<size{size=remain};out=append(out,Chunk{index,offset,size});offset+=size };return out,nil
}
// OpenChunk streams one validated byte range; caller must close it.
func OpenChunk(path string,c Chunk)(io.ReadCloser,error){if c.Offset<0||c.Size<0{return nil,fmt.Errorf("invalid chunk range")};f,err:=os.Open(path);if err!=nil{return nil,err};info,err:=f.Stat();if err!=nil{f.Close();return nil,err};if c.Offset>info.Size()||c.Size>info.Size()-c.Offset{f.Close();return nil,fmt.Errorf("chunk range exceeds file size")};return &chunkReadCloser{io.NewSectionReader(f,c.Offset,c.Size),f},nil}
type chunkReadCloser struct{*io.SectionReader;file *os.File}
func(r *chunkReadCloser)Close()error{return r.file.Close()}
