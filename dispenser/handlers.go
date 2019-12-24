package dispenser

import (
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/geoip"
	"gitlab.com/nixhub/nixhub.io/templates"
	"gitlab.com/shihoya-inc/errchi"
)

func Handler(w http.ResponseWriter, r *http.Request) (int, error) {
	return geoip.Middleware(errchi.HandlerFunc(handler)).ServeHTTP(w, r)
}

func handler(w http.ResponseWriter, r *http.Request) (int, error) {
	var timezone = r.Context().Value("tz").(*time.Location)
	if c, err := templates.RenderHomepage(w, CopyPool(), timezone); err != nil {
		log.Println("Error rendering pool:", err)
		return c, err
	}

	// We need flusher to print messages live
	fl, ok := w.(http.Flusher)
	if !ok {
		return 200, nil
	}

	fl.Flush()
	return RegisterWriter(w, r.Context())
}

func AddMessage(m *discordgo.Message) {
	if MessagePool == nil {
		// Clearly hasn't called Initialize yet
		return
	}

	msg := templates.RenderMessage(m)

	MessageMu.Lock()

	// Check last author
	setLastAuthor(msg)

	// Move 1->end to 0->(end-1), then set the last element
	copy(MessagePool[0:BufSz-1], MessagePool[1:BufSz])
	MessagePool[BufSz-1] = msg

	MessageMu.Unlock()

	Broadcast(msg)
}

func DeleteMessage(m *discordgo.Message) {
	if MessagePool == nil {
		// Clearly hasn't called Initialize yet
		return
	}

	MessageMu.Lock()
	defer MessageMu.Unlock()

	for i, msg := range MessagePool {
		if msg.ID == m.ID {
			msg.Content = "<deleted>"
		}

		if i == len(MessagePool)-1 {
			// Reset last author
			LastAuthor = nil
		}
	}

	Broadcast(&templates.MessageDelete{
		ID: m.ID,
	})
}

func EditMessage(m *discordgo.Message) {
	if MessagePool == nil {
		return
	}

	MessageMu.Lock()
	defer MessageMu.Unlock()

	for _, msg := range MessagePool {
		if msg.ID == m.ID {
			old := msg.Message
			old.Content = m.Content
			old.Embeds = m.Embeds
			old.EditedTimestamp = m.EditedTimestamp
			old.Mentions = m.Mentions
			old.Attachments = m.Attachments

			new := templates.RenderMessage(old)
			*msg = *new

			return
		}
	}
}

func runHook(fn func(*discordgo.Message), chID string, m *discordgo.Message) {
	if m.ChannelID != chID {
		return
	}

	fn(m)
}
