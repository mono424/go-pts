package pts

import (
	"bytes"
	"net/http"
	"testing"
)

func TestConnector(t *testing.T) {
	t.Run("Join", func(t *testing.T) {
		testStrKey := "foo"
		testStrVal := "bar"
		testBoolKey := "bool"
		clientProperties := map[string]interface{}{
			testStrKey:  testStrVal,
			testBoolKey: true,
		}
		clientSendFunc := func(message []byte) error {
			return nil
		}

		connector := NewConnector(func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			return nil
		}, func(_ *Error) {})

		var joinedClient *Client

		connector.hook(&Hooks{
			OnConnect: func(client *Client) {
				joinedClient = client
			},
		})

		connector.Join(clientSendFunc, clientProperties)
		if joinedClient == nil {
			t.Errorf("OnConnect hook was not called, want OnConnect to be called.")
		} else if val, _ := joinedClient.Get(testStrKey); val != testStrVal {
			t.Errorf("joinedClient.Get(\"foo\") returns %s, want %s.", val, testStrVal)
			return
		} else if val, _ := joinedClient.Get(testBoolKey); val != true {
			boolStr := "false"
			if val.(bool) {
				boolStr = "true"
			}

			t.Errorf("joinedClient.Get(\"foo\") returns %s, want %s.", boolStr, "true")
			return
		}

	})

	t.Run("Message", func(t *testing.T) {
		testMsg := []byte{1, 2, 3}

		clientSendFunc := func(message []byte) error {
			return nil
		}

		connector := NewConnector(func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			return nil
		}, func(_ *Error) {})

		var joinedClient *Client
		var receivedClient *Client
		var receivedMessage []byte

		connector.hook(&Hooks{
			OnConnect: func(client *Client) {
				joinedClient = client
			},
			OnMessage: func(client *Client, msg []byte) {
				receivedClient = client
				receivedMessage = msg
			},
		})

		connector.Join(clientSendFunc, map[string]interface{}{})

		if joinedClient == nil {
			t.Errorf("OnConnect hook was not called, want OnConnect to be called.")
		}

		connector.Message(joinedClient.Id, testMsg)

		if receivedClient == nil || receivedMessage == nil {
			t.Errorf("OnMessage hook was not called, want OnMessage to be called with client and message.")
		} else if joinedClient != receivedClient {
			t.Errorf("OnMessage gets called with different client, want same client than on OnConnect call")
			return
		} else if bytes.Compare(receivedMessage, testMsg) != 0 {
			t.Errorf("receivedMessage %v, want %v.", receivedMessage, testMsg)
			return
		}
	})

	t.Run("Leave", func(t *testing.T) {
		clientSendFunc := func(message []byte) error {
			return nil
		}

		connector := NewConnector(func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			return nil
		}, func(_ *Error) {})

		var joinedClient *Client
		var leftClient *Client

		connector.hook(&Hooks{
			OnConnect: func(client *Client) {
				joinedClient = client
			},
			OnDisconnect: func(client *Client) {
				leftClient = client
			},
		})

		connector.Join(clientSendFunc, map[string]interface{}{})

		if joinedClient == nil {
			t.Errorf("OnConnect hook was not called, want OnConnect to be called.")
		}

		connector.Leave(joinedClient.Id)

		if leftClient == nil {
			t.Errorf("OnDisconnect hook was not called, want OnDisconnect to be called with client and message.")
		} else if joinedClient != leftClient {
			t.Errorf("OnDisconnect gets called with different client, want same client than on OnConnect call")
			return
		}
	})
}
