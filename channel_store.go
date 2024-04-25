package pts

import (
	"strings"
)

// ChannelStore stores pointers to all Channels
type ChannelStore struct {
	channels     map[string]*Channel
	errorHandler ErrorHandlerFunc
}

type ChannelMatch struct {
	Channel *Channel
	Params  map[string]string
}

func (s *ChannelStore) init(errorHandler ErrorHandlerFunc) {
	s.channels = map[string]*Channel{}
	s.errorHandler = errorHandler
}

func (s *ChannelStore) Register(path string, handlers ChannelHandlers) *Channel {
	channel := Channel{
		path:        strings.Split(path, channelPathSep),
		handlers:    handlers,
		subscribers: ChannelSubscribers{},
		onError:     s.errorHandler,
	}
	channel.subscribers.init()
	s.channels[path] = &channel
	return &channel
}

// Get finds all channels with matching paths.
func (s *ChannelStore) Get(path string) []ChannelMatch {
	var matching []ChannelMatch
	for _, channel := range s.channels {
		if ok, params := channel.PathMatches(path); ok {
			matching = append(matching, ChannelMatch{
				Channel: channel,
				Params:  params,
			})
		}
	}
	return matching
}

// GetByExactPath finds a channel by its exact path name.
func (s *ChannelStore) GetByExactPath(path string) (bool, *Channel) {
	channel, found := s.channels[path]
	return found, channel
}

func (s *ChannelStore) OnMessage(client *Client, message *Message) {
	if matches := s.Get(message.Channel); len(matches) > 0 {
		for _, match := range matches {
			match.Channel.HandleMessage(client, message)
		}
		return
	}
	s.errorHandler(NewError(nil, ErrorUnknownChannel, "unknown channel on websocket message: '"+message.Channel+"'", nil))
}

func (s *ChannelStore) Subscribe(client *Client, channelPath string) bool {
	if matches := s.Get(channelPath); len(matches) > 0 {
		for _, match := range matches {
			match.Channel.Subscribe(&Context{
				Client:     client,
				FullPath:   channelPath,
				Channel:    match.Channel,
				params:     match.Params,
				properties: map[string]interface{}{},
			})
		}
		return true
	}
	return false
}

func (s *ChannelStore) Unsubscribe(clientId string, channelPath string) bool {
	if matches := s.Get(channelPath); len(matches) > 0 {
		for _, match := range matches {
			match.Channel.Unsubscribe(clientId, channelPath)
		}
		return true
	}
	return false
}

func (s *ChannelStore) UnsubscribeAll(clientId string) {
	for _, channel := range s.channels {
		channel.UnsubscribeAllPaths(clientId)
	}
}
