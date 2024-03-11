package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var DownloadCmd = &cobra.Command{
	Use:   "download [version]",
	Short: "Download specify version of Go",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		for _, arg := range args {
			v, err := parseVersionString(arg)
			if err != nil {
				fatal(ctx, err)
			}
			if err := Download(ctx, v); err != nil {
				fatal(ctx, err)
			}
		}
	},
}

var maxWorkers = runtime.NumCPU() * 4

type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha_256"`
		Size     int    `json:"size"`
		Kind     string `json:"kind"`
	} `json:"files"`
	priority int
}

const (
	goVersionURL = "https://go.dev/dl/?mode=json&include=all"
	downloadURL  = "https://storage.googleapis.com/golang/"
)

func findTarget(ctx context.Context, v *version) (*GoVersion, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, goVersionURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("status is %d. response %s", resp.StatusCode, body)
	}

	var versions []*GoVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	var filtered []*GoVersion
	for _, version := range versions {
		splits := strings.Split(strings.TrimLeft(version.Version, "go"), ".")
		if compareVersionString(splits, v) {
			version.priority = calcPriority(splits)
			filtered = append(filtered, version)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("specify version is not found")
	}

	slices.SortFunc(filtered, func(l, r *GoVersion) int {
		return r.priority - l.priority
	})

	return filtered[0], nil
}

func (g *GoVersion) getDownloadFile() string {
	for _, file := range g.Files {
		if file.Arch == runtime.GOARCH && file.OS == runtime.GOOS {
			return file.Filename
		}
	}
	return ""
}

func Download(ctx context.Context, v *version) error {
	base, err := checkInit()
	if err != nil {
		return err
	}
	goversion, err := findTarget(ctx, v)
	if err != nil {
		return err
	}

	file := goversion.getDownloadFile()
	url, err := url.JoinPath(downloadURL, file)
	if err != nil {
		return err
	}
	infof(ctx, "download %s", url)
	tmpFile, err := download(ctx, url)
	if err != nil {
		return err
	}

	infof(ctx, "extract %s", tmpFile.Name())
	dir, err := extract(tmpFile)
	if err != nil {
		return err
	}

	fromDir := filepath.Join(dir, "go")
	infof(ctx, "copy from %s", fromDir)
	targetPath := filepath.Join(base, "versions", goversion.Version)
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("remove %s: %w", targetPath, err)
	}
	if err := os.Rename(fromDir, targetPath); err != nil {
		return err
	}

	return nil
}

func extract(file *os.File) (string, error) {
	gr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("new gzip reader: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gr); err != nil {
		return "", fmt.Errorf("copy to buffer: %w", err)
	}

	dir := file.Name()[:strings.LastIndex(file.Name(), ".tar.gz")]
	if err := os.Mkdir(dir, 0o755); err != nil {
		return "", err
	}

	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if hdr.Typeflag == tar.TypeDir {
			dir := filepath.Join(dir, hdr.Name)
			if _, err = os.Stat(dir); os.IsNotExist(err) {
				if err := os.Mkdir(dir, hdr.FileInfo().Mode()); err != nil {
					return "", err
				}
			}
			continue
		}
		if hdr.Typeflag == tar.TypeSymlink {
			if err := os.Symlink(hdr.Linkname, filepath.Join(dir, hdr.Name)); err != nil {
				return "", err
			}
			continue
		}

		file, err := os.OpenFile(filepath.Join(dir, hdr.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(file, tr); err != nil {
			file.Close()
			return "", fmt.Errorf("copy to %s: %w", file.Name(), err)
		}
		file.Close()
	}

	return dir, err
}

func download(ctx context.Context, url string) (*os.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("response status is %d and body is '%s'", resp.StatusCode, body)
	}

	size := resp.ContentLength
	chunk := int(size / int64(maxWorkers))
	if size%int64(maxWorkers) != 0 {
		chunk++
	}
	var (
		wg   sync.WaitGroup
		buf  = make([]bytes.Buffer, maxWorkers)
		errs error
	)
	for i := 0; i < maxWorkers; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}

			if i+1 == maxWorkers {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", i*chunk))
			} else {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i*chunk, (i+1)*chunk-1))
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusPartialContent {
				errs = errors.Join(errs, fmt.Errorf("response status is %d", resp.StatusCode))
				return
			}
			if _, err := io.Copy(&buf[i], resp.Body); err != nil {
				errs = errors.Join(errs, err)
				return
			}
		}()
	}

	wg.Wait()

	if errs != nil {
		return nil, errs
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "*.tar.gz")
	if err != nil {
		return nil, err
	}
	for _, b := range buf {
		if _, err := tmpFile.Write(b.Bytes()); err != nil {
			return nil, err
		}
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}
	tmpFile, err = os.Open(tmpFile.Name())
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}
