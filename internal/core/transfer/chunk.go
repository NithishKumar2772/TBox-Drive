package transfer

import "fmt"

const MaxPlannedChunks = 1_000_000

type Chunk struct { Index int; Offset, Size int64 }

// PlanChunks creates a bounded-size plan. Provider limits must be validated
// by the adapter before these chunks are sent to a remote service.
func PlanChunks(fileSize,chunkSize int64)([]Chunk,error){
 if fileSize<0{return nil,fmt.Errorf("file size cannot be negative")}
 if chunkSize<=0{return nil,fmt.Errorf("chunk size must be positive")}
 if fileSize==0{return []Chunk{},nil}
 // This form avoids overflow in (fileSize + chunkSize - 1).
 count:=1+(fileSize-1)/chunkSize
 if count>int64(^uint(0)>>1)||count>MaxPlannedChunks{return nil,fmt.Errorf("chunk plan exceeds the maximum of %d chunks",MaxPlannedChunks)}
 chunks:=make([]Chunk,0,int(count))
 for offset,index:=int64(0),0;offset<fileSize;index++{
  size:=chunkSize;if remaining:=fileSize-offset;remaining<size{size=remaining}
  chunks=append(chunks,Chunk{Index:index,Offset:offset,Size:size});offset+=size
 }
 return chunks,nil
}
