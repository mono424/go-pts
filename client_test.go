package pts

import (
	"testing"
)

func TestClient(t *testing.T) {
	t.Run("Simple New Client", func(t *testing.T) {
		properties := map[string]interface{}{
			"foo":  "bar",
			"bool": true,
		}

		client := NewClient(func(message []byte) error { return nil }, properties)

		if val, _ := client.Get("foo"); val != properties["foo"] {
			t.Errorf("client.Get(\"foo\") = %s, want %s", val, properties["foo"])
		}

		if val, _ := client.Get("bool"); val != properties["bool"] {
			t.Errorf("client.Get(\"foo\") = %b, want %b", val, properties["foo"])
		}
	})

	t.Run("Client Set/Get", func(t *testing.T) {
		testKey := "barFoo"
		testVal := "fooBar"

		client := NewClient(func(message []byte) error { return nil }, map[string]interface{}{})

		client.Set(testKey, testVal)

		if val, _ := client.Get(testKey); val != testVal {
			t.Errorf("client.Get(\"%s\") = %s, want %s", testKey, val, testVal)
		}
	})

	t.Run("Client MustGet", func(t *testing.T) {
		testKey := "barFoo"
		testKeyNotExistent := "barFooNotThere"
		testVal := "fooBar"

		client := NewClient(func(message []byte) error { return nil }, map[string]interface{}{})

		client.Set(testKey, testVal)

		if val := client.MustGet(testKey); val != testVal {
			t.Errorf("client.MustGet(\"%s\") = %s, want %s", testKey, val, testVal)
		}

		defer func() { recover() }()
		client.MustGet(testKeyNotExistent)
		t.Errorf("client.MustGet(%s) did not panic, but it should have", testKeyNotExistent)
	})

	t.Run("Client Get Missing", func(t *testing.T) {
		testKey := "barFoo"

		client := NewClient(func(message []byte) error { return nil }, map[string]interface{}{})

		if val, ok := client.Get(testKey); ok != false || val != nil {
			boolStr := "false"
			if ok {
				boolStr = "true"
			}
			t.Errorf("client.Get(\"%s\") = (%s, %s), want (false, nil)", testKey, val, boolStr)
		}
	})
}
