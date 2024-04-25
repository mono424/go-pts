package pts

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Context struct {
	Client     *Client
	FullPath   string
	Channel    *Channel
	params     map[string]string
	properties map[string]interface{}
}

type ErrorHandlerFunc func(*Error)

const (
	ErrorInvalidMessage       = iota // ErrorInvalidMessage if an incoming message could not be parsed
	ErrorUnknownType                 // ErrorUnknownType if a message with an unknown type is received
	ErrorUnknownChannel              // ErrorUnknownChannel if a message to an unknown channel is received or sent
	ErrorClientNotSubscribed         // ErrorClientNotSubscribed if a message is sent through a channel that is not subscribed by the client
	ErrorSendingErrorFailed          // ErrorSendingErrorFailed if a error message could not be send to a client
	ErrorSendingMessageFailed        // ErrorSendingMessageFailed if a message could not be sent to a client
	ErrorMultipleErrors              // ErrorMultipleErrors multiple errors happened
)

type Error struct {
	Context     *Context `json:"-"`
	Code        int      `json:"code"`
	Description string   `json:"description"`
	Raw         error    `json:"-"`
	Errors      []*Error `json:"errors"`
}

func NewError(context *Context, code int, description string, err error) *Error {
	if err == nil {
		err = errors.New(description)
	}

	return &Error{
		Context:     context,
		Code:        code,
		Description: description,
		Raw:         err,
		Errors:      []*Error{},
	}
}

func NewMultiError(context *Context, description string, err error, errs []*Error) *Error {
	if err == nil {
		err = errors.New(description)
	}

	return &Error{
		Context:     context,
		Code:        ErrorMultipleErrors,
		Description: description,
		Raw:         err,
		Errors:      errs,
	}
}

func (context *Context) MustGet(key string) interface{} {
	if value, exists := context.Get(key); exists {
		return value
	}
	panic(fmt.Sprintf("Key '%s' does not exist", key))
}

func (context *Context) Get(key string) (value interface{}, exists bool) {
	if val, ok := context.properties[key]; ok {
		return val, ok
	}
	return nil, false
}

func (context *Context) Set(key string, value interface{}) {
	context.properties[key] = value
}

func (context *Context) SendError(error *Error) *Error {
	data, err := json.Marshal(error)
	if err != nil {
		return NewError(context, ErrorSendingErrorFailed, "failed to send error to client", err)
	}
	message := Message{
		Type:    MessageTypeChannelMessage,
		Channel: context.FullPath,
		Payload: data,
	}
	data, err = json.Marshal(message)
	if err != nil {
		return NewError(context, ErrorSendingErrorFailed, "failed to send error to client", err)
	}

	if err = context.Client.Send(data); err != nil {
		return NewError(context, ErrorSendingErrorFailed, "failed to send error to client", err)
	}
	return nil
}

func (context *Context) Send(payload []byte) *Error {
	message := Message{
		Type:    MessageTypeChannelMessage,
		Channel: context.FullPath,
		Payload: payload,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return NewError(context, ErrorSendingMessageFailed, "failed to send error to client", err)
	}
	if err = context.Client.Send(data); err != nil {
		return NewError(context, ErrorSendingMessageFailed, "failed to send error to client", err)
	}
	return nil
}

func (context *Context) SetParams(params map[string]string) {
	context.params = params
}

func (context *Context) Param(key string) string {
	return context.params[key]
}

type ContextBroadcastOptions struct {
	ExcludeContextOwner bool
}

func (context *Context) Broadcast(payload []byte, options *ContextBroadcastOptions) *ChannelBroadcastResult {
	opts := &ChannelBroadcastOptions{}
	if options != nil && options.ExcludeContextOwner {
		opts.SkipClientIds = []string{context.Client.Id}
	}
	return context.Channel.Broadcast(context.FullPath, payload, opts)
}
