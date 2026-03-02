package updater

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultUser is the GitHub user hosting the database repo.
	DefaultUser = "Rvn0xsy"
	// DefaultRepo is the GitHub repository name for the database.
	DefaultRepo = "red-tldr-db"
)

type githubTag struct {
	Name       string `json:"name"`
	ZipBallURL string `json:"zipball_url"`
}

// FetchLatestFromGithub downloads the latest database release and extracts it.
func FetchLatestFromGithub(destDir string) error {
	tag, err := fetchLatestTag(DefaultUser + "/" + DefaultRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch latest tag: %w", err)
	}

	fmt.Printf("[Updating Database version: %s]\n", strings.TrimSpace(tag.Name))

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := downloadAndUnzip(ctx, tag.ZipBallURL, destDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	fmt.Println("[Update Database Success.]")
	return nil
}

func fetchLatestTag(repo string) (*githubTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tags []githubTag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags found for %s", repo)
	}
	return &tags[0], nil
}

func downloadAndUnzip(ctx context.Context, downloadURL string, destDir string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	reader := bytes.NewReader(buf)
	z, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}

	return SafeUnzip(z, destDir)
}

// SafeUnzip extracts zip contents with protection against zip-slip attacks.
func SafeUnzip(r *zip.Reader, dest string) error {
	dest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	for _, f := range r.File {
		if err := extractZipEntry(f, dest); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(f *zip.File, dest string) error {
	parts := strings.SplitN(f.Name, "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return nil
	}
	relPath := parts[1]

	target := filepath.Join(dest, relPath)

	// Zip-slip protection MUST run before any skip logic to prevent
	// path-traversal attacks disguised as hidden files (e.g. "../../etc/passwd")
	if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("zip slip detected: %s attempts to escape to %s", f.Name, target)
	}

	for _, segment := range strings.Split(relPath, "/") {
		if strings.HasPrefix(segment, ".") {
			return nil
		}
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Limit extraction size to prevent zip bombs (100 MB per file)
	_, err = io.Copy(outFile, io.LimitReader(rc, 100*1024*1024))
	return err
}
