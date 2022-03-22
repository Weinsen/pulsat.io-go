package pulsatio_client

import (
	"testing"
)

func TestGetData(t *testing.T) {

	p := New("01", "https://127.0.0.1")

	var tests = []struct {
		key string
		val string
	}{
		{"id", "01"},
		{"name", ""},
	}

	for _, tst := range tests {
		t.Run(tst.key, func(t *testing.T) {
			val := p.GetData(tst.key)
			if tst.val != val {
				t.Errorf("Value is not correct")
			}
		})
	}

}

func TestSetData(t *testing.T) {

	p := New("01", "https://127.0.0.1")

	var tests = []struct {
		key string
		val string
	}{
		{"key1", "val1"},
		{"key2", "val2"},
		{"key3", ""},
	}

	for _, tst := range tests {
		t.Run(tst.key, func(t *testing.T) {
			p.SetData(tst.key, tst.val)
			val := p.GetData(tst.key)
			if tst.val != val {
				t.Errorf("Value is not correct: %s != %s", val, tst.val)
			}
		})
	}

}
