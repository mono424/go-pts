package pts

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("Init Error with go error", func(t *testing.T) {
		testContext := &Context{}
		testDescription := "Foo Bar Error"
		testCode := ErrorSendingErrorFailed
		testErr := errors.New("A error")

		err := NewError(testContext, testCode, testDescription, testErr)
		if err.Context != testContext {
			t.Errorf("err.Context = %p, want %p", err.Context, testContext)
		}
		if err.Code != testCode {
			t.Errorf("err.Code = %d, want %d", err.Code, testCode)
		}
		if err.Description != testDescription {
			t.Errorf("err.Description = %s, want %s", err.Description, testDescription)
		}
		if err.Raw != testErr {
			t.Errorf("err.Raw = %e, want %e", err.Raw, testErr)
		}
	})

	t.Run("Init Error without go error", func(t *testing.T) {
		testContext := &Context{}
		testDescription := "Foo Bar Error"
		testCode := ErrorSendingErrorFailed

		err := NewError(testContext, testCode, testDescription, nil)
		if err.Context != testContext {
			t.Errorf("err.Context = %p, want %p", err.Context, testContext)
		}
		if err.Code != testCode {
			t.Errorf("err.Code = %d, want %d", err.Code, testCode)
		}
		if err.Description != testDescription {
			t.Errorf("err.Description = %s, want %s", err.Description, testDescription)
		}
		if err.Raw.Error() != testDescription {
			t.Errorf("err.Raw = %s, want %s", err.Raw, testDescription)
		}
	})
}

func TestContext(t *testing.T) {

	t.Run("Get/Set properties", func(t *testing.T) {
		testKey := "fooBar"
		testKeyNotExistent := "fooBarBarFoo"
		testVal := "barFoo"
		testContext := &Context{
			properties: map[string]interface{}{},
		}

		testContext.Set(testKey, testVal)
		if val, _ := testContext.Get(testKey); val != testVal {
			t.Errorf("testContext.Get(%s) = %s, want %s", testKey, val, testVal)
		}

		if val, exists := testContext.Get(testKeyNotExistent); exists != false || val != nil {
			valStr := "nil"
			if val != nil {
				valStr = "interface{}"
			}

			boolStr := "false"
			if exists {
				boolStr = "true"
			}

			t.Errorf("testContext.Get(%s) = (%s, %s), want (nil, false)", testKeyNotExistent, valStr, boolStr)
		}
	})

	t.Run("MustGet property", func(t *testing.T) {
		testKey := "fooBar"
		testKeyNotExistent := "fooBarBarFoo"
		testVal := "barFoo"
		testContext := &Context{
			properties: map[string]interface{}{},
		}

		testContext.Set(testKey, testVal)
		if val := testContext.MustGet(testKey); val != testVal {
			t.Errorf("testContext.MustGet(%s) = %s, want %s", testKey, val, testVal)
		}

		defer func() { recover() }()
		testContext.MustGet(testKeyNotExistent)
		t.Errorf("testContext.MustGet(%s) did not panic, but it should have", testKeyNotExistent)
	})

	t.Run("Get/Set properties", func(t *testing.T) {
		testKeyNotExistent := "fooBarXX"
		testKey := "fooBar"
		testVal := "barFoo"
		testContext := &Context{
			properties: map[string]interface{}{},
		}
		testParams := map[string]string{
			testKey: testVal,
		}

		testContext.SetParams(testParams)
		if val := testContext.Param(testKey); val != testVal {
			t.Errorf("testContext.Param(%s) = %s, want %s", testKey, val, testVal)
		}

		if val := testContext.Param(testKeyNotExistent); val != "" {
			t.Errorf("testContext.Param(%s) != %s, want \"\"", testKeyNotExistent, val)
		}
	})

	t.Run("Send Client error", func(t *testing.T) {
		testContext := &Context{
			properties: map[string]interface{}{},
			Client: &Client{Id: "ABC123", sendMessage: func(message []byte) error {
				return errors.New("something failed")
			}},
		}

		payload, _ := json.Marshal(map[string]any{
			"foo": "bar",
		})

		if err := testContext.Send(payload); err == nil {
			t.Errorf("testContext.Send(...) returns no error, want error")
		} else if err.Code != ErrorSendingMessageFailed {
			t.Errorf("testContext.Send(...) returns Error{Code: %d}, want Error{Code: %d}", err.Code, ErrorSendingMessageFailed)
		}
	})

	t.Run("Broadcast should correctly exclude context owner if option is set", func(t *testing.T) {
		testPath := "example/path"
		payload, _ := json.Marshal(map[string]any{
			"foo": "bar",
		})
		opts := &ContextBroadcastOptions{
			ExcludeContextOwner: true,
		}

		channel := &Channel{
			path:        strings.Split(testPath, channelPathSep),
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		ownerClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				t.Error("sendMessage should not have been called for ownerClient")
				return nil
			},
		}

		otherClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		context := &Context{
			FullPath: testPath,
			Client:   ownerClient,
			Channel:  channel,
		}

		channel.Subscribe(context)
		channel.Subscribe(&Context{FullPath: testPath, Client: otherClient})

		// when
		context.Broadcast(payload, opts)

		// there's no explicit "then" step here because we're asserting in the sendMessage function
	})

	t.Run("Broadcast should not exclude context owner if option is not set", func(t *testing.T) {
		testPath := "example/path"
		payload, _ := json.Marshal(map[string]any{
			"foo": "bar",
		})
		opts := &ContextBroadcastOptions{
			ExcludeContextOwner: false,
		}

		channel := &Channel{
			path:        strings.Split(testPath, channelPathSep),
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		sendCount := 0
		ownerClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				sendCount++
				return nil
			},
		}

		otherClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		context := &Context{
			FullPath: testPath,
			Client:   ownerClient,
			Channel:  channel,
		}

		channel.Subscribe(context)
		channel.Subscribe(&Context{FullPath: testPath, Client: otherClient})

		// when
		context.Broadcast(payload, opts)

		// then
		if sendCount != 1 {
			t.Errorf("Expected sendMessage to be called once for ownerClient, was called %d times", sendCount)
		}
	})

	t.Run("Broadcast should not exclude context owner if options is nil", func(t *testing.T) {
		testPath := "example/path"
		payload, _ := json.Marshal(map[string]any{
			"foo": "bar",
		})

		channel := &Channel{
			path:        strings.Split(testPath, channelPathSep),
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		sendCount := 0
		ownerClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				sendCount++
				return nil
			},
		}

		otherClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		context := &Context{
			FullPath: testPath,
			Client:   ownerClient,
			Channel:  channel,
		}

		channel.Subscribe(context)
		channel.Subscribe(&Context{FullPath: testPath, Client: otherClient})

		// when
		context.Broadcast(payload, nil)

		// then
		if sendCount != 1 {
			t.Errorf("Expected sendMessage to be called once for ownerClient, was called %d times", sendCount)
		}
	})

}
