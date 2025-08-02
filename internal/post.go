package internal

import (
	"bytes"
	"github.com/grokify/html-strip-tags-go"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"log"
	"strings"
	"time"
)

type Post struct {
	Slug    string
	Title   string
	Content string
	Date    time.Time
	HTML    template.HTML
	TLDR    string
}


func extract_tldr(content string) string {
	lines := strings.Split(content, "\n")
	inTldrSection := false
	var tldrBuilder strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "> tldr;") {
			inTldrSection = true
			trimmedLine = strings.TrimSpace(trimmedLine[7:])
			if len(trimmedLine) > 0 {
				trimmedLine = strings.ToUpper(trimmedLine[0:1]) + trimmedLine[1:]
			}
			tldrBuilder.WriteString(trimmedLine + " ")
			continue
		}

		if inTldrSection {
			if !strings.HasPrefix(trimmedLine, "> ") {
				break
			}

			tldrBuilder.WriteString(strings.TrimSpace(trimmedLine[2:]) + " ")
		}
	}

	content = strings.TrimSpace(tldrBuilder.String())

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			extension.NewTypographer(
				extension.WithTypographicSubstitutions(map[extension.TypographicPunctuation]string{
					extension.LeftSingleQuote:  "'",
					extension.RightSingleQuote: "'",
					extension.LeftDoubleQuote:  "",
					extension.RightDoubleQuote: "",
					extension.EnDash:           "–",
					extension.EmDash:           "—",
					extension.Ellipsis:         "...",
					extension.LeftAngleQuote:   "<",
					extension.RightAngleQuote:  ">",
					extension.Apostrophe:       "'",
				}),
			),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	err := mdRenderer.Convert([]byte(content), &buf)
	if err != nil {
		log.Printf("error parsing post to markdown (%v)", err)
		return ""
	}

	content = strings.TrimSpace(strip.StripTags(string(template.HTML(buf.String()))))

	return content
}
