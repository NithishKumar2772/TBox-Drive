//go:build desktop

package main

import (
	"bytes"
	"testing"
)

func TestDownloadProgressWriterReportsFinalSmallTransfer(t *testing.T) {
	var output bytes.Buffer
	var reported int64
	var total int64
	writer := &downloadProgressWriter{
		writer: &output,
		total: 5,
		onProgress: func(done, size int64) {
			reported = done
			total = size
		},
	}
	if _, err := writer.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("de")); err != nil {
		t.Fatal(err)
	}
	writer.report()
	if output.String() != "abcde" {
		t.Fatalf("download output = %q, want abcde", output.String())
	}
	if reported != 5 || total != 5 {
		t.Fatalf("final progress = %d/%d, want 5/5", reported, total)
	}
}

func TestDownloadProgressWriterDoesNotReportWhenNoBytesWereWritten(t *testing.T) {
	var reports int
	writer := &downloadProgressWriter{
		writer: &bytes.Buffer{},
		total: 0,
		onProgress: func(int64, int64) { reports++ },
	}
	writer.report()
	if reports != 0 {
		t.Fatalf("reports = %d, want 0", reports)
	}
}
