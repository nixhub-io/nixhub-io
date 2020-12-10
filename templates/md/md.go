package md

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/ningen/v2/md"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func renderInline(w util.BufWriter, src []byte, n ast.Node, enter bool) (ast.WalkStatus, error) {
	node := n.(*md.Inline)
	attr := node.Attr

	if attr.Has(md.AttrBold) {
		if enter {
			w.WriteString("<b>")
		} else {
			w.WriteString("</b>")
		}
	}
	if attr.Has(md.AttrItalics) {
		if enter {
			w.WriteString("<i>")
		} else {
			w.WriteString("</i>")
		}
	}
	if attr.Has(md.AttrUnderline) {
		if enter {
			w.WriteString("<u>")
		} else {
			w.WriteString("</u>")
		}
	}
	if attr.Has(md.AttrStrikethrough) {
		if enter {
			w.WriteString("<s>")
		} else {
			w.WriteString("</s>")
		}
	}
	if attr.Has(md.AttrSpoiler) {
		if enter {
			w.WriteString("<span style=\"font-color: #777777\">")
		} else {
			w.WriteString("</span>")
		}
	}
	if attr.Has(md.AttrMonospace) {
		if enter {
			w.WriteString("<code>")
		} else {
			w.WriteString("</code>")
		}
	}

	return ast.WalkContinue, nil
}

func renderEmoji(w util.BufWriter, src []byte, n ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		node := n.(*md.Emoji)

		var class = "emoji"
		if node.Large {
			class += " large"
		}

		fmt.Fprintf(w, `<img class="%s" src="%s" />`, class, node.EmojiURL())
	}

	return ast.WalkContinue, nil
}

func renderMention(w util.BufWriter, src []byte, n ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		node := n.(*md.Mention)

		switch {
		case node.Channel != nil:
			fmt.Fprintf(w,
				`<span class="mention">#%d</span>`,
				template.HTMLEscapeString(node.Channel.Name),
			)

		case node.GuildRole != nil:
			var color = 0x7289DA
			if node.GuildRole.Color > 0 {
				color = node.GuildRole.Color.Int()
			}

			fmt.Fprintf(w,
				// #xxxxxx22 for alpha bits.
				`<span class="mention" style="background-color: #%x22">@%s</span>`,
				color, template.HTMLEscapeString(node.GuildRole.Name),
			)

		case node.GuildUser != nil:
			var name = node.GuildUser.Username
			if node.GuildUser.Member.Nick != "" {
				name = node.GuildUser.Member.Nick
			}

			fmt.Fprintf(w,
				`<span class="mention">@%s</span>`,
				template.HTMLEscapeString(name),
			)
		}
	}

	return ast.WalkContinue, nil
}

func renderCodeBlock(w util.BufWriter, src []byte, n ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		node := n.(*ast.FencedCodeBlock)

		var builder strings.Builder
		for i := 0; i < node.Lines().Len(); i++ {
			line := node.Lines().At(i)
			builder.Write(line.Value(src))
		}

		RenderCodeBlock(string(node.Language(src)), builder.String())
	}

	return ast.WalkContinue, nil
}

func renderText(w util.BufWriter, src []byte, n ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		node := n.(*ast.Text)

		template.HTMLEscape(w, node.Segment.Value(src))

		switch {
		case node.HardLineBreak():
			w.WriteByte('\n')
		case node.SoftLineBreak():
			w.WriteString("<br>")
		}
	}

	return ast.WalkContinue, nil
}

type Renderer struct {
	base renderer.NodeRenderer
}

var htmlRenderer = func() renderer.Renderer {
	r := renderer.NewRenderer()

	reg := r.(renderer.NodeRendererFuncRegisterer)
	reg.Register(md.KindEmoji, renderEmoji)
	reg.Register(md.KindInline, renderInline)
	reg.Register(md.KindMention, renderMention)
	reg.Register(ast.KindText, renderText)
	reg.Register(ast.KindCodeBlock, renderCodeBlock)

	return r
}()

func Parse(d *state.State, m *discord.Message) template.HTML {
	var buf bytes.Buffer

	s := []byte(m.Content)
	n := md.ParseWithMessage(s, d.Cabinet, m, true)
	htmlRenderer.Render(&buf, s, n)
	return template.HTML(buf.String())
}

func ParseString(c string) template.HTML {
	var buf bytes.Buffer

	s := []byte(c)
	n := md.Parse(s)
	htmlRenderer.Render(&buf, s, n)
	return template.HTML(buf.String())
}
