package templates

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/discord"
	"gitlab.com/nixhub/nixhub.io/templates/md"
)

type Messages []*Message

func (ms Messages) Render(w io.Writer) error {
	return Frontpage.ExecuteTemplate(w, "messages", ms)
}

func (ms Messages) RenderWithTimezone(w io.Writer, location *time.Location) error {
	cp := append([]*Message{}, ms...)

	for i, msg := range cp {
		cp[i].timestamp = msg.timestamp.In(location)
		cp[i].Timestamp = fmtTime(msg.timestamp)
	}

	return Messages(cp).Render(w)
}

type Message struct {
	ID        string
	Author    template.HTML
	Timestamp template.HTML
	Content   template.HTML
	AvatarURL string

	// If the last author is the same
	Small bool

	Message *discordgo.Message

	timestamp time.Time
}

var _ RendererWithTimezone = (*Message)(nil)

func (m *Message) Render(w io.Writer) error {
	return Frontpage.ExecuteTemplate(w, "message", m)
}

func (m *Message) RenderWithTimezone(w io.Writer, location *time.Location) error {
	cp := &(*m)
	cp.timestamp = m.timestamp.In(location)
	cp.Timestamp = fmtTime(m.timestamp)

	return cp.Render(w)
}

func RenderMessages(dmsgs []*discordgo.Message) Messages {
	var msgs = make([]*Message, len(dmsgs))
	for i, dm := range dmsgs {
		msgs[i] = RenderMessage(dm)
	}

	return msgs
}

func RenderMessage(dm *discordgo.Message) *Message {
	var m = Message{
		ID:      dm.ID,
		Message: dm,
	}

	// Parse everything, really
	m.Content = md.Parse(Session, dm)

	// Parse timestamp
	t, err := dm.Timestamp.Parse()
	if err == nil {
		m.timestamp = t.UTC()
		m.Timestamp = fmtTime(m.timestamp)
	}

	// Parse author
	m.Author = parseAuthor(dm.Author, dm.GuildID)

	// Parse AvatarURL
	m.AvatarURL = dm.Author.AvatarURL("64")

	return &m
}

func fmtTime(t time.Time) template.HTML {
	if t.Before(time.Now().Add(-24 * time.Hour)) {
		// If the message was from yesterday
		return template.HTML(t.Format("02/01/2006"))
	}

	return template.HTML(t.Format(time.Kitchen))
}

func parseAuthor(u *discordgo.User, guildID string) (n template.HTML) {
	n = escapeHTML(u.Username)

	// Webhooks don't have a discriminator
	if u.Discriminator == "0000" {
		return
	}

	if guildID == "" {
		log.Println("GuildID is empty")
		return
	}

	// We should try and be conservative, so no session calls
	mem, err := discord.Member(Session, guildID, u.ID)
	if err != nil {
		log.Println("Member state failed:", err)
		return
	}

	if mem.Nick != "" {
		n = escapeHTML(mem.Nick)
	}

	var top *discordgo.Role

	for _, role := range mem.Roles {
		r, err := discord.Role(Session, guildID, role)
		if err != nil {
			log.Println("Role state failed", err)
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
		`<span style="color: #%x">%s</span>`,
		top.Color, n,
	))

	if u.Bot {
		n += ` <span class="bot">BOT</span>`
	}

	return
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
	"hex": func(color int) string {
		return strconv.FormatInt(16777215, 16)
	},
	"md": func(str string) template.HTML {
		return md.ParseString(str)
	},
	"rfc3339": func(rfc3339 string) template.HTML {
		t, err := time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return ""
		}

		return fmtTime(t)
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
