package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
)

const siteOrigin = "https://www.ethanthoma.com"

func rssHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := postsByDateDesc()
	if err != nil {
		log.Printf("rss: failed to fetch posts (%v)", err)
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	writeRSSFeed(w, posts)
}

func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf("sitemap: failed to fetch posts (%v)", err)
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writeSitemap(w, posts)
}

func writeRSSFeed(w io.Writer, posts []internal.Post) {
	buildDate := time.Now().UTC()
	if len(posts) > 0 {
		buildDate = posts[0].Date
	}
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(w, `<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	fmt.Fprintln(w, `  <channel>`)
	fmt.Fprintln(w, `    <title>Ethan Thoma</title>`)
	fmt.Fprintf(w, "    <link>%s</link>\n", siteOrigin)
	fmt.Fprintln(w, `    <description>Notes on machine learning, NLP, and programming languages.</description>`)
	fmt.Fprintln(w, `    <language>en-us</language>`)
	fmt.Fprintf(w, "    <atom:link href=\"%s/rss.xml\" rel=\"self\" type=\"application/rss+xml\" />\n", siteOrigin)
	fmt.Fprintf(w, "    <lastBuildDate>%s</lastBuildDate>\n", buildDate.Format(time.RFC1123Z))
	for _, p := range posts {
		writeRSSItem(w, p)
	}
	fmt.Fprintln(w, `  </channel>`)
	fmt.Fprintln(w, `</rss>`)
}

func writeRSSItem(w io.Writer, p internal.Post) {
	link := siteOrigin + "/post/" + p.Slug
	fmt.Fprintln(w, `    <item>`)
	fmt.Fprintf(w, "      <title>%s</title>\n", html.EscapeString(p.Title))
	fmt.Fprintf(w, "      <link>%s</link>\n", link)
	fmt.Fprintf(w, "      <guid isPermaLink=\"true\">%s</guid>\n", link)
	fmt.Fprintf(w, "      <pubDate>%s</pubDate>\n", p.Date.Format(time.RFC1123Z))
	if p.TLDR != "" {
		fmt.Fprintf(w, "      <description>%s</description>\n", html.EscapeString(p.TLDR))
	}
	fmt.Fprintln(w, `    </item>`)
}

func writeSitemap(w io.Writer, posts []internal.Post) {
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(w, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, loc := range []string{"/home", "/rss.xml"} {
		fmt.Fprintf(w, "  <url><loc>%s%s</loc></url>\n", siteOrigin, loc)
	}
	for _, p := range posts {
		lastmod := p.Date
		if !p.LastModified.IsZero() {
			lastmod = p.LastModified
		}
		fmt.Fprintf(w, "  <url><loc>%s/post/%s</loc><lastmod>%s</lastmod></url>\n",
			siteOrigin, p.Slug, lastmod.Format("2006-01-02"))
	}
	fmt.Fprintln(w, `</urlset>`)
}
