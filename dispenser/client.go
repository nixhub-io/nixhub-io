package dispenser

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nixhub-io/nixhub-io/templates"
)

const PoolBuf = 5

func (s *State) Broadcast(r templates.Renderer) {
	s.ClientMu.RLock()
	defer s.ClientMu.RUnlock()

	for i, ch := range s.ClientPool {
		select {
		case ch <- r:
			log.Println("Sent to client", i)
		case <-time.After(time.Second / 2):
			log.Println("Client", i, "timed out")
		}
	}
}

func (s *State) RegisterWriter(w io.Writer, ctx context.Context) (int, error) {
	var inc = make(chan templates.Renderer, PoolBuf)
	defer s.registerClient(inc)()

	var flush = w.(http.Flusher).Flush

	var tz = time.Local
	if loc, ok := ctx.Value("tz").(*time.Location); ok {
		tz = loc
	}

	for {
		select {
		case m, ok := <-inc:
			if !ok {
				// Exit, channel closed
				return 200, nil
			}

			if c, err := templates.Render(w, m, tz); err != nil {
				log.Println("Error rendering message:", err)
				return c, err
			}

			flush()

		case <-ctx.Done():
			return 200, nil
		}
	}
}

func (s *State) registerClient(inc chan<- templates.Renderer) (cancel func()) {
	s.ClientMu.Lock()
	defer s.ClientMu.Unlock()

	var c = s.Counter
	s.Counter++

	s.ClientPool[c] = inc
	log.Println("Registered", c)

	return s.makeFree(c)
}

func (s *State) makeFree(c uint64) func() {
	return func() {
		s.ClientMu.Lock()
		defer s.ClientMu.Unlock()

		log.Println("Freeing", c)
		delete(s.ClientPool, c)
	}
}
