package templates

import (
	"fmt"
	"html/template"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

// Helper functions for messages

var htmlFns = template.FuncMap{
	"messageContent": func(m *discordgo.Message) template.HTML {
		return escapeHTML(m.ContentWithMentionsReplaced())
	},
	"messageTimestamp": func(m *discordgo.Message) template.HTML {
		t, err := m.Timestamp.Parse()
		if err != nil {
			return ""
		}

		return template.HTML(humanize.Time(t))
	},
	"messageAuthorName": func(m *discordgo.Message) (n template.HTML) {
		n = escapeHTML(m.Author.Username)

		if m.GuildID == "" {
			return
		}

		// We should try and be conservative, so no session calls
		mem, err := Session.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		if mem.Nick != "" {
			n = escapeHTML(mem.Nick)
		}

		var top *discordgo.Role

		for _, role := range mem.Roles {
			r, err := Session.State.Role(m.GuildID, role)
			if err != nil {
				continue
			}

			if r.Color == 0 {
				continue
			}

			if top == nil || r.Position > top.Position {
				top = r
			}
		}

		if top == nil {
			return
		}

		// Wrap the username around color codes
		n = template.HTML(fmt.Sprintf(
			`<span style="color: %x">%s</span>`,
			top.Color, n,
		))

		return
	},
}

func escapeHTML(s string) template.HTML {
	return template.HTML(template.HTMLEscapeString(s))
}
