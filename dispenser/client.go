package dispenser

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.com/nixhub/nixhub.io/templates"
)

var ClientPool = map[uint64]chan<- templates.Renderer{}
var ClientMu sync.Mutex

const PoolBuf = 5

func Broadcast(r templates.Renderer) {
	for i, ch := range ClientPool {
		select {
		case ch <- r:
			log.Println("Sent to client", i)
		case <-time.After(time.Second / 2):
			log.Println("Client", i, "timed out")
		}
	}
}

func RegisterWriter(w io.Writer, ctx context.Context) (int, error) {
	var inc = make(chan templates.Renderer, PoolBuf)
	defer registerClient(inc)()

	var flush = w.(http.Flusher).Flush

	var tz = time.Local
	if loc, ok := ctx.Value("tz").(*time.Location); ok {
		tz = loc
	}

	for {
		select {
		case m := <-inc:
			if c, err := templates.RenderHomepage(w, m, tz); err != nil {
				log.Println("Error rendering message:", err)
				return c, err
			}

			flush()

		case <-ctx.Done():
			return 200, nil
		}
	}

	// Is this really unreachable?
	// return 500, errors.New("unexpected channel death")
}

var counter uint64

func registerClient(inc chan<- templates.Renderer) (cancel func()) {
	ClientMu.Lock()
	defer ClientMu.Unlock()

	var c = counter
	counter++

	ClientPool[c] = inc
	log.Println("Registered", c)

	return func() {
		ClientMu.Lock()
		defer ClientMu.Unlock()

		log.Println("Freeing", c)
		delete(ClientPool, c)
	}
}
