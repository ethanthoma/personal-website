package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"slices"
	"sort"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/highlight"
	"personal-website/services/webserver/pages"
)

var (
	duplicatedTitleH1Re        = regexp.MustCompile(`(?s)^\s*<h1[^>]*>.*?</h1>\s*`)
	duplicatedTLDRBlockquoteRe = regexp.MustCompile(`(?s)^\s*<blockquote[^>]*>.*?</blockquote>\s*`)
	headingAnchorRe            = regexp.MustCompile(`(?s)<(h[23]) id="([^"]+)">(.*?)</h[23]>`)
)

// Safe to share-cache per slug (no per-visitor data); s-maxage tracks the 5m post cache.
const postCacheControl = "public, s-maxage=300, max-age=60, stale-while-revalidate=600, stale-if-error=86400"

func postHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !internal.IsValidSlug(slug) {
		notFoundHandler(w, r)
		return
	}

	post, err := slugToHTML(slug)
	if err != nil {
		log.Printf("post %s not found (%v)", slug, err)
		notFoundHandler(w, r)
		return
	}
	older, newer := postNeighbors(slug)
	w.Header().Set("Cache-Control", postCacheControl)
	pages.Post{}.View(post, older, newer).Render(r.Context(), w)
}

func postsByDateDesc() ([]internal.Post, error) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		return posts, err
	}
	sort.SliceStable(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	return posts, nil
}

func slugToHTML(slug string) (internal.Post, error) {
	post, err := loadPost(slug)
	if err != nil {
		return post, err
	}
	post.HTML, err = renderPostBody(post.Content, post.TLDR != "")
	return post, err
}

func loadPost(slug string) (internal.Post, error) {
	post, err := cache.Cache.GetPost(slug)
	if err == nil {
		return post, nil
	}
	if cache.Cache.IsRecentlyAbsent(slug) {
		return internal.Post{}, fmt.Errorf("post %q recently absent", slug)
	}

	post, err = internal.GetPostFromGitHub(slug)
	if err != nil {
		cache.Cache.MarkAbsent(slug)
	}
	return post, err
}

func renderPostBody(markdown string, hasTLDR bool) (template.HTML, error) {
	var buf bytes.Buffer
	if err := highlight.Renderer.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	html := duplicatedTitleH1Re.ReplaceAllString(buf.String(), "")
	if hasTLDR {
		html = duplicatedTLDRBlockquoteRe.ReplaceAllString(html, "")
	}
	html = headingAnchorRe.ReplaceAllString(html,
		`<$1 id="$2">$3<a href="#$2" class="heading-anchor" aria-label="Link to this section">#</a></$1>`)
	return template.HTML(html), nil
}

func postNeighbors(slug string) (older, newer *pages.PostNeighbor) {
	posts, err := postsByDateDesc()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
		return nil, nil
	}
	i := slices.IndexFunc(posts, func(p internal.Post) bool { return p.Slug == slug })
	if i < 0 {
		return nil, nil
	}
	total := len(posts)
	if i+1 < total {
		older = postNeighborAt(posts, i+1, total)
	}
	if i > 0 {
		newer = postNeighborAt(posts, i-1, total)
	}
	return older, newer
}

func postNeighborAt(posts []internal.Post, i, total int) *pages.PostNeighbor {
	p := posts[i]
	return &pages.PostNeighbor{
		Slug:   p.Slug,
		Title:  p.Title,
		Number: fmt.Sprintf("%03d", total-i-1),
	}
}
