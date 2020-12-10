package dispenser

import (
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/nixhub-io/nixhub-io/templates"
)

const TypingTimeout = 10 * time.Second

func (s *State) subscribeTyping() error {
	// err := s.Session.Gateway.GuildSubscribe(gateway.GuildSubscribeData{
	// 	GuildID: s.GuildID,
	// 	Typing:  true,
	// })

	// if err != nil {
	// 	return errors.Wrap(err, "GuildSubscribe failed")
	// }

	s.Typing = make(chan *gateway.TypingStartEvent)
	s.Typers = &templates.Typing{
		Typers: make([]templates.Typer, 0, 5),
	}
	s.StopType = make(chan discord.User)

	go s.typingLoop()

	s.addHandler(func(t *gateway.TypingStartEvent) {
		if t.ChannelID != s.ChannelID {
			return
		}

		s.Typing <- t
	})

	return nil
}

func (s *State) StopTyping(user discord.User) {
	s.StopType <- user
}

func (s *State) typingLoop() {
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t, ok := <-s.Typing:
			if !ok {
				return
			}

			m, err := s.Session.Member(s.GuildID, t.UserID)
			if err != nil {
				// Shouldn't happen, we just ignore it
				continue
			}

			var typer = templates.Typer{
				UserID: m.User.ID,
				Name:   m.User.Username,
				When:   time.Unix(int64(t.Timestamp), 0),
			}

			if m.Nick != "" {
				typer.Name = m.Nick
			}

			s.Typers.AddTyper(typer)

		case user := <-s.StopType:
			if !s.Typers.Filter(func(t templates.Typer) bool { return user.ID == t.UserID }) {
				// Not changed, continue
				continue
			}

		case now := <-ticker.C:
			if !s.Typers.Filter(func(t templates.Typer) bool {
				// After means it's not timed out yet, so we keep it (true)
				return t.When.Add(TypingTimeout).After(now)
			}) {
				// Not changed, continue
				continue
			}
		}

		// All of the above mutate something, so we broadcast it. They don't if
		// they continue, though.
		go s.Broadcast(s.Typers)
	}
}
