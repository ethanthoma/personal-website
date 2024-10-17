package post

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"personal-website/cmd/webserver/cache"
	"personal-website/cmd/webserver/layouts/base"
	"personal-website/internal"
)

type Props struct {
	Ascii       [][]string
	PageCurrent string
	Pages       []string
	Post        internal.Post
}

func (p *Props) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := "post"

	// Page props
	p.PageCurrent = "blog"

	slug := r.PathValue("slug")
	post, err := getPost(slug)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	p.Post = post

	// Page Template
	t, err := template.New("").Funcs(template.FuncMap{
		"formatDate": func(date time.Time) string {
			return date.Format("20060102")
		},
	}).ParseFiles("cmd/webserver/pages/" + name + "/" + name + ".tmpl")
	if err != nil {
		log.Printf(name+": failed to parse tmpl file (%v)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Layout
	if err = (base.Props{
		Ascii:       p.Ascii,
		PageCurrent: p.PageCurrent,
		Pages:       p.Pages,
	}.Layout(t)); err != nil {
		log.Printf(name+": failed to render layout (%v)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Render
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")

	if strings.HasSuffix(r.URL.Path, "/content") {
		if err = t.ExecuteTemplate(w, "content", p); err != nil {
			log.Printf(name+": failed to render content (%v)", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = t.ExecuteTemplate(w, "oob", p); err != nil {
			log.Printf(name+": failed to render oob (%v)", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err = t.ExecuteTemplate(w, "page", p); err != nil {
		log.Printf(name+": failed to render page (%v)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getPost(slug string) (internal.Post, error) {
	post, err := cache.Cache.GetPost(slug)
	if err != nil {
		log.Printf("Error getting post %s: %v", slug, err)
		return post, err
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("rose-pine"),
			),
		),
	)

	var buf bytes.Buffer
	err = mdRenderer.Convert([]byte(post.Content), &buf)
	if err != nil {
		log.Printf("Error parsing post %s to markdown: %v", slug, err)
		return post, err
	}

	post.HTML = template.HTML(buf.String())

	return post, nil
}
