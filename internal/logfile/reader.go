package logfile

import (
	"compress/gzip"
	"io"
	"os"
	"strings"
)

// gzipMagic is the two-byte magic number at the start of gzip files.
var gzipMagic = []byte{0x1f, 0x8b}

// Open opens a log file for reading, transparently decompressing gzip files.
// It detects gzip files by checking for the gzip magic bytes at the start of
// the file. The returned ReadCloser must be closed by the caller.
func Open(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Read first 2 bytes to check for gzip magic
	header := make([]byte, 2)
	n, err := f.Read(header)
	if err != nil || n < 2 {
		// File is too short to be gzip; rewind and return as-is
		f.Seek(0, io.SeekStart)
		return f, nil
	}

	// Rewind for either path
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, err
	}

	if header[0] == gzipMagic[0] && header[1] == gzipMagic[1] {
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &gzipReadCloser{gz: gz, f: f}, nil
	}

	return f, nil
}

// ReadFile reads the entire contents of a log file, transparently
// decompressing gzip files. It is a drop-in replacement for os.ReadFile
// that also handles .log.gz files.
func ReadFile(path string) ([]byte, error) {
	rc, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// CompanionPath derives a companion file path (e.g. .timing, .input) from a
// log file path. It handles both plain .log and compressed .log.gz extensions.
//
// Examples:
//
//	CompanionPath("/path/session.log", ".timing")    → "/path/session.timing"
//	CompanionPath("/path/session.log.gz", ".timing") → "/path/session.timing"
func CompanionPath(logPath string, ext string) string {
	if strings.HasSuffix(logPath, ".log.gz") {
		return logPath[:len(logPath)-len(".log.gz")] + ext
	}
	if strings.HasSuffix(logPath, ".log") {
		return logPath[:len(logPath)-len(".log")] + ext
	}
	return logPath + ext
}

// gzipReadCloser wraps a gzip.Reader and the underlying file so that
// closing the wrapper closes both.
type gzipReadCloser struct {
	gz *gzip.Reader
	f  *os.File
}

func (g *gzipReadCloser) Read(p []byte) (int, error) {
	return g.gz.Read(p)
}

func (g *gzipReadCloser) Close() error {
	gzErr := g.gz.Close()
	fErr := g.f.Close()
	if gzErr != nil {
		return gzErr
	}
	return fErr
}
