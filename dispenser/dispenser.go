package dispenser

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gitlab.com/nixhub/nixhub.io/templates"
)

const BufSz = 25

var MessagePool templates.Messages
var LastAuthor *discordgo.User // calculate small
var MessageMu sync.Mutex

var Session *discordgo.Session

func CopyPool() templates.Messages {
	MessageMu.Lock()
	defer MessageMu.Unlock()

	return append([]*templates.Message{}, MessagePool...)
}

// setLastAuthor asserts and returns new.Small
func setLastAuthor(new *templates.Message) bool {
	if LastAuthor == nil {
		LastAuthor = new.Message.Author
		return false
	}

	new.Small = (LastAuthor.ID == new.Message.Author.ID &&
		LastAuthor.Username == new.Message.Author.Username)
	LastAuthor = new.Message.Author

	return new.Small
}

func Initialize(s *discordgo.Session, channelID string) error {
	Session = s
	templates.Session = s

	msgs, err := s.ChannelMessages(channelID, BufSz, "", "", "")
	if err != nil {
		return errors.Wrap(err, "Failed to get messages from "+channelID)
	}

	// The first message is the earliest.

	// Discord is retarded, so we have to fetch GuildID ourselves
	c, err := s.Channel(channelID)
	if err != nil {
		return errors.Wrap(err, "Failed to get channelID "+channelID)
	}

	MessageMu.Lock()
	defer MessageMu.Unlock()

	MessagePool = make([]*templates.Message, BufSz)

	// This loop iterates messages from the latest to the earliest.
	for i, j := 0, BufSz-1; i < len(msgs); i++ {
		// Grab and insert missing GuildID
		msg := msgs[i]
		msg.GuildID = c.GuildID

		MessagePool[j] = templates.RenderMessage(msg)

		if j--; j < 0 {
			break
		}
	}

	// We need another loop to calculate small, as it requires iterating from
	// earliest to latest. We iterate from the start of MessagePool, as that's
	// earliest.
	for _, msg := range MessagePool {
		setLastAuthor(msg)
	}

	// Add hooks
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		runHook(AddMessage, channelID, m.Message)
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		runHook(DeleteMessage, channelID, m.Message)
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		runHook(EditMessage, channelID, m.Message)
	})

	// Ask to fill up state
	if err := s.RequestGuildMembers(c.GuildID, "", 0); err != nil {
		return errors.Wrap(err, "Failed to request all members")
	}

	return nil
}
