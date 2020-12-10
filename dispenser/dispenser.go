package dispenser

import (
	"sync"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/nixhub-io/nixhub-io/templates"
	"github.com/pkg/errors"
)

const BufSz = 25

type State struct {
	MessagePool templates.Messages
	MessageMu   sync.Mutex
	LastAuthor  *discord.User

	Typers   *templates.Typing
	Typing   chan *gateway.TypingStartEvent
	StopType chan discord.User // userID

	Session   *state.State
	ChannelID discord.ChannelID
	GuildID   discord.GuildID

	ClientPool map[uint64]chan<- templates.Renderer
	ClientMu   sync.RWMutex
	Counter    uint64

	closers []func()
}

func (s *State) CopyPool() templates.Messages {
	s.MessageMu.Lock()
	defer s.MessageMu.Unlock()

	return append([]templates.Message(nil), s.MessagePool...)
}

// setLastAuthor asserts and returns new.Small
func (s *State) setLastAuthor(newMessage *templates.Message) {
	if s.LastAuthor != nil {
		newMessage.Small = true &&
			s.LastAuthor.ID == newMessage.Message.Author.ID &&
			s.LastAuthor.Username == newMessage.Message.Author.Username
	}

	var author = newMessage.Message.Author
	s.LastAuthor = &author
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

func Initialize(s *state.State, channelID discord.ChannelID) (*State, error) {
	var state = State{
		Session:    s,
		ChannelID:  channelID,
		ClientPool: make(map[uint64]chan<- templates.Renderer),
	}

	msgs, err := s.Client.Messages(channelID, BufSz)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get messages from "+channelID.String())
	}

	// The first message is the earliest.

	// Discord is retarded, so we have to fetch GuildID ourselves
	c, err := s.Channel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get channelID "+channelID.String())
	}

	state.GuildID = c.GuildID

	state.MessageMu.Lock()
	defer state.MessageMu.Unlock()

	state.MessagePool = make([]templates.Message, BufSz)

	// This loop iterates messages from the latest to the earliest.
	for i, j := 0, BufSz-1; i < len(msgs); i++ {
		// Grab and insert missing GuildID
		msg := msgs[i]
		msg.GuildID = c.GuildID

		state.MessagePool[j] = templates.RenderMessage(state.Session, msg)

		if j--; j < 0 {
			break
		}
	}

	// We need another loop to calculate small, as it requires iterating from
	// earliest to latest. We iterate from the start of MessagePool, as that's
	// earliest.
	for i := range state.MessagePool {
		state.setLastAuthor(&state.MessagePool[i])
	}

	// Add hooks
	state.addHandler(func(m *gateway.MessageCreateEvent) {
		if m.ChannelID == channelID {
			state.AddMessage(m.Message)
		}
	})
	state.addHandler(func(m *gateway.MessageDeleteEvent) {
		if m.ChannelID == channelID {
			state.DeleteMessage(m.ID)
		}
	})
	state.addHandler(func(m *gateway.MessageUpdateEvent) {
		if m.ChannelID == channelID {
			state.EditMessage(m.Message)
		}
	})

	request := gateway.RequestGuildMembersData{
		GuildID: []discord.GuildID{c.GuildID},
	}

	// Ask to fill up state
	if err := s.Gateway.RequestGuildMembers(request); err != nil {
		return nil, errors.Wrap(err, "Failed to request all members")
	}

	// Subscribe to typing events
	if err := state.subscribeTyping(); err != nil {
		return nil, errors.Wrap(err, "Failed to subscribe to guild")
	}

	return &state, nil
}
