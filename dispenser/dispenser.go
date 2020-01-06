package dispenser

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gitlab.com/nixhub/nixhub.io/templates"
)

const BufSz = 25

type State struct {
	MessagePool templates.Messages
	LastAuthor  *discordgo.User
	MessageMu   sync.Mutex

	Typers   *templates.Typing
	Typing   chan *discordgo.TypingStart
	StopType chan<- *discordgo.User // userID

	Session   *discordgo.Session
	ChannelID string
	GuildID   string

	ClientPool map[uint64]chan<- templates.Renderer
	ClientMu   sync.RWMutex
	Counter    uint64

	closers []func()
}

func (s *State) CopyPool() templates.Messages {
	s.MessageMu.Lock()
	defer s.MessageMu.Unlock()

	return append([]*templates.Message{}, s.MessagePool...)
}

// setLastAuthor asserts and returns new.Small
func (s *State) setLastAuthor(new *templates.Message) bool {
	if s.LastAuthor == nil {
		if new != nil {
			s.LastAuthor = new.Message.Author
		}

		return false
	}

	new.Small = (s.LastAuthor.ID == new.Message.Author.ID &&
		s.LastAuthor.Username == new.Message.Author.Username)
	s.LastAuthor = new.Message.Author

	return new.Small
}

func (s *State) Close() error {
	s.ClientMu.Lock()
	defer s.ClientMu.Unlock()

	// Delete handlers
	for _, c := range s.closers {
		c()
	}

	// Gracefully end all connections
	for id, ch := range s.ClientPool {
		close(ch)
		delete(s.ClientPool, id)
	}

	// Stop the type loop
	close(s.Typing)

	return nil
}

func (s *State) addHandler(fn interface{}) {
	s.closers = append(s.closers, s.Session.AddHandler(fn))
}

func Initialize(s *discordgo.Session, channelID string) (*State, error) {
	var state = State{
		Session:    s,
		ChannelID:  channelID,
		ClientPool: make(map[uint64]chan<- templates.Renderer),
	}

	templates.Session = s

	msgs, err := s.ChannelMessages(channelID, BufSz, "", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get messages from "+channelID)
	}

	// The first message is the earliest.

	// Discord is retarded, so we have to fetch GuildID ourselves
	c, err := s.Channel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get channelID "+channelID)
	}

	state.GuildID = c.GuildID

	state.MessageMu.Lock()
	defer state.MessageMu.Unlock()

	state.MessagePool = make([]*templates.Message, BufSz)

	// This loop iterates messages from the latest to the earliest.
	for i, j := 0, BufSz-1; i < len(msgs); i++ {
		// Grab and insert missing GuildID
		msg := msgs[i]
		msg.GuildID = c.GuildID

		state.MessagePool[j] = templates.RenderMessage(msg)

		if j--; j < 0 {
			break
		}
	}

	// We need another loop to calculate small, as it requires iterating from
	// earliest to latest. We iterate from the start of MessagePool, as that's
	// earliest.
	for _, msg := range state.MessagePool {
		state.setLastAuthor(msg)
	}

	// Add hooks
	state.addHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		runHook(state.AddMessage, channelID, m.Message)
	})
	state.addHandler(func(_ *discordgo.Session, m *discordgo.MessageDelete) {
		runHook(state.DeleteMessage, channelID, m.Message)
	})
	state.addHandler(func(_ *discordgo.Session, m *discordgo.MessageUpdate) {
		runHook(state.EditMessage, channelID, m.Message)
	})

	// Ask to fill up state
	if err := s.RequestGuildMembers(c.GuildID, "", 0); err != nil {
		return nil, errors.Wrap(err, "Failed to request all members")
	}

	// Subscribe to typing events
	if err := state.subscribeTyping(); err != nil {
		return nil, errors.Wrap(err, "Failed to subscribe to guild")
	}

	return &state, nil
}
