package history

import (
	"bytes"
	"github.com/Rnoadm/rpg"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func testSetup0(t *testing.T) (*History, func()) {
	f, err := ioutil.TempFile(os.TempDir(), "testhistory")
	if err != nil {
		t.Error(err)
		t.SkipNow()
	}
	return NewHistory(f), func() {
		n := f.Name()
		f.Close()
		os.Remove(n)
	}
}

func testSetup1(t *testing.T) (*History, []byte, func()) {
	h, cleanup := testSetup0(t)

	s := rpg.NewState()
	err := h.Append(s)
	if err != nil {
		t.Fatal(err)
	}
	g0, err := s.GobEncode()
	if err != nil {
		t.Error(err)
	}

	return h, g0, cleanup
}

func testSetup2(t *testing.T) (*History, []byte, []byte, func()) {
	h, cleanup := testSetup0(t)

	s := rpg.NewState()
	err := h.Append(s)
	if err != nil {
		t.Fatal(err)
	}
	g0, err := s.GobEncode()
	if err != nil {
		t.Error(err)
	}
	if !s.Atomic(func(s *rpg.State) bool {
		_, _ = s.Create()
		return true
	}) {
		t.Fatal("Atomic failed")
	}
	err = h.Append(s)
	if err != nil {
		t.Fatal(err)
	}
	g1, err := s.GobEncode()
	if err != nil {
		t.Error(err)
	}

	return h, g0, g1, cleanup
}

func expect(t *testing.T, state []byte, expected error, tell int64) func(*rpg.State, error) func(int64) {
	return func(s *rpg.State, err error) func(int64) {
		if err != expected {
			t.Error("not equal: ", err, " != ", expected)
		}
		if (s == nil) != (state == nil) {
			t.Error("not equal: ", s, " != ", state)
		} else if state != nil {
			b, err := s.GobEncode()
			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(b, state) {
				t.Error("not equal: ", b, " != ", state)
			}
		}

		return func(i int64) {
			if i != tell {
				t.Error("not equal: ", i, " != ", tell)
			}
		}
	}
}

func TestSeekPrevEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekCur))
}

func TestSeekCurEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(0, SeekCur))
}

func TestSeekFirstEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(0, SeekStart))
}

func TestSeekSecondEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(1, SeekStart))
}

func TestSeekSecondLastEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekEnd))
}

func TestSeekLastEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(0, SeekEnd))
}

func TestSeekReverseEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	h.Reset()

	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekCur))
}

func TestSeekForwardEmpty(t *testing.T) {
	h, cleanup := testSetup0(t)
	defer cleanup()

	h.Reset()

	expect(t, nil, io.EOF, -1)(h.Seek(1, SeekCur))
}

func TestSeekPrevOne(t *testing.T) {
	h, _, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekCur))
}

func TestSeekCurOne(t *testing.T) {
	h, b0, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(0, SeekCur))(h.Tell())
}

func TestSeekFirstOne(t *testing.T) {
	h, b0, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(0, SeekStart))(h.Tell())
}

func TestSeekSecondOne(t *testing.T) {
	h, _, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(1, SeekStart))
}

func TestSeekSecondLastOne(t *testing.T) {
	h, _, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekEnd))
}

func TestSeekLastOne(t *testing.T) {
	h, b0, cleanup := testSetup1(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(0, SeekEnd))(h.Tell())
}

func TestSeekReverseOne(t *testing.T) {
	h, b0, cleanup := testSetup1(t)
	defer cleanup()

	h.Reset()

	expect(t, b0, nil, 0)(h.Seek(-1, SeekCur))(h.Tell())
	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekCur))
}

func TestSeekForwardOne(t *testing.T) {
	h, b0, cleanup := testSetup1(t)
	defer cleanup()

	h.Reset()

	expect(t, b0, nil, 0)(h.Seek(1, SeekCur))(h.Tell())
	expect(t, nil, io.EOF, -1)(h.Seek(1, SeekCur))
}

func TestSeekPrevTwo(t *testing.T) {
	h, b0, _, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(-1, SeekCur))(h.Tell())
}

func TestSeekCurTwo(t *testing.T) {
	h, _, b1, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b1, nil, 1)(h.Seek(0, SeekCur))(h.Tell())
}

func TestSeekFirstTwo(t *testing.T) {
	h, b0, _, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(0, SeekStart))(h.Tell())
}

func TestSeekSecondTwo(t *testing.T) {
	h, _, b1, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b1, nil, 1)(h.Seek(1, SeekStart))(h.Tell())
}

func TestSeekSecondLastTwo(t *testing.T) {
	h, b0, _, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b0, nil, 0)(h.Seek(-1, SeekEnd))(h.Tell())
}

func TestSeekLastTwo(t *testing.T) {
	h, _, b1, cleanup := testSetup2(t)
	defer cleanup()

	expect(t, b1, nil, 1)(h.Seek(0, SeekEnd))(h.Tell())
}

func TestSeekReverseTwo(t *testing.T) {
	h, b0, b1, cleanup := testSetup2(t)
	defer cleanup()

	h.Reset()

	expect(t, b1, nil, 1)(h.Seek(-1, SeekCur))(h.Tell())
	expect(t, b0, nil, 0)(h.Seek(-1, SeekCur))(h.Tell())
	expect(t, nil, io.EOF, -1)(h.Seek(-1, SeekCur))
}

func TestSeekForwardTwo(t *testing.T) {
	h, b0, b1, cleanup := testSetup2(t)
	defer cleanup()

	h.Reset()

	expect(t, b0, nil, 0)(h.Seek(1, SeekCur))(h.Tell())
	expect(t, b1, nil, 1)(h.Seek(1, SeekCur))(h.Tell())
	expect(t, nil, io.EOF, -1)(h.Seek(1, SeekCur))
}
