package main

import (
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/pages"
)

var navList = []string{"home", "resources"}

var navHandlers = map[string]http.HandlerFunc{
	"home":      homeHandler,
	"resources": resourcesHandler,
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
	}
	log.Printf("Home handler: rendering with %d posts", len(posts))
	pages.Home{Pages: navList}.View(posts).Render(r.Context(), w)
}

func resourcesHandler(w http.ResponseWriter, r *http.Request) {
	pages.InfoRes{Pages: navList}.View().Render(r.Context(), w)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	pages.NotFound{Pages: navList, Path: r.URL.Path}.View().Render(r.Context(), w)
}

func postsListHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := postsByDateDesc()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
	}
	start, end := pageBounds(r, len(posts), pages.PostsPageSize)
	pages.PostsList(posts[start:end], len(posts), start).Render(r.Context(), w)
}

func projectsListHandler(w http.ResponseWriter, r *http.Request) {
	projects := slices.Clone(internal.Projects)
	sort.SliceStable(projects, func(i, j int) bool {
		return projects[i].Date.After(projects[j].Date)
	})
	start, end := pageBounds(r, len(projects), pages.ProjectsPageSize)
	pages.ProjectsList(projects[start:end], len(projects), start).Render(r.Context(), w)
}

func navFragmentHandler(w http.ResponseWriter, r *http.Request) {
	h, ok := navHandlers[r.PathValue("name")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	asFragment(h)(w, r)
}

func pageBounds(r *http.Request, total, size int) (start, end int) {
	start, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end = start + size
	if end > total {
		end = total
	}
	return start, end
}
