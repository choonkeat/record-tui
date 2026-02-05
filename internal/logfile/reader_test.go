package logfile

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_PlainFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log")
	content := []byte("Script started on 2026-01-01\nhello world\nScript done\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOpen_GzipFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log.gz")
	content := []byte("Script started on 2026-01-01\nhello world\nScript done\n")

	// Write gzipped file
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	gz.Write(content)
	gz.Close()
	f.Close()

	rc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOpen_GzipDetectedByMagicNotExtension(t *testing.T) {
	// A gzipped file with a plain .log extension should still be decompressed
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log")
	content := []byte("hello compressed world")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	gz.Write(content)
	gz.Close()
	f.Close()

	rc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOpen_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d bytes", len(got))
	}
}

func TestOpen_SingleByte(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log")
	if err := os.WriteFile(path, []byte{0x42}, 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0x42}) {
		t.Errorf("got %v, want [0x42]", got)
	}
}

func TestOpen_FileNotFound(t *testing.T) {
	_, err := Open("/nonexistent/path.log")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadFile_Plain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log")
	content := []byte("plain content")
	os.WriteFile(path, content, 0644)

	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestReadFile_Gzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.log.gz")
	content := []byte("compressed content")

	f, _ := os.Create(path)
	gz := gzip.NewWriter(f)
	gz.Write(content)
	gz.Close()
	f.Close()

	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestCompanionPath(t *testing.T) {
	tests := []struct {
		logPath string
		ext     string
		want    string
	}{
		{"/path/session.log", ".timing", "/path/session.timing"},
		{"/path/session.log", ".input", "/path/session.input"},
		{"/path/session-abc.log", ".timing", "/path/session-abc.timing"},
		{"/path/session.log.gz", ".timing", "/path/session.timing"},
		{"/path/session.log.gz", ".input", "/path/session.input"},
		{"/path/session-abc.log.gz", ".timing", "/path/session-abc.timing"},
		{"/path/noext", ".timing", "/path/noext.timing"},
	}
	for _, tt := range tests {
		got := CompanionPath(tt.logPath, tt.ext)
		if got != tt.want {
			t.Errorf("CompanionPath(%q, %q) = %q, want %q", tt.logPath, tt.ext, got, tt.want)
		}
	}
}
