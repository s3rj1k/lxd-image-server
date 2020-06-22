package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	cmdPort int
	cmdPath string
)

/*
  # Location example for reverse SSL proxy:

  location / {
    proxy_buffering off;
    proxy_http_version 1.1;
    proxy_set_header X-Forwarded-Host $server_name;
    proxy_pass http://127.0.0.1:8000;
  }
*/

func getFullURL(r *http.Request) string {
	// LXD Image server only works from https, so we assume here that reverse proxy is configured correctly
	return "https://" + r.Header.Get("X-Forwarded-Host") + r.URL.Path
}

func isExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func getSHA256HashSum(path string) ([]byte, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, fd); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func writeFileToWriter(path string, dst io.Writer) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer fd.Close()

	if _, err := io.Copy(dst, fd); err != nil {
		return err
	}

	return nil
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(cmdPath + r.URL.Path)

	if !strings.HasSuffix(path, ".tar.gz") {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if !isExist(path) {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	hash, err := getSHA256HashSum(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/gzip")
	w.Header().Add("LXD-Image-Hash", fmt.Sprintf("%x", hash))
	w.Header().Add("LXD-Image-URL", getFullURL(r))

	if err := writeFileToWriter(path, w); err != nil {
		return
	}
}

func main() {
	flag.IntVar(&cmdPort, "p", 8000, "port to serve template images from")
	flag.StringVar(&cmdPath, "d", "/vz/template/cache/lxd", "the directory path for template images")

	flag.Parse()

	cmdPath = filepath.Clean(cmdPath)

	http.HandleFunc("/", serveTemplate)

	if err := http.ListenAndServe(
		fmt.Sprintf("127.0.0.1:%d", cmdPort),
		nil,
	); err != nil {
		log.Fatal(err)
	}
}
