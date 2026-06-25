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

// Per-asset version prefix so changing one asset doesn't invalidate the cache for every other asset.
var assetVersions = computeAssetVersions(publicRoot)

func computeAssetVersions(root string) map[string]string {
	out := map[string]string{}
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
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = hex.EncodeToString(h.Sum(nil))[:8]
		return nil
	})
	if err != nil {
		log.Printf("static: failed to hash %s (%v) - serving without version prefix", root, err)
	}
	return out
}

func Asset(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "public/")
	if devMode {
		return urlPrefix + path
	}
	v, ok := assetVersions[path]
	if !ok {
		return urlPrefix + path
	}
	return urlPrefix + versionTag + v + "/" + path
}

// Versioned URLs get 1-year immutable; unversioned get 5-min so stale bookmarks/social-crawler caches don't poison a year of edge cache.
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
