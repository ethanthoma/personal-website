package blog

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"personal-website/cmd/webserver/layouts/base"

	spacer "personal-website/cmd/webserver/components/spacer"

	"personal-website/cmd/webserver/cache"
	"personal-website/internal"
)

type Props struct {
	Ascii       [][]string
	PageCurrent string
	Pages       []string
	Posts       []internal.Post
}

func (p *Props) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := "blog"

	// Page props
	p.PageCurrent = name

	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf(name+": failed to get posts (%v)", err)
	}
	p.Posts = posts

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
	if err = (base.
		Props{
		Ascii:       p.Ascii,
		PageCurrent: p.PageCurrent,
		Pages:       p.Pages,
	}.Layout(t)); err != nil {
		log.Printf(name+": failed to render layout (%v)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Components
	if err = (spacer.Props{}).Component(t); err != nil {
		log.Printf(name+": failed to render spacer (%v)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Render
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
