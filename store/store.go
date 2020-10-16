// Package store implements custom state stores.
package store

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
)

// Store extends DefaultStore to not include emojis, messages, presences and
// voice states inside its store.
type Store struct {
	*state.DefaultStore
}

func New() *Store {
	return &Store{
		DefaultStore: state.NewDefaultStore(
			&state.DefaultStoreOptions{
				MaxMessages: 0,
			},
		),
	}
}

func (s *Store) EmojiSet(discord.GuildID, []discord.Emoji) error {
	return nil
}

func (s *Store) MessageSet(discord.Message) error {
	return nil
}

func (s *Store) MessageRemove(discord.ChannelID, discord.MessageID) error {
	return nil
}

func (s *Store) PresenceSet(discord.GuildID, discord.Presence) error {
	return nil
}

func (s *Store) PresenceRemove(discord.GuildID, discord.UserID) error {
	return nil
}

func (s *Store) VoiceStateSet(discord.GuildID, discord.VoiceState) error {
	return nil
}

func (s *Store) VoiceStateRemove(discord.GuildID, discord.UserID) error {
	return nil
}
