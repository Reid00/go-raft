package raft

import (
	"regexp"
	"testing"
	"time"
)

func TestRandomTiemout(t *testing.T) {
	start := time.Now()
	timeout := randomTimeout(time.Millisecond)

	select {
	case <-timeout:
		// diff := time.Now().Sub(start)
		// FIXED above is not recommanded
		diff := time.Since(start)
		if diff < time.Millisecond {
			t.Fatalf("fired early")
		}
	case <-time.After(3 * time.Millisecond):
		t.Fatalf("timeout")
	}
}

func TestNewSeed(t *testing.T) {
	vals := make(map[int64]bool)

	for i := 0; i < 10000; i++ {
		seed := newSeed()
		if _, exists := vals[seed]; exists {
			t.Fatalf("newSeed() return a value it'd previously returned")
		}
		vals[seed] = true
	}
}

func TestRandomTimeout_NoTime(t *testing.T) {
	timeout := randomTimeout(0)
	if timeout != nil {
		t.Fatalf("expected nil channel")
	}
}

func TestMin(t *testing.T) {
	if min(1, 1) != 1 {
		t.Fatalf("bad min")
	}
	if min(2, 1) != 1 {
		t.Fatal("bad min")
	}
	if min(1, 2) != 1 {
		t.Fatalf("bad min")
	}
}

func TestMax(t *testing.T) {
	if max(1, 1) != 1 {
		t.Fatalf("bad max")
	}
	if max(2, 1) != 2 {
		t.Fatal("bad max")
	}
	if max(1, 2) != 2 {
		t.Fatalf("bad max")
	}
}

func TestGenerateUUID(t *testing.T) {
	prev := generateUUID()
	pattern, err := regexp.Compile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`)
	if err != nil {
		panic("regexp Compile err: " + err.Error())
	}
	for i := 0; i < 100; i++ {
		id := generateUUID()
		if prev == id {
			t.Fatalf("should get a new ID!")
		}
		matched := pattern.MatchString(id)
		if !matched {
			t.Fatalf("expected match %s, %v, %v", id, matched, err)
		}
		// FIXED below has poor performace
		// matched, err := regexp.MatchString(
		// 	`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`, id)
		// if !matched || err != nil {
		// 	t.Fatalf("expected match %s, %v, %v", id, matched, err)
		// }
	}
}
func TestBackoff(t *testing.T) {
	b := backoff(10*time.Millisecond, 1, 8)
	if b != 10*time.Millisecond {
		t.Fatalf("bad: %v", b)
	}

	b = backoff(20*time.Millisecond, 2, 8)
	if b != 20*time.Millisecond {
		t.Fatalf("bad: %v", b)
	}

	b = backoff(10*time.Millisecond, 8, 8)
	if b != 640*time.Millisecond {
		t.Fatalf("bad: %v", b)
	}

	b = backoff(10*time.Millisecond, 9, 8)
	if b != 640*time.Millisecond {
		t.Fatalf("bad: %v", b)
	}
}

func TestOverrideNotifyBool(t *testing.T) {
	ch := make(chan bool, 1)

	// sanity check - buffered channel don't have any values
	select {
	case v := <-ch:
		t.Fatalf("unexpected receive: %v", v)
	default:
	}

	// simple case of single push
	overrideNotifyBool(ch, false)
	select {
	case v := <-ch:
		if v != false {
			t.Fatalf("expected false but got %v", v)
		}
	default:
		t.Fatalf("expected a value but is not ready")
	}

	// assert that function never blocks and only last item is received
	overrideNotifyBool(ch, false)
	overrideNotifyBool(ch, false)
	overrideNotifyBool(ch, false)
	overrideNotifyBool(ch, false)
	overrideNotifyBool(ch, true)

	select {
	case v := <-ch:
		if v != true {
			t.Fatalf("expected true but got %v", v)
		}
	default:
		t.Fatalf("expected a value but is not ready")
	}

	// no further value is available
	select {
	case v := <-ch:
		t.Fatalf("unexpected receive: %v", v)
	default:
	}

}
