package timewheel

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	l := newList()
	l.Push(1)
	l.Push("2")
	if l.Size() != 2 {
		t.Fatalf("invalid list size, want %d, got %d", 2, l.Size())
	}
	vs := make([]interface{}, 0, l.Size())
	for cur := l.Head(); cur != nil; cur = cur.Next() {
		vs = append(vs, cur.Value())
	}
	if len(vs) != 2 {
		t.Fatalf("invalid values size, want %d, got %d", 2, len(vs))
	}

	if v, ok := vs[0].(int); !ok || v != 1 {
		t.Fatalf("invalid value: %v", vs[0])
	}

	if v, ok := vs[1].(string); !ok || v != "2" {
		t.Fatalf("invalid value: %v", vs[1])
	}

	fmt.Println(l)
}
