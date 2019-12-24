package md

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type mdState struct {
	strings.Builder
	matches   [][]match
	last      int
	chunk     string
	prev      string
	context   []string
	onlyEmoji bool
}

type match struct {
	from, to int
	str      string
}

func (s *mdState) tag(token string) string {
	var tags [2]string

	switch token {
	case "*":
		tags[0] = "<i>"
		tags[1] = "</i>"
	case "_":
		tags[0] = "<i>"
		tags[1] = "</i>"
	case "**":
		tags[0] = "<b>"
		tags[1] = "</b>"
	case "__":
		tags[0] = "<u>"
		tags[1] = "</u>"
	case "***":
		tags[0] = "<i><b>"
		tags[1] = "</b></i>"
	case "~~":
		tags[0] = "<s>"
		tags[1] = "</s>"
	case "||":
		tags[0] = `<span style="font-color: #777777">`
		tags[1] = "</span>"
	default:
		return token
	}

	var index = -1
	for i, t := range s.context {
		if t == token {
			index = i
			break
		}
	}

	if index >= 0 { // len(context) > 0 always
		s.context = append(s.context[:index], s.context[index+1:]...)
		return tags[1]
	} else {
		s.context = append(s.context, token)
		return tags[0]
	}
}

func (s mdState) getLastIndex(currentIndex int) int {
	if currentIndex >= len(s.matches) {
		return 0
	}

	return s.matches[currentIndex][0].to
}

func (s *mdState) switchTreeMessage(i int, d *discordgo.Session, m *discordgo.Message) {
	switch {
	case s.matches[i][7].str != "":
		// user mentions
		s.chunk = UserNicknameHTML(d, m, s.matches[i][7].str)
	case s.matches[i][8].str != "":
		// role mentions
		s.chunk = RoleNameHTML(d, m, s.matches[i][8].str)
	case s.matches[i][9].str != "":
		// channel mentions
		s.chunk = ChannelNameHTML(d, m, s.matches[i][9].str)
	default:
		s.switchTree(i)
	}
}

func (s *mdState) switchTree(i int) {
	switch {
	case s.matches[i][2].str != "":
		// codeblock
		s.chunk = RenderCodeBlock(s.matches[i][1].str, s.matches[i][2].str)

	case s.matches[i][3].str != "":
		// blockquotes, greentext
		s.chunk += "<blockquote>"
		// Slice away the first character, as it's just a ">"
		s.chunk += template.HTMLEscapeString(strings.TrimPrefix(
			s.matches[i][3].str[1:], "\n"))
		s.chunk += "</blockquote>"

	case s.matches[i][4].str != "":
		// inline code
		s.chunk += "<code>"
		s.chunk += template.HTMLEscapeString(s.matches[i][4].str)
		s.chunk += "</code>"

	case s.matches[i][5].str != "":
		// inline stuff
		s.chunk = s.tag(s.matches[i][5].str)

	case s.matches[i][6].str != "":
		// hyperlink
		var url = template.HTMLEscapeString(s.matches[i][6].str)
		s.chunk = fmt.Sprintf(`<a href="%s">%s</a>`, url, url)

	case s.matches[i][11].str != "":
		// emojis
		var url = EmojiURL(
			s.matches[i][11].str,
			s.matches[i][10].str == "a",
		)
		var class = "emoji"
		if s.prev == "" {
			class += " large"
		}
		s.chunk = `<img class="` + class + `" src="` + url + `" />`

	default:
		s.chunk = template.HTMLEscapeString(s.matches[i][0].str)
	}
}

func submatch(r *regexp.Regexp, s string) [][]match {
	found := r.FindAllStringSubmatchIndex(s, -1)
	indices := make([][]match, len(found))

	var m = match{-1, -1, ""}

	for i := range found {
		indices[i] = make([]match, len(found[i])/2)

		for a, b := range found[i] {
			if a%2 == 0 { // first pair
				m.from = b
			} else {
				m.to = b

				if m.from >= 0 && m.to >= 0 {
					m.str = s[m.from:m.to]
				} else {
					m.str = ""
				}

				indices[i][a/2] = m
			}
		}
	}

	return indices
}
