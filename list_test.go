package timewheel

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	l := newList()
	l.push(1)
	n2 := l.push("2")
	if l.size != 2 {
		t.Fatalf("invalid list size, want %d, got %d", 2, l.size)
	}
	vs := make([]interface{}, 0, l.size)
	for cur := l.head; cur != nil; cur = cur.next {
		vs = append(vs, cur)
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

	l.remove(n2)
	if l.size != 1 {
		t.Fatalf("invalid size after remove")
	}

	fmt.Println(l)
}
