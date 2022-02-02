package tea

import (
	"fmt"
	"testing"
	"time"
)

func TestEvery(t *testing.T) {
	expected := "every ms"
	msg := Every(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	if expected != msg {
		t.Fatalf("expected a msg %v but got %v", expected, msg)
	}
}

func TestTick(t *testing.T) {
	expected := "tick"
	msg := Tick(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	if expected != msg {
		t.Fatalf("expected a msg %v but got %v", expected, msg)
	}
}

func TestSequentially(t *testing.T) {
	var expectedErrMsg = fmt.Errorf("some err")
	var expectedStrMsg = "some msg"

	var nilReturnCmd = func() Msg {
		return nil
	}

	tests := []struct {
		name     string
		cmds     []Cmd
		expected Msg
	}{
		{
			name:     "all nil",
			cmds:     []Cmd{nilReturnCmd, nilReturnCmd},
			expected: nil,
		},
		{
			name:     "null cmds",
			cmds:     []Cmd{nil, nil},
			expected: nil,
		},
		{
			name: "one error",
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedErrMsg
				},
				nilReturnCmd,
			},
			expected: expectedErrMsg,
		},
		{
			name: "some msg",
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedStrMsg
				},
				nilReturnCmd,
			},
			expected: expectedStrMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if msg := Sequentially(test.cmds...)(); msg != test.expected {
				t.Fatalf("expected a msg %v but got %v", test.expected, msg)
			}
		})
	}
}
