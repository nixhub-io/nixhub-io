package templates

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"github.com/nixhub-io/nixhub-io/templates/md"
)

type Messages []Message

func (ms Messages) Render(w io.Writer) error {
	return Frontpage.ExecuteTemplate(w, "messages", ms)
}

func (ms Messages) RenderWithTimezone(w io.Writer, location *time.Location) error {
	cp := append(Messages(nil), ms...)

	for i, msg := range cp {
		cp[i].timestamp = msg.timestamp.In(location)
		cp[i].Timestamp = fmtTime(msg.timestamp)
	}

	return Messages(cp).Render(w)
}

type Message struct {
	ID        discord.MessageID
	Author    template.HTML
	Timestamp template.HTML
	Content   template.HTML
	AvatarURL string

	// If the last author is the same
	Small bool

	Message discord.Message

	timestamp time.Time
}

var _ RendererWithTimezone = (*Message)(nil)

func (m Message) Render(w io.Writer) error {
	return Frontpage.ExecuteTemplate(w, "message", m)
}

func (m Message) RenderWithTimezone(w io.Writer, location *time.Location) error {
	m.timestamp = m.timestamp.In(location)
	m.Timestamp = fmtTime(m.timestamp)

	return m.Render(w)
}

func RenderMessages(state *state.State, dmsgs []discord.Message) Messages {
	var msgs = make([]Message, len(dmsgs))
	for i, dm := range dmsgs {
		msgs[i] = RenderMessage(state, dm)
	}

	return msgs
}

func RenderMessage(state *state.State, dm discord.Message) Message {
	var m = Message{
		ID:      dm.ID,
		Message: dm,
	}

	// Parse everything, really
	m.Content = md.Parse(state, &dm)

	// Parse timestamp
	m.timestamp = dm.Timestamp.Time()
	m.Timestamp = fmtTime(m.timestamp)

	// Parse author
	m.Author = parseAuthor(state, dm.Author, dm.GuildID)

	// Parse AvatarURL
	m.AvatarURL = dm.Author.AvatarURL() + "?size=64"

	return m
}

func fmtTime(t time.Time) template.HTML {
	if t.Before(time.Now().Add(-24 * time.Hour)) {
		// If the message was from yesterday
		return template.HTML(t.Format("02/01/2006"))
	}

	return template.HTML(t.Format(time.Kitchen))
}

func parseAuthor(
	state *state.State, u discord.User, guildID discord.GuildID) template.HTML {

	var name = u.Username
	var color = discord.Color(0xFFFFFF)

	if !guildID.IsValid() {
		log.Println("GuildID is empty")
		return ""
	}

	// We should try and be conservative, so no session calls
	mem, err := state.Member(guildID, u.ID)
	if err == nil {
		if mem.Nick != "" {
			name = mem.Nick
		}

		c, err := state.MemberColor(guildID, u.ID)
		if err == nil {
			color = c
		}
	}

	// Wrap the username around color codes
	var html = fmt.Sprintf(
		`<span style="color: #%x">%s</span>`,
		color, html.EscapeString(name),
	)

	if u.Bot {
		html += ` <span class="bot">BOT</span>`
	}

	return template.HTML(html)
}

var (
	ImageFormats = []string{"jpg", "jpeg", "png", "webm", "gif"}
	VideoFormats = []string{"mkv", "webm", "mp4", "ogv"}
	AudioFormats = []string{"mp3", "flac", "ogg", "opus", "aac"}
)

// Helper functions for messages
var htmlFns = template.FuncMap{
	"URLIsImage": func(url string) bool {
		return checkURLfiletype(url, ImageFormats)
	},
	"URLIsVideo": func(url string) bool {
		return checkURLfiletype(url, VideoFormats)
	},
	"URLIsAudio": func(url string) bool {
		return checkURLfiletype(url, AudioFormats)
	},
	"hex": func(color discord.Color) string {
		return strconv.FormatUint(uint64(color.Uint32()), 16)
	},
	"md": func(str string) template.HTML {
		return md.ParseString(str)
	},
	"rfc3339": func(t discord.Timestamp) template.HTML {
		if !t.IsValid() {
			return ""
		}
		return fmtTime(t.Time())
	},
}

func checkURLfiletype(url string, filetypes []string) bool {
	ext := path.Ext(url)
	if ext == "" {
		return false
	}

	ext = strings.ToLower(ext)

	for _, ft := range filetypes {
		if ext == "."+ft {
			return true
		}
	}

	return false
}

func escapeHTML(s string) template.HTML {
	return template.HTML(template.HTMLEscapeString(s))
}
