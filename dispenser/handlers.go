package dispenser

import (
	"log"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/nixhub-io/nixhub-io/geoip"
	"github.com/nixhub-io/nixhub-io/templates"
	"github.com/pkg/errors"
	"gitlab.com/shihoya-inc/errchi"
)

func (s *State) Handler(w http.ResponseWriter, r *http.Request) (int, error) {
	return geoip.Middleware(errchi.HandlerFunc(s.handler)).ServeHTTP(w, r)
}

func (s *State) handler(w http.ResponseWriter, r *http.Request) (int, error) {
	var timezone = r.Context().Value("tz").(*time.Location)
	if c, err := templates.Render(w, s.CopyPool(), timezone); err != nil {
		log.Println("Error rendering pool:", err)
		return c, errors.Wrap(err, "Failed to render pool")
	}

	// We need flusher to print messages live
	fl, ok := w.(http.Flusher)
	if !ok {
		return 200, nil
	}

	// Flush the typing indicators too, why not
	if err := s.Typers.Render(w); err != nil {
		return 500, errors.Wrap(err, "Failed to render typers")
	}

	fl.Flush()
	return s.RegisterWriter(w, r.Context())
}

func (s *State) AddMessage(m discord.Message) {
	if s.MessagePool == nil {
		// Clearly hasn't called Initialize yet
		return
	}

	msg := templates.RenderMessage(s.Session, m)

	s.MessageMu.Lock()

	// Check last author
	s.setLastAuthor(&msg)

	// Move 1->end to 0->(end-1), then set the last element
	copy(s.MessagePool[0:BufSz-1], s.MessagePool[1:BufSz])
	s.MessagePool[BufSz-1] = msg

	s.MessageMu.Unlock()

	s.Broadcast(msg)
	s.StopTyping(m.Author)
}

func (s *State) DeleteMessage(id discord.MessageID) {
	if s.MessagePool == nil {
		// Clearly hasn't called Initialize yet
		return
	}

	s.MessageMu.Lock()
	defer s.MessageMu.Unlock()

	for i, msg := range s.MessagePool {
		if msg.ID == id {
			msg.Content = "<deleted>"
		}

		if i == len(s.MessagePool)-1 {
			// Reset last author
			s.LastAuthor = nil
		}
	}

	s.Broadcast(templates.MessageDelete{ID: id})
}

func (s *State) EditMessage(m discord.Message) {
	if s.MessagePool == nil {
		return
	}

	s.MessageMu.Lock()
	defer s.MessageMu.Unlock()

	for i, msg := range s.MessagePool {
		if msg.ID == m.ID {
			old := msg.Message
			old.Content = m.Content
			old.Embeds = m.Embeds
			old.EditedTimestamp = m.EditedTimestamp
			old.Mentions = m.Mentions
			old.Attachments = m.Attachments

			s.MessagePool[i] = templates.RenderMessage(s.Session, old)

			return
		}
	}
}
