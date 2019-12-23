package dispenser

import (
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/templates"
)

func Handler(w http.ResponseWriter, r *http.Request) (int, error) {
	if c, err := templates.RenderHomepage(w, CopyPool()); err != nil {
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
	msg.Small = LastAuthor == m.Author.ID
	LastAuthor = m.Author.ID

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

	for _, msg := range MessagePool {
		if msg.ID == m.ID {
			msg.Content = "<deleted>"
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
