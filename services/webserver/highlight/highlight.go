package highlight

import (
	"bytes"
	"log"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

// chroma palette for fenced code blocks. Defined once at package init so
// every post render reuses the same Style + Markdown renderer instead of
// rebuilding them per request.
var (
	style    = buildStyle()
	CSS      = renderCSS(style)
	Renderer = buildRenderer(style)
)

func buildStyle() *chroma.Style {
	b := chroma.NewStyleBuilder("default")
	b.AddEntry(chroma.Background, chroma.MustParseStyleEntry("bg:#ansiblack"))
	b.AddEntry(chroma.Keyword, chroma.MustParseStyleEntry("#FFA500"))
	b.AddEntry(chroma.Name, chroma.MustParseStyleEntry("#ansilightgray"))
	b.AddEntry(chroma.NameVariable, chroma.MustParseStyleEntry("#A500FF"))
	b.AddEntry(chroma.NameBuiltin, chroma.MustParseStyleEntry("#FF00A5"))
	b.AddEntry(chroma.NameFunction, chroma.MustParseStyleEntry("#00A5FF"))
	b.AddEntry(chroma.Literal, chroma.MustParseStyleEntry("#ansigreen"))
	b.AddEntry(chroma.LiteralNumber, chroma.MustParseStyleEntry("#ansigreen"))
	b.AddEntry(chroma.LiteralString, chroma.MustParseStyleEntry("#ansigreen"))
	b.AddEntry(chroma.LineNumbers, chroma.MustParseStyleEntry("#ansidarkgray"))
	b.AddEntry(chroma.Punctuation, chroma.MustParseStyleEntry("#a5a5a5"))
	b.AddEntry(chroma.Generic, chroma.MustParseStyleEntry("#ansiwhite"))
	b.AddEntry(chroma.Operator, chroma.MustParseStyleEntry("#ansiwhite"))
	b.AddEntry(chroma.Text, chroma.MustParseStyleEntry("#ansiwhite"))
	s, err := b.Build()
	if err != nil {
		log.Fatalf("highlight: build style: %v", err)
	}
	return s
}

// renderCSS emits the chroma stylesheet for `style` as a single string,
// using the same class names the renderer below stamps on each <span>.
// The result is injected into post.templ via a <style> block — small
// (~1KB), only loaded with posts, and replaces ~280 inline style="..."
// attributes per post page.
func renderCSS(s *chroma.Style) string {
	f := chromahtml.New(chromahtml.WithClasses(true))
	var buf bytes.Buffer
	if err := f.WriteCSS(&buf, s); err != nil {
		log.Fatalf("highlight: write css: %v", err)
	}
	return buf.String()
}

func buildRenderer(s *chroma.Style) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithCustomStyle(s),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
					chromahtml.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)
}
