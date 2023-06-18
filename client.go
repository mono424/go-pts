package pts

import "fmt"

type MessageSendFunc func(message []byte) error

type Client struct {
	Id          string
	sendMessage MessageSendFunc
	properties  map[string]interface{}
}

func NewClient(sendMessage MessageSendFunc, properties map[string]interface{}) *Client {
	return &Client{
		Id:          "",
		sendMessage: sendMessage,
		properties:  properties,
	}
}

func (client *Client) Send(message []byte) error {
	return client.sendMessage(message)
}

func (client *Client) MustGet(key string) interface{} {
	if value, exists := client.Get(key); exists {
		return value
	}
	panic(fmt.Sprintf("Key '%s' does not exist", key))
}

func (client *Client) Get(key string) (value interface{}, exists bool) {
	if val, ok := client.properties[key]; ok {
		return val, ok
	}
	return nil, false
}

func (client *Client) Set(key string, value interface{}) {
	client.properties[key] = value
}
