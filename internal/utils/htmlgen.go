package utils

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func IsHTML(s string) bool {
	return strings.Count(s, "<") >= 2 && strings.Count(s, ">") >= 2
}

func IsMarkdown(s string) bool {
	return strings.Count(s, "*")+strings.Count(s, "[") >= 2
}

func CreateHTML(s string) template.HTML {
	if IsHTML(s) {
		return template.HTML(s)
	}
	if IsMarkdown(s) {
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(s))

		// create HTML renderer with extensions
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		return template.HTML(string(markdown.Render(doc, renderer)))
	}

	s = strings.ReplaceAll(s, "\n", "<br>")
	reHttp := regexp.MustCompile(`(https?://\S+)`)
	reHttp.ReplaceAllString(s, `<a href="$1" target="_blank">$1</a>`)
	return template.HTML(s)
}
