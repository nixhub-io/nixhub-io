package md

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/discord"
)

var regexes = []string{
	// codeblock
	`(?m)(?:\x60\x60\x60 *(\w*)\n([\s\S]*?)\x60\x60\x60$)`,
	// blockquote
	`((?:(?:^|\n)>\s+.*)+)`,
	// I think this is inline code?
	`(?:(?:^|\n)(?:[>*+-]|\d+\.)\s+.*)+|(?:\x60([^\x60].*?)\x60)`,
	// Inline markup stuff
	`(__|\*\*\*|\*\*|[_*]|~~|\|\|)`,
	// Hyperlinks
	`(https?:\/\S+(?:\.|:)\S+)`,
	// User mentions
	`(?:<@!?(\d+)>)`,
	// Role mentions
	`(?:<@&(\d+)>)`,
	// Channel mentions
	`(?:<#(\d+)>)`,
	// Emojis
	`(?:<(a?):.*:(\d+)>)`,
}

var r1 = regexp.MustCompile(`(?m)` + strings.Join(regexes, "|"))

func Parse(d *discordgo.Session, m *discordgo.Message) template.HTML {
	return parse("", d, m)
}

func ParseString(c string) template.HTML {
	return parse(c, nil, nil)
}

func parse(c string, d *discordgo.Session, m *discordgo.Message) template.HTML {
	var s mdState

	var md string
	if c != "" {
		md = c
	} else {
		if d == nil || m == nil {
			return ""
		}

		md = m.Content
	}

	s.matches = submatch(r1, md)

	for i := 0; i < len(s.matches); i++ {
		s.prev = md[s.last:s.matches[i][0].from]
		s.last = s.getLastIndex(i)
		s.chunk = "" // reset chunk

		switch {
		case strings.Count(s.prev, "\\")%2 != 0:
			s.chunk = template.HTMLEscapeString(s.matches[i][0].str)
		case c != "":
			s.switchTree(i)
		default:
			s.switchTreeMessage(i, d, m)
		}

		s.WriteString(template.HTMLEscapeString(s.prev))
		s.WriteString(s.chunk)
	}

	s.WriteString(md[s.last:])

	for len(s.context) > 0 {
		s.WriteString(s.tag(s.context[len(s.context)-1]))
	}

	return template.HTML(strings.TrimSpace(s.String()))
}

func UserNicknameHTML(s *discordgo.Session, m *discordgo.Message,
	userID string) string {

	var mentioned *discordgo.User

	for _, mention := range m.Mentions {
		if mention.ID == userID {
			mentioned = mention
			break
		}
	}

	if mentioned == nil {
		return "@" + userID
	}

	var name = mentioned.Username

	mem, err := discord.Member(s, m.GuildID, mentioned.ID)
	if err == nil && mem.Nick != "" {
		name = mem.Nick
	}

	return `<span class="mention">@` +
		template.HTMLEscapeString(name) + `</span>`
}

func RoleNameHTML(s *discordgo.Session, m *discordgo.Message,
	roleID string) string {

	var mroleID string

	for _, mention := range m.MentionRoles {
		if mention == roleID {
			mroleID = mention
			break
		}
	}

	if mroleID == "" {
		return "<@&" + roleID + ">"
	}

	r, err := discord.Role(s, m.GuildID, mroleID)
	if err != nil {
		return "<@&" + roleID + ">"
	}

	// Default color
	if r.Color == 0 {
		r.Color = 0x7289da
	}

	return fmt.Sprintf(
		`<span class="mention" style="background-color: #%x">@%s</span>`,
		r.Color, template.HTMLEscapeString(r.Name),
	)
}

func ChannelNameHTML(s *discordgo.Session, m *discordgo.Message,
	channelID string) string {

	var mentioned *discordgo.Channel

	for _, ch := range m.MentionChannels {
		if ch.ID == channelID {
			mentioned = ch
			break
		}
	}

	if mentioned == nil {
		return "<#" + channelID + ">"
	}

	c, err := discord.Channel(s, channelID)
	if err != nil {
		return "<#" + channelID + ">"
	}

	return `<span class="mention">#` +
		template.HTMLEscapeString(c.Name) + `</span>`
}

const EmojiBaseURL = "https://cdn.discordapp.com/emojis/"

func EmojiURL(emojiID string, animated bool) string {
	if animated {
		return EmojiBaseURL + emojiID + ".gif"
	}

	return EmojiBaseURL + emojiID + ".png"
}
