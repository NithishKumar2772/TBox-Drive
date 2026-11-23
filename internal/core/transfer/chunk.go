package transfer

import "fmt"

type Chunk struct { Index int; Offset, Size int64 }

// PlanChunks creates a bounded-size plan. Provider limits must be validated
// by the adapter before these chunks are sent to a remote service.
func PlanChunks(fileSize, chunkSize int64) ([]Chunk, error) {
	if fileSize < 0 { return nil, fmt.Errorf("file size cannot be negative") }
	if chunkSize <= 0 { return nil, fmt.Errorf("chunk size must be positive") }
	if fileSize == 0 { return []Chunk{}, nil }
	count := (fileSize + chunkSize - 1) / chunkSize
	if count > int64(^uint(0)>>1) { return nil, fmt.Errorf("too many chunks") }
	chunks := make([]Chunk, 0, int(count))
	for offset, index := int64(0), 0; offset < fileSize; index++ {
		size := chunkSize
		if remaining := fileSize-offset; remaining < size { size = remaining }
		chunks = append(chunks, Chunk{Index:index, Offset:offset, Size:size})
		offset += size
	}
	return chunks, nil
}
