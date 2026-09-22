package services

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func newDownloadFixture(t *testing.T) (*DownloadService, *fs.LocalFileRepository) {
	t.Helper()
	dir := t.TempDir()
	repo := fs.NewLocalFileRepository(dir)
	scoper := policy.NewPathScoper()
	return NewDownloadService(NewDownloadFileService(repo, scoper), NewDownloadZipService(repo, scoper)), repo
}

func writeRepoFile(t *testing.T, repo *fs.LocalFileRepository, path, content string) {
	t.Helper()
	if _, err := repo.WriteFile(path, io.NopCloser(strings.NewReader(content))); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// A real file whose name ends in ".zip" must download as a file, not be
// mistaken for a directory-zip request.
func TestDownloadServiceServesZipNamedFileAsFile(t *testing.T) {
	svc, repo := newDownloadFixture(t)
	writeRepoFile(t, repo, "archive.zip", "PK-not-really")

	download, err := svc.Execute(nil, "archive.zip")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	defer download.Stream.Close()

	if download.ContentType != "application/octet-stream" {
		t.Errorf("content type = %q, want application/octet-stream", download.ContentType)
	}
	if download.Filename != "archive.zip" {
		t.Errorf("filename = %q, want archive.zip", download.Filename)
	}
	body, _ := io.ReadAll(download.Stream)
	if string(body) != "PK-not-really" {
		t.Errorf("body = %q, want %q", body, "PK-not-really")
	}
}

// Directories are zipped on the fly.
func TestDownloadServiceZipsDirectory(t *testing.T) {
	svc, repo := newDownloadFixture(t)
	if err := repo.CreateDirectory("docs"); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeRepoFile(t, repo, "docs/a.txt", "hello")

	download, err := svc.Execute(nil, "docs")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	defer download.Stream.Close()

	if download.ContentType != "application/zip" {
		t.Errorf("content type = %q, want application/zip", download.ContentType)
	}
	if download.Filename != "docs.zip" {
		t.Errorf("filename = %q, want docs.zip", download.Filename)
	}
	data, err := io.ReadAll(download.Stream)
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("parse zip: %v", err)
	}

	var found bool
	for _, f := range reader.File {
		if f.Name == "a.txt" {
			found = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open entry: %v", err)
			}
			body, _ := io.ReadAll(rc)
			rc.Close()
			if string(body) != "hello" {
				t.Errorf("entry body = %q, want hello", body)
			}
		}
	}
	if !found {
		t.Errorf("zip entries = %v, want to contain a.txt", reader.File)
	}
}
