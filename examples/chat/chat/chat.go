package chat

import (
	"encoding/json"
	"fmt"
	"github.com/mono424/go-pts"
)

type Chat struct {
	users      map[string]bool
	prefix     string
	tubeSystem *pts.TubeSystem
}

func New(prefix string, tubeSystem *pts.TubeSystem) *Chat {
	chat := &Chat{
		users:      map[string]bool{},
		prefix:     prefix,
		tubeSystem: tubeSystem,
	}

	tubeSystem.RegisterChannel(fmt.Sprintf("/%s/users", prefix), pts.ChannelHandlers{
		OnSubscribe:   chat.onUserJoin,
		OnUnsubscribe: chat.onUserLeave,
	})

	tubeSystem.RegisterChannel(fmt.Sprintf("/%s", prefix), pts.ChannelHandlers{
		OnMessage: chat.onChatMessage,
	})

	return chat
}

func (c *Chat) broadcastUsers(s *pts.Context) {
	payload, _ := json.Marshal(c.users)
	s.Broadcast(payload, &pts.ContextBroadcastOptions{
		ExcludeContextOwner: false,
	})
}

func (c *Chat) onChatMessage(s *pts.Context, message *pts.Message) {
	println("Received Message: " + s.FullPath)
	payload, _ := json.Marshal(fmt.Sprintf("%s: %s", s.Client.Id, message.Payload))
	s.Broadcast(payload, &pts.ContextBroadcastOptions{
		ExcludeContextOwner: false,
	})
}

func (c *Chat) onUserJoin(s *pts.Context) {
	println("Client joined: " + s.FullPath)
	c.users[s.Client.Id] = true
	c.broadcastUsers(s)
}

func (c *Chat) onUserLeave(s *pts.Context) {
	println("Client left: " + s.FullPath)
	c.users[s.Client.Id] = false
	c.broadcastUsers(s)
}
