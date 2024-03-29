package pts

import (
	"strings"
)

// ChannelStore stores pointers to all Channels
type ChannelStore struct {
	channels     map[string]*Channel
	errorHandler ErrorHandlerFunc
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

// Get finds a channel with a matching path.
func (s *ChannelStore) Get(path string) (bool, *Channel, map[string]string) {
	if found, channel := s.GetByExactPath(path); found {
		return true, channel, map[string]string{}
	}

	for _, channel := range s.channels {
		if ok, params := channel.PathMatches(path); ok {
			return true, channel, params
		}
	}

	return false, nil, nil
}

// GetByExactPath finds a channel by its exact path name.
func (s *ChannelStore) GetByExactPath(path string) (bool, *Channel) {
	channel, found := s.channels[path]
	return found, channel
}

func (s *ChannelStore) OnMessage(client *Client, message *Message) {
	if ok, channel, _ := s.Get(message.Channel); ok {
		channel.HandleMessage(client, message)
		return
	}
	s.errorHandler(NewError(nil, ErrorUnknownChannel, "unknown channel on websocket message: '"+message.Channel+"'", nil))
}

func (s *ChannelStore) Subscribe(client *Client, channelPath string) bool {
	if found, channel, params := s.Get(channelPath); found {
		channel.Subscribe(&Context{
			Client:     client,
			FullPath:   channelPath,
			Channel:    channel,
			params:     params,
			properties: map[string]interface{}{},
		})
		return true
	}
	return false
}

func (s *ChannelStore) Unsubscribe(clientId string, channelPath string) bool {
	if found, channel, _ := s.Get(channelPath); found {
		return channel.Unsubscribe(clientId, channelPath)
	}
	return false
}

func (s *ChannelStore) UnsubscribeAll(clientId string) {
	for _, channel := range s.channels {
		channel.UnsubscribeAllPaths(clientId)
	}
}
