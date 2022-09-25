package pts

import (
	"errors"
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
		testKeyNotExistant := "fooBarBarFoo"
		testVal := "barFoo"
		testContext := &Context{
			properties: map[string]interface{}{},
		}

		testContext.Set(testKey, testVal)
		if val, _ := testContext.Get(testKey); val != testVal {
			t.Errorf("testContext.Get(%s) = %s, want %s", testKey, val, testVal)
		}

		if val, exists := testContext.Get(testKeyNotExistant); exists != false || val != nil {
			valStr := "nil"
			if val != nil {
				valStr = "interface{}"
			}

			boolStr := "false"
			if exists {
				boolStr = "true"
			}

			t.Errorf("testContext.Get(%s) = (%s, %s), want (nil, false)", testKeyNotExistant, valStr, boolStr)
		}
	})

	t.Run("Get/Set properties", func(t *testing.T) {
		testKeyNotExistant := "fooBarXX"
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

		if val := testContext.Param(testKeyNotExistant); val != "" {
			t.Errorf("testContext.Param(%s) != %s, want \"\"", testKeyNotExistant, val)
		}
	})

	t.Run("Send Client error", func(t *testing.T) {
		testContext := &Context{
			properties: map[string]interface{}{},
			Client: &Client{Id: "ABC123", sendMessage: func(message []byte) error {
				return errors.New("something failed")
			}},
		}

		if err := testContext.Send([]byte{1, 2, 3}); err == nil {
			t.Errorf("testContext.Send(...) returns no error, want error")
		} else if err.Code != ErrorSendingMessageFailed {
			t.Errorf("testContext.Send(...) returns Error{Code: %d}, want Error{Code: %d}", err.Code, ErrorSendingMessageFailed)
		}
	})

}
