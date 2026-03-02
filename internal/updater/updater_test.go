package updater

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, entries [][2]string) *zip.Reader {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for _, e := range entries {
		f, err := w.Create(e[0])
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", e[0], err)
		}
		if _, err := f.Write([]byte(e[1])); err != nil {
			t.Fatalf("failed to write zip entry %s: %v", e[0], err)
		}
	}
	w.Close()

	reader := bytes.NewReader(buf.Bytes())
	r, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		t.Fatalf("failed to open test zip: %v", err)
	}
	return r
}

func TestSafeUnzip_NormalExtraction(t *testing.T) {
	dest := t.TempDir()

	r := createTestZip(t, [][2]string{
		{"repo-abc123/files/test.yaml", "name: test"},
		{"repo-abc123/db/db.json", "{}"},
	})

	if err := SafeUnzip(r, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "files", "test.yaml")); err != nil {
		t.Error("expected files/test.yaml to be extracted")
	}
	if _, err := os.Stat(filepath.Join(dest, "db", "db.json")); err != nil {
		t.Error("expected db/db.json to be extracted")
	}
}

func TestSafeUnzip_SkipsHiddenFiles(t *testing.T) {
	dest := t.TempDir()

	r := createTestZip(t, [][2]string{
		{"repo-abc123/.github/workflows/ci.yml", "ci config"},
		{"repo-abc123/.gitignore", "*.exe"},
		{"repo-abc123/files/test.yaml", "name: test"},
	})

	if err := SafeUnzip(r, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, ".github")); err == nil {
		t.Error("expected .github to be skipped")
	}
	if _, err := os.Stat(filepath.Join(dest, ".gitignore")); err == nil {
		t.Error("expected .gitignore to be skipped")
	}
	if _, err := os.Stat(filepath.Join(dest, "files", "test.yaml")); err != nil {
		t.Error("expected files/test.yaml to be extracted")
	}
}

// Zip-slip is a critical security vulnerability where malicious zip archives
// use path traversal (../) to write files outside the intended destination.
func TestSafeUnzip_ZipSlipPrevention(t *testing.T) {
	dest := t.TempDir()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("repo-abc123/../../etc/passwd")
	f.Write([]byte("malicious content"))
	w.Close()

	reader := bytes.NewReader(buf.Bytes())
	r, _ := zip.NewReader(reader, reader.Size())

	err := SafeUnzip(r, dest)
	if err == nil {
		t.Error("expected zip slip to be detected and blocked")
	}

	if _, err := os.Stat(filepath.Join(dest, "..", "etc", "passwd")); err == nil {
		t.Error("zip slip attack succeeded - malicious file was created")
	}
}

func TestSafeUnzip_EmptyZip(t *testing.T) {
	dest := t.TempDir()
	r := createTestZip(t, [][2]string{})

	if err := SafeUnzip(r, dest); err != nil {
		t.Fatalf("unexpected error for empty zip: %v", err)
	}
}
