package pts

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func contains(s []*Context, e *Context) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestChannelPathMatch(t *testing.T) {
	t.Run("Simple Case", func(t *testing.T) {
		simplePath := []string{"example", "path", "simple"}
		channel := Channel{
			path:        simplePath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}

		tooShortPath := strings.Join(simplePath[0:1], channelPathSep)
		if match, _ := channel.PathMatches(tooShortPath); match {
			t.Errorf("channel.PathMatches(%s) = (true, [...]), want (false, [...])", tooShortPath)
		}

		pathString := strings.Join(simplePath, channelPathSep)
		if match, _ := channel.PathMatches(pathString); !match {
			t.Errorf("channel.PathMatches(%s) = (false, [...]), want (true, [...])", pathString)
		}

		pathString = pathString + "2"
		if match, _ := channel.PathMatches(pathString); match {
			t.Errorf("channel.PathMatches(%s) = (true, [...]), want (false, [...])", pathString)
		}
	})

	t.Run("With Params", func(t *testing.T) {
		variablePath := []string{"example", ":var1", "path", ":var2", ":var3"}
		var1 := "foo"
		var2 := "bar"
		var3 := "blah123"
		validPath := "example/" + var1 + "/path/" + var2 + "/" + var3
		invalidPath := "example/" + var1 + "/pathX/" + var2 + "/" + var3

		channel := Channel{
			path:        variablePath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}

		if match, vars := channel.PathMatches(validPath); match {
			if vars["var1"] != var1 {
				t.Errorf("channel.PathMatches(%s) = (true, { \"var1\": \"%s\", [...] }), want (true, { \"var1\": \"%s\", [...] })", validPath, vars["var1"], var1)
			}
			if vars["var2"] != var2 {
				t.Errorf("channel.PathMatches(%s) = (true, { \"var2\": \"%s\", [...] }), want (true, { \"var2\": \"%s\", [...] })", validPath, vars["var2"], var2)
			}
			if vars["var3"] != var3 {
				t.Errorf("channel.PathMatches(%s) = (true, { \"var3\": \"%s\", [...] }), want (true, { \"var3\": \"%s\", [...] })", validPath, vars["var3"], var3)
			}
		} else {
			t.Errorf("channel.PathMatches(%s) = (false, [...]), want (true, [...])", validPath)
		}

		if match, _ := channel.PathMatches(invalidPath); match {
			t.Errorf("channel.PathMatches(%s) = (true, [...]), want (false, [...])", invalidPath)
		}
	})

	t.Run("With Wildcard", func(t *testing.T) {
		variablePath := []string{"example", "*", "path"}
		validPath1 := "example/var1/path"
		validPath2 := "example/*/path"
		invalidPath1 := "example//path"
		invalidPath2 := "example/var1/pathX"

		channel := Channel{
			path:        variablePath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}

		if match, _ := channel.PathMatches(validPath1); !match {
			t.Errorf("channel.PathMatches(%s) = (false, [...]), want (true, [...])", validPath1)
		}

		if match, _ := channel.PathMatches(validPath2); !match {
			t.Errorf("channel.PathMatches(%s) = (false, [...]), want (true, [...])", validPath2)
		}

		if match, _ := channel.PathMatches(invalidPath1); match {
			t.Errorf("channel.PathMatches(%s) = (true, [...]), want (false, [...])", invalidPath1)
		}

		if match, _ := channel.PathMatches(invalidPath2); match {
			t.Errorf("channel.PathMatches(%s) = (true, [...]), want (false, [...])", invalidPath2)
		}
	})

	t.Run("Simple Subscribe", func(t *testing.T) {
		simplePath := []string{"example", "path", "simple"}
		testId := "ABC123"

		pathString := strings.Join(simplePath, channelPathSep)
		testContext := &Context{
			FullPath: pathString,
			Client:   &Client{Id: testId},
		}

		channel := Channel{
			path:        simplePath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		channel.Subscribe(testContext)

		if !channel.IsSubscribed(testId, pathString) {
			t.Errorf("channel.IsSubscribed(%s, %s) = false, want true", testId, pathString)
		}
	})

	t.Run("Subscribe Middleware throws error", func(t *testing.T) {
		simplePath := []string{"example", "path", "simple"}
		testId := "ABC123"
		testErrCode := 999
		testErrDescription := "Unauthorized"

		var onErrResult *Error
		var errMessage map[string]interface{}

		pathString := strings.Join(simplePath, channelPathSep)
		testContext := &Context{
			FullPath: pathString,
			Client: &Client{Id: testId, sendMessage: func(message []byte) error {
				_ = json.Unmarshal(message, &errMessage)
				return nil
			}},
		}

		channel := Channel{
			path: simplePath,
			handlers: ChannelHandlers{
				SubscriptionMiddlewares: []SubscriptionMiddleware{
					func(s *Context) *Error {
						return NewError(s, testErrCode, testErrDescription, nil)
					},
				},
			},
			subscribers: ChannelSubscribers{},
			onError: func(e *Error) {
				onErrResult = e
			},
		}
		channel.subscribers.init()

		channel.Subscribe(testContext)

		if channel.IsSubscribed(testId, pathString) {
			t.Errorf("channel.IsSubscribed(%s, %s) = true, want false", testId, pathString)
		}

		if onErrResult == nil {
			t.Errorf("onErr was not called, want onErr to be called")
		} else if onErrResult.Code != testErrCode || onErrResult.Description != testErrDescription {
			t.Errorf("onErr was called with Error{Code=%d, Description=%s}, want Error{Code=%d, Description=%s}", onErrResult.Code, onErrResult.Description, testErrCode, testErrDescription)
		}

		if errMessage == nil {
			t.Errorf("sendMessage was not called, want sendMessage to be called")
		} else if errMessage["type"] != MessageTypeChannelMessage || errMessage["channel"] != testContext.FullPath {
			t.Errorf("sendMessage was called with {type: %s, channel: %s}, want {type: %s, channel: %s}", errMessage["type"], errMessage["channel"], MessageTypeChannelMessage, testContext.FullPath)
		} else if errMessage["payload"] == nil {
			t.Errorf("sendMessage was called with message.payload = nil, want message.payload != nil")
		}

		errPayload := errMessage["payload"].(map[string]interface{})
		if int(errPayload["code"].(float64)) != testErrCode || errPayload["description"] != testErrDescription {
			t.Errorf("sendMessage was called with {payload: {code: %d, description: %s}}, want {payload: {code: %d, description: %s}}", errPayload["code"], errPayload["description"], testErrCode, testErrDescription)
		}
	})

	t.Run("Subscribe Middleware & client throws error", func(t *testing.T) {
		simplePath := []string{"example", "path", "simple"}
		testId := "ABC123"
		testErrCode := 999
		testErrDescription := "Unauthorized"

		var onErrResults []*Error

		pathString := strings.Join(simplePath, channelPathSep)
		testContext := &Context{
			FullPath: pathString,
			Client: &Client{Id: testId, sendMessage: func(message []byte) error {
				return errors.New("disconnected")
			}},
		}

		channel := Channel{
			path: simplePath,
			handlers: ChannelHandlers{
				SubscriptionMiddlewares: []SubscriptionMiddleware{
					func(s *Context) *Error {
						return NewError(s, testErrCode, testErrDescription, nil)
					},
				},
			},
			subscribers: ChannelSubscribers{},
			onError: func(e *Error) {
				onErrResults = append(onErrResults, e)
			},
		}
		channel.subscribers.init()

		channel.Subscribe(testContext)

		if channel.IsSubscribed(testId, pathString) {
			t.Errorf("channel.IsSubscribed(%s, %s) = true, want false", testId, pathString)
		}

		if len(onErrResults) < 2 {
			t.Errorf("onErr was not called twice, want onErr to be called twice")
		}

		if onErrResults[0].Code != testErrCode || onErrResults[0].Description != testErrDescription {
			t.Errorf("onErr was called with Error{Code=%d, Description=%s}, want Error{Code=%d, Description=%s}", onErrResults[0].Code, onErrResults[0].Description, testErrCode, testErrDescription)
		}

		if onErrResults[1].Code != ErrorSendingErrorFailed {
			t.Errorf("onErr was called with Error{Code=%d}, want Error{Code=%d}", onErrResults[1].Code, testErrCode)
		}
	})

	t.Run("Unsubscribe all", func(t *testing.T) {
		testPath := []string{"example", "path", ":var"}
		testParams := []string{"foo", "bar", "var"}
		testId := "ABC123"
		testClient := &Client{Id: testId, sendMessage: func(message []byte) error {
			return nil
		}}

		var testContexts []*Context

		for _, v := range testParams {
			pathString := strings.Join(append(testPath[0:2], v), channelPathSep)
			testContexts = append(testContexts, &Context{
				FullPath: pathString,
				Client:   testClient,
			})
		}

		var unsubContexts []*Context

		channel := Channel{
			path: testPath,
			handlers: ChannelHandlers{
				OnUnsubscribe: func(s *Context) {
					unsubContexts = append(unsubContexts, s)
				},
			},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		for _, context := range testContexts {
			channel.Subscribe(context)
		}

		for _, context := range testContexts {
			if !channel.IsSubscribed(testId, context.FullPath) {
				t.Errorf("channel.IsSubscribed(%s, %s) = false, want true", testId, context.FullPath)
			}
		}

		if !channel.UnsubscribeAllPaths(testId) {
			t.Errorf("channel.UnsubscribeAllPaths(%s) = false, want true", testId)
		}

		for _, context := range testContexts {
			if channel.IsSubscribed(testId, context.FullPath) {
				t.Errorf("channel.IsSubscribed(%s, %s) = true, want false", testId, context.FullPath)
			}
			if !contains(unsubContexts, context) {
				t.Errorf("OnUnsubscribe was not called with all correct contexts")
			}
		}
	})

	t.Run("Optional Message Handler should not panic", func(t *testing.T) {
		testPath := []string{"example", "path"}
		pathString := strings.Join(testPath, channelPathSep)
		testClient := &Client{Id: "ABC123", sendMessage: func(message []byte) error {
			return nil
		}}

		channel := Channel{
			path:        testPath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}

		// should not panic
		channel.HandleMessage(testClient, &Message{
			Type:    MessageTypeChannelMessage,
			Channel: pathString,
			Payload: json.RawMessage{},
		})

		if res := channel.Unsubscribe(testClient.Id, pathString); res != false {
			t.Errorf("channel.Unsubscribe(%s, %s) = true, want false", testClient.Id, pathString)
		}
	})

	t.Run("Unsubscribe should return false on non-subscriber", func(t *testing.T) {
		testPath := []string{"example", "path"}
		pathString := strings.Join(testPath, channelPathSep)
		testClient := &Client{Id: "ABC123", sendMessage: func(message []byte) error {
			return nil
		}}

		channel := Channel{
			path:        testPath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}

		if res := channel.Unsubscribe(testClient.Id, pathString); res != false {
			t.Errorf("channel.Unsubscribe(%s, %s) = true, want false", testClient.Id, pathString)
		}
	})

	t.Run("Get subscribers should return subscribers for path", func(t *testing.T) {
		testPath := []string{"example", "path", ":var"}
		testContexts := map[string][]*Context{}

		pathA := strings.Join(append(testPath[0:2], "foo"), channelPathSep)
		pathB := strings.Join(append(testPath[0:2], "bar"), channelPathSep)

		testContexts[pathA] = []*Context{}
		testContexts[pathB] = []*Context{}

		client1Id := "ABC123"
		client1Paths := []string{
			pathA,
			pathB,
		}
		testClient := &Client{Id: client1Id, sendMessage: func(message []byte) error {
			return nil
		}}

		client2Id := "ABC321"
		client2Paths := []string{
			pathA,
		}
		testClient2 := &Client{Id: client2Id, sendMessage: func(message []byte) error {
			return nil
		}}

		for _, path := range client1Paths {
			testContexts[path] = append(testContexts[path], &Context{
				FullPath: path,
				Client:   testClient,
			})
		}

		for _, path := range client2Paths {
			testContexts[path] = append(testContexts[path], &Context{
				FullPath: path,
				Client:   testClient2,
			})
		}

		channel := Channel{
			path:        testPath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		for _, contexts := range testContexts {
			for _, context := range contexts {
				channel.Subscribe(context)
			}
		}

		contextsA := channel.GetSubscribers(pathA)
		if len(contextsA) != len(testContexts[pathA]) {
			t.Errorf("channel.GetSubscribers(pathA) returned %d items, want %d items.", len(contextsA), len(testContexts[pathA]))
			return
		}
		for _, context := range testContexts[pathA] {
			if !contains(contextsA, context) {
				t.Errorf("channel.GetSubscribers(pathA) is missing a context.")
				return
			}
		}

		contextsB := channel.GetSubscribers(pathB)
		if len(contextsB) != len(testContexts[pathB]) {
			t.Errorf("channel.GetSubscribers(pathB) returned %d items, want %d items.", len(contextsB), len(testContexts[pathB]))
			return
		}
		for _, context := range testContexts[pathB] {
			if !contains(contextsB, context) {
				t.Errorf("channel.GetSubscribers(pathB) is missing a context.")
				return
			}
		}

	})

	t.Run("Get subscribers should return subscribers for path", func(t *testing.T) {
		testPath := []string{"example", "path", ":var"}
		var testContexts []*Context

		pathA := strings.Join(append(testPath[0:2], "foo"), channelPathSep)
		pathB := strings.Join(append(testPath[0:2], "bar"), channelPathSep)
		pathC := strings.Join(append(testPath[0:2], "wow"), channelPathSep)

		client1Id := "ABC123"
		client1Paths := []string{
			pathA,
			pathB,
		}
		testClient := &Client{Id: client1Id, sendMessage: func(message []byte) error {
			return nil
		}}

		client2Id := "ABC321"
		client2Paths := []string{
			pathC,
		}
		testClient2 := &Client{Id: client2Id, sendMessage: func(message []byte) error {
			return nil
		}}

		client3Id := "ABC010"
		client3Paths := []string{
			pathA,
			pathC,
		}
		testClient3 := &Client{Id: client3Id, sendMessage: func(message []byte) error {
			return nil
		}}

		for _, path := range client1Paths {
			testContexts = append(testContexts, &Context{
				FullPath: path,
				Client:   testClient,
			})
		}

		for _, path := range client2Paths {
			testContexts = append(testContexts, &Context{
				FullPath: path,
				Client:   testClient2,
			})
		}

		for _, path := range client3Paths {
			testContexts = append(testContexts, &Context{
				FullPath: path,
				Client:   testClient3,
			})
		}

		channel := Channel{
			path:        testPath,
			handlers:    ChannelHandlers{},
			subscribers: ChannelSubscribers{},
		}
		channel.subscribers.init()

		for _, context := range testContexts {
			channel.Subscribe(context)
		}

		contextsA := channel.GetAllSubscribers()
		if len(contextsA) != len(testContexts) {
			t.Errorf("channel.GetAllSubscribers() returned %d items, want %d items.", len(contextsA), len(testContexts))
			return
		}
		for _, context := range testContexts {
			if !contains(contextsA, context) {
				t.Errorf("channel.GetSubscribers(pathA) is missing a context.")
				return
			}
		}

	})

	t.Run("Broadcast should correctly call sendMessage for clients", func(t *testing.T) {
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
		errorCount := 0

		testClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				sendCount++
				return nil
			},
		}

		errorClient := &Client{
			Id: "DEF456",
			sendMessage: func(message []byte) error {
				errorCount++
				return errors.New("failed to send message")
			},
		}

		skipClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		channel.Subscribe(&Context{FullPath: testPath, Client: testClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: errorClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: skipClient})

		options := &ChannelBroadcastOptions{
			SkipClientIds: []string{"XYZ789"},
		}

		// when
		channel.Broadcast(testPath, payload, options)

		// then
		if sendCount != 1 {
			t.Errorf("Expected sendMessage to be called once for testClient, was called %d times", sendCount)
		}
		if errorCount != 1 {
			t.Errorf("Expected sendMessage to be called once for errorClient, was called %d times", sendCount)
		}
	})

	t.Run("Broadcast should not call sendMessage for skipped clients", func(t *testing.T) {
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

		skipSendCount := 0

		testClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		errorClient := &Client{
			Id: "DEF456",
			sendMessage: func(message []byte) error {
				return errors.New("failed to send message")
			},
		}

		skipClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				skipSendCount++
				return nil
			},
		}

		channel.Subscribe(&Context{FullPath: testPath, Client: testClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: errorClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: skipClient})

		options := &ChannelBroadcastOptions{
			SkipClientIds: []string{"XYZ789"},
		}

		// when
		channel.Broadcast(testPath, payload, options)

		// then
		if skipSendCount != 0 {
			t.Errorf("Expected sendMessage not to be called for skipClient, was called %d times", skipSendCount)
		}
	})

	t.Run("Broadcast should return HasErrors true when there is an error", func(t *testing.T) {
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

		testClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		errorClient := &Client{
			Id: "DEF456",
			sendMessage: func(message []byte) error {
				return errors.New("failed to send message")
			},
		}

		skipClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		channel.Subscribe(&Context{FullPath: testPath, Client: testClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: errorClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: skipClient})

		options := &ChannelBroadcastOptions{
			SkipClientIds: []string{"XYZ789"},
		}

		// when
		result := channel.Broadcast(testPath, payload, options)

		// then
		if !result.HasErrors {
			t.Errorf("Expected HasErrors to be true, got false")
		}
	})

	t.Run("Broadcast should return correct number of results", func(t *testing.T) {
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

		testClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		errorClient := &Client{
			Id: "DEF456",
			sendMessage: func(message []byte) error {
				return errors.New("failed to send message")
			},
		}

		skipClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		channel.Subscribe(&Context{FullPath: testPath, Client: testClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: errorClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: skipClient})

		options := &ChannelBroadcastOptions{
			SkipClientIds: []string{"XYZ789"},
		}

		// when
		result := channel.Broadcast(testPath, payload, options)

		// then
		if len(result.Results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(result.Results))
		}
	})

	t.Run("Broadcast should return correct results for each client", func(t *testing.T) {
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

		testClient := &Client{
			Id: "ABC123",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		errorClient := &Client{
			Id: "DEF456",
			sendMessage: func(message []byte) error {
				return errors.New("failed to send message")
			},
		}

		skipClient := &Client{
			Id: "XYZ789",
			sendMessage: func(message []byte) error {
				return nil
			},
		}

		channel.Subscribe(&Context{FullPath: testPath, Client: testClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: errorClient})
		channel.Subscribe(&Context{FullPath: testPath, Client: skipClient})

		options := &ChannelBroadcastOptions{
			SkipClientIds: []string{"XYZ789"},
		}

		// when
		result := channel.Broadcast(testPath, payload, options)

		// then
		for _, res := range result.Results {
			if res.Context.Client.Id == testClient.Id && (res.Err != nil || res.Skipped) {
				t.Errorf("Unexpected result for client %s", testClient.Id)
			}

			if res.Context.Client.Id == errorClient.Id && (res.Err == nil || res.Skipped) {
				t.Errorf("Expected error for client %s", errorClient.Id)
			}

			if res.Context.Client.Id == skipClient.Id && (!res.Skipped || res.Err != nil) {
				t.Errorf("Expected skip for client %s", skipClient.Id)
			}
		}
	})

}
