package static

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	publicRoot = "public"
	urlPrefix  = "/public/"
	versionTag = "v_"
)

var devMode = os.Getenv("DEV") == "1"

// Version is the deploy-time content hash of public/, computed once at package
// init
var Version = computeVersion(publicRoot)

func computeVersion(root string) string {
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		h.Write([]byte(rel))
		h.Write([]byte{0})
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("static: failed to hash %s (%v) - falling back to dev marker", root, err)
		return "dev"
	}
	return hex.EncodeToString(h.Sum(nil))[:8]
}

// Asset returns the long-cached, version-prefixed URL for a file under public/.
// Accepts paths with or without a leading slash and with or without the
// "public/" prefix
func Asset(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "public/")
	if devMode {
		return urlPrefix + path
	}
	return urlPrefix + versionTag + Version + "/" + path
}

// URLs starting with /public/v_<hash>/ are deploy-versioned and get a 1-year
// immutable Cache-Control. Any other /public/ URL gets a 5-minute cache so
// stale bookmarks or external refs (favicon links, OG images cached by social
// crawlers) still resolve without poisoning a year of edge cache.
func Handler() http.Handler {
	fs := http.FileServer(http.Dir(publicRoot))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, urlPrefix)
		immutable := false
		if strings.HasPrefix(rest, versionTag) {
			if i := strings.IndexByte(rest, '/'); i > 0 {
				rest = rest[i+1:]
				immutable = true
			}
		}
		switch {
		case devMode:
			w.Header().Set("Cache-Control", "no-store")
		case immutable:
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/" + rest
		fs.ServeHTTP(w, r2)
	})
}
