package md

import (
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// HighlightStyle determines the syntax highlighting colorstyle:
// https://xyproto.github.io/splash/docs/all.html
const HighlightStyle = "solarized-dark"

var (
	style  = styles.Get(HighlightStyle)
	fmtter = html.New(
		// html.WithLineNumbers(true),
		html.WithClasses(false),
	)
)

// RenderCodeBlock renders the node to a syntax
// highlighted code
func RenderCodeBlock(lang, content string) string {
	if style == nil {
		style = styles.Fallback
	}

	var lexer = lexers.Fallback
	if lang := string(lang); lang != "" {
		if l := lexers.Get(lang); l != nil {
			lexer = l
		} else {
			content = lang + "\n" + content
		}
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return "<pre>" + template.HTMLEscapeString(content) + "</pre>"
	}

	var code strings.Builder

	if err := fmtter.Format(&code, style, iterator); err != nil {
		return content
	}

	return code.String()
}
