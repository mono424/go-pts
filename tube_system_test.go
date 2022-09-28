package pts

import (
	"encoding/json"
	"testing"
)

type FakeSocketHandleConnectFunc func(session *FakeSocketSession)
type FakeSocketHandleDisconnectFunc func(session *FakeSocketSession)
type FakeSocketHandleMessageFunc func(session *FakeSocketSession, msg []byte)
type FakeSocketHandleOutgoingMessageFunc func(msg []byte)

type FakeSocket struct {
	handleConnect    FakeSocketHandleConnectFunc
	handleDisconnect FakeSocketHandleDisconnectFunc
	handleMessage    FakeSocketHandleMessageFunc
}

// NewClientConnects simulate a new client connects from the frontend.
// Messages that would be received by the frontend are passed to outgoingMessageHandler.
func (f *FakeSocket) NewClientConnects(outgoingMessageHandler FakeSocketHandleOutgoingMessageFunc) *FakeSocketSession {
	s := &FakeSocketSession{
		onDisconnect:      f.handleDisconnect,
		onMessage:         f.handleMessage,
		onOutgoingMessage: outgoingMessageHandler,
	}
	f.handleConnect(s)
	return s
}

type FakeSocketSession struct {
	Id                string
	onDisconnect      FakeSocketHandleDisconnectFunc
	onMessage         FakeSocketHandleMessageFunc
	onOutgoingMessage FakeSocketHandleOutgoingMessageFunc
}

// / Send emulate a data message from the frontend
func (s *FakeSocketSession) Send(data []byte) {
	s.onMessage(s, data)
}

// Disconnect emulate disconnecting initiated by the frontend
func (s *FakeSocketSession) Disconnect() {
	s.onDisconnect(s)
}

func NewFakeConnector(errorHandler ErrorHandlerFunc) (*Connector, *FakeSocket) {
	fakeSocket := &FakeSocket{}

	connector := NewConnector(nil, errorHandler)
	connector.clients.init()

	fakeSocket.handleConnect = func(s *FakeSocketSession) {
		client := Client{
			Id: "",
			sendMessage: func(msg []byte) error {
				s.onOutgoingMessage(msg)
				return nil
			},
			properties: map[string]interface{}{},
		}

		connector.clients.Join(&client)
		s.Id = client.Id

		if connector.hooks.OnConnect != nil {
			connector.hooks.OnConnect(&client)
		}
	}

	fakeSocket.handleDisconnect = func(s *FakeSocketSession) {
		client := connector.clients.Get(s.Id)
		if connector.hooks.OnDisconnect != nil {
			connector.hooks.OnDisconnect(client)
		}
		connector.clients.Remove(client.Id)
	}

	fakeSocket.handleMessage = func(s *FakeSocketSession, data []byte) {
		client := connector.clients.Get(s.Id)
		if connector.hooks.OnMessage != nil {
			connector.hooks.OnMessage(client, data)
		}
	}

	return connector, fakeSocket
}

func SubMessage(path string) []byte {
	message := Message{
		Type:    MessageTypeSubscribe,
		Channel: path,
		Payload: nil,
	}
	data, _ := json.Marshal(message)
	return data
}

func ChannelMessage(path string, payload json.RawMessage) []byte {
	message := Message{
		Type:    MessageTypeChannelMessage,
		Channel: path,
		Payload: payload,
	}
	data, _ := json.Marshal(message)
	return data
}

func UnsubMessage(path string) []byte {
	message := Message{
		Type:    MessageTypeUnsubscribe,
		Channel: path,
		Payload: nil,
	}
	data, _ := json.Marshal(message)
	return data
}

func TestTubeSystemConnection(t *testing.T) {

	t.Run("Simple Connect Disconnect", func(t *testing.T) {
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})

		fakeClient := fakeSocket.NewClientConnects(func(_ []byte) {})

		if len(fakeConnector.clients.clients) != 1 {
			t.Errorf("len(fakeConnector.clients.clients) = %d, want %d", len(fakeConnector.clients.clients), 1)
			return
		}

		if fakeConnector.clients.Get(fakeClient.Id) == nil {
			t.Errorf("fakeConnector.clients.Get(fakeClient.Id) = nil, want *FakeSocketSession")
		}

		fakeClient.Disconnect()

		if len(fakeConnector.clients.clients) != 0 {
			t.Errorf("len(fakeConnector.clients.clients) = %d, want %d", len(fakeConnector.clients.clients), 0)
			return
		}

	})

	t.Run("IsConnected should return connection status", func(t *testing.T) {
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		fakeClient := fakeSocket.NewClientConnects(func(_ []byte) {})

		if !tubeSystem.IsConnected(fakeClient.Id) {
			t.Errorf("tubeSystem.IsConnected(fakeClient.Id) = false, want true")
			return
		}

		fakeClient.Disconnect()

		if tubeSystem.IsConnected(fakeClient.Id) {
			t.Errorf("channel.IsConnected(fakeClient.Id, testChannelPath) = true, want false")
			return
		}
	})

	t.Run("Handle Sub/Unsub", func(t *testing.T) {
		channelPath := "example/path/:var"
		testVar := "test"
		testChannelPath := "example/path/" + testVar
		testChannelPath2 := "example/path/foobar"
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		var subContext *Context
		var unsubContext *Context

		channel := tubeSystem.RegisterChannel(channelPath, ChannelHandlers{
			OnSubscribe: func(s *Context) {
				subContext = s
			},
			OnUnsubscribe: func(s *Context) {
				unsubContext = s
			},
		})

		fakeClient := fakeSocket.NewClientConnects(func(_ []byte) {})
		fakeClient.Send(SubMessage(testChannelPath))

		if !tubeSystem.IsSubscribed(testChannelPath, fakeClient.Id) {
			t.Errorf("tubeSystem.IsSubscribed(testChannelPath, fakeClient.Id) = false, want true")
			return
		}
		if tubeSystem.IsSubscribed(channelPath, fakeClient.Id) {
			t.Errorf("tubeSystem.IsSubscribed(channelPath, fakeClient.Id) = true, want false")
			return
		}
		if tubeSystem.IsSubscribed(testChannelPath2, fakeClient.Id) {
			t.Errorf("tubeSystem.IsSubscribed(testChannelPath2, fakeClient.Id) = true, want false")
			return
		}

		if !channel.IsSubscribed(fakeClient.Id, testChannelPath) {
			t.Errorf("channel.IsSubscribed(fakeClient.Id, testChannelPath) = false, want true")
			return
		}
		if channel.IsSubscribed(fakeClient.Id, channelPath) {
			t.Errorf("channel.IsSubscribed(fakeClient.Id, channelPath) = true, want false")
			return
		}
		if channel.IsSubscribed(fakeClient.Id, testChannelPath2) {
			t.Errorf("channel.IsSubscribed(fakeClient.Id, testChannelPath2) = true, want false")
			return
		}

		if subContext == nil {
			t.Errorf("subContext = nil, want *Context")
			return
		}
		if subContext.FullPath != testChannelPath {
			t.Errorf("subContext = %s, want %s", subContext.FullPath, testChannelPath)
			return
		}

		fakeClient.Send(UnsubMessage(testChannelPath))

		if channel.IsSubscribed(fakeClient.Id, testChannelPath) {
			t.Errorf("channel.IsSubscribed(fakeClient.Id, testChannelPath) = true, want false")
			return
		}
		if unsubContext == nil {
			t.Errorf("unsubContext = nil, want *Context")
			return
		}

	})

}

func TestTubeSystemMessaging(t *testing.T) {

	t.Run("Client Sends Message", func(t *testing.T) {
		testChannelPath := "example/path/foobar"
		testPayload := map[string]interface{}{"name": "Jon Doe", "admin": false}
		testPayloadJson, _ := json.Marshal(testPayload)
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		var receivedMessage *Message

		tubeSystem.RegisterChannel(testChannelPath, ChannelHandlers{
			OnMessage: func(s *Context, message *Message) {
				receivedMessage = message
			},
		})

		fakeClient := fakeSocket.NewClientConnects(func(_ []byte) {})
		fakeClient.Send(SubMessage(testChannelPath))
		fakeClient.Send(ChannelMessage(testChannelPath, testPayloadJson))

		var receivedPayload map[string]interface{}
		_ = json.Unmarshal(receivedMessage.Payload, &receivedPayload)
		if receivedPayload["name"] != testPayload["name"] || receivedPayload["admin"] != testPayload["admin"] {
			t.Errorf(`equal({ "name": "%s", "admin": %b }, { "name": "%s", "admin": %b }) = false, want true`, receivedPayload["name"], receivedPayload["admin"], testPayload["name"], testPayload["admin"])
			return
		}
	})

	t.Run("Client sends Messages to different channels", func(t *testing.T) {
		channelA := "example/path/a"
		channelB := "example/path/b"
		payloadA := map[string]interface{}{"name": "Jon Doe", "admin": false}
		payloadB := map[string]interface{}{"name": "Tom Ford", "admin": true}
		payloadJsonA, _ := json.Marshal(payloadA)
		payloadJsonB, _ := json.Marshal(payloadB)
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		var recMessageChannelA *Message
		var recMessageChannelB *Message

		tubeSystem.RegisterChannel(channelA, ChannelHandlers{
			OnMessage: func(s *Context, message *Message) {
				recMessageChannelA = message
			},
		})

		tubeSystem.RegisterChannel(channelB, ChannelHandlers{
			OnMessage: func(s *Context, message *Message) {
				recMessageChannelB = message
			},
		})

		fakeClient := fakeSocket.NewClientConnects(func(msg []byte) {

		})
		fakeClient.Send(SubMessage(channelA))
		fakeClient.Send(SubMessage(channelB))
		fakeClient.Send(ChannelMessage(channelB, payloadJsonB))
		fakeClient.Send(ChannelMessage(channelA, payloadJsonA))

		var receivedPayload map[string]interface{}
		_ = json.Unmarshal(recMessageChannelA.Payload, &receivedPayload)
		if receivedPayload["name"] != payloadA["name"] || receivedPayload["admin"] != payloadA["admin"] {
			t.Errorf(`equal({ "name": "%s", "admin": %b }, { "name": "%s", "admin": %b }) = false, want true`, receivedPayload["name"], receivedPayload["admin"], payloadA["name"], payloadA["admin"])
			return
		}

		_ = json.Unmarshal(recMessageChannelB.Payload, &receivedPayload)
		if receivedPayload["name"] != payloadB["name"] || receivedPayload["admin"] != payloadB["admin"] {
			t.Errorf(`equal({ "name": "%s", "admin": %b }, { "name": "%s", "admin": %b }) = false, want true`, receivedPayload["name"], receivedPayload["admin"], payloadB["name"], payloadB["admin"])
			return
		}
	})

	t.Run("Client Receives Message", func(t *testing.T) {
		testChannelPath := "example/path/foobar"
		testPayload := map[string]interface{}{"name": "Jon Doe", "admin": false}
		testPayloadJson, _ := json.Marshal(testPayload)
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		var receivedMessage *Message

		tubeSystem.RegisterChannel(testChannelPath, ChannelHandlers{})

		fakeClient := fakeSocket.NewClientConnects(func(msg []byte) {
			if err := json.Unmarshal(msg, &receivedMessage); err != nil {
				t.Errorf("could not unmarshal message: %e", err)
			}
		})
		fakeClient.Send(SubMessage(testChannelPath))
		if err := tubeSystem.Send(testChannelPath, fakeClient.Id, testPayloadJson); err != nil {
			t.Errorf("tubeSystem.Send(...) throws an error: %s", err.Description)
			return
		}

		if receivedMessage.Channel != testChannelPath {
			t.Errorf("receivedMessage.Channel = %s, want %s", receivedMessage.Channel, testChannelPath)
			return
		}

		var receivedPayload map[string]interface{}
		_ = json.Unmarshal(receivedMessage.Payload, &receivedPayload)
		if receivedPayload["name"] != testPayload["name"] || receivedPayload["admin"] != testPayload["admin"] {
			t.Errorf(`equal({ "name": "%s", "admin": %b }, { "name": "%s", "admin": %b }) = false, want true`, receivedPayload["name"], receivedPayload["admin"], testPayload["name"], testPayload["admin"])
			return
		}
	})

	t.Run("TubeSystem send errors", func(t *testing.T) {
		testChannelPath := "example/path/foobar"
		testPayload := map[string]interface{}{"name": "Jon Doe", "admin": false}
		testPayloadJson, _ := json.Marshal(testPayload)
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {})
		tubeSystem := New(fakeConnector)

		var receivedMessage *Message

		tubeSystem.RegisterChannel(testChannelPath, ChannelHandlers{})

		fakeClient := fakeSocket.NewClientConnects(func(msg []byte) {
			if err := json.Unmarshal(msg, &receivedMessage); err != nil {
				t.Errorf("could not unmarshal message: %e", err)
			}
		})

		fakeClient.Send(SubMessage(testChannelPath))

		if err := tubeSystem.Send(testChannelPath, fakeClient.Id, testPayloadJson); err != nil {
			t.Errorf("tubeSystem.Send(...) throws an error: %s", err.Description)
			return
		}

		channelNotExistingErr := tubeSystem.Send("example/non-existant/foobar", fakeClient.Id, testPayloadJson)
		if channelNotExistingErr == nil {
			t.Errorf("tubeSystem.Send(...) return nil, want channel not exists error.")
			return
		} else if channelNotExistingErr.Code != ErrorUnknownChannel {
			t.Errorf("tubeSystem.Send(...) return Error{Code: %d}, want Error{Code: %d}.", channelNotExistingErr.Code, ErrorUnknownChannel)
			return
		}

		fakeClient.Send(UnsubMessage(testChannelPath))

		notSubscribedErr := tubeSystem.Send(testChannelPath, fakeClient.Id, testPayloadJson)
		if notSubscribedErr == nil {
			t.Errorf("tubeSystem.Send(...) return nil, want not subscribed error.")
			return
		} else if notSubscribedErr.Code != ErrorClientNotSubscribed {
			t.Errorf("tubeSystem.Send(...) return Error{Code: %d}, want Error{Code: %d}.", channelNotExistingErr.Code, ErrorClientNotSubscribed)
			return
		}
	})

	t.Run("Client send errors", func(t *testing.T) {
		testChannelPath := "example/path/foobar"
		testPayload := map[string]interface{}{"name": "Jon Doe", "admin": false}
		testPayloadJson, _ := json.Marshal(testPayload)

		var retrievedErr *Error
		fakeConnector, fakeSocket := NewFakeConnector(func(err *Error) {
			retrievedErr = err
		})
		tubeSystem := New(fakeConnector)

		var receivedMessage *Message

		tubeSystem.RegisterChannel(testChannelPath, ChannelHandlers{})

		fakeClient := fakeSocket.NewClientConnects(func(msg []byte) {
			if err := json.Unmarshal(msg, &receivedMessage); err != nil {
				t.Errorf("could not unmarshal message: %e", err)
			}
		})

		fakeClient.Send(SubMessage(testChannelPath))

		if err := tubeSystem.Send(testChannelPath, fakeClient.Id, testPayloadJson); err != nil {
			t.Errorf("tubeSystem.Send(...) throws an error: %s", err.Description)
			return
		}

		invalidTypeMessage := Message{
			Type:    "_INVALID_TYPE_",
			Channel: testChannelPath,
			Payload: nil,
		}
		data, _ := json.Marshal(invalidTypeMessage)
		fakeClient.Send(data)

		if retrievedErr == nil {
			t.Errorf("tubeSystem.messageHandler(...) did not throw any error, want invalid type error.")
		} else if retrievedErr.Code != ErrorUnknownType {
			t.Errorf("tubeSystem.messageHandler(...) return Error{Code: %d}, want Error{Code: %d}.", retrievedErr.Code, ErrorUnknownType)
			return
		}

		fakeClient.Send([]byte{1, 2, 3})
		if retrievedErr == nil {
			t.Errorf("tubeSystem.messageHandler(...) did not throw any error, want invalid message error.")
		} else if retrievedErr.Code != ErrorInvalidMessage {
			t.Errorf("tubeSystem.messageHandler(...) return Error{Code: %d}, want Error{Code: %d}.", retrievedErr.Code, ErrorInvalidMessage)
			return
		}

	})

}
