package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func Download(url string, dst string) error {
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// temporarily skip insecure certificates
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadIfOutdated(url string, dst string, minMtime time.Time) error {
	mtime, err := GetMtime(dst)
	if err != nil && mtime.After(minMtime) {
		return nil
	}

	return Download(url, dst)
}
