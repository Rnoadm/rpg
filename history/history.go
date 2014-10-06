package history

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/BenLubar/bindiff"
	"github.com/Rnoadm/rpg"
	"io"
)

const (
	SeekStart = iota
	SeekCur
	SeekEnd
)

const sizeof_int64 = 64 / 8

type History struct {
	i int64
	b []byte
	f io.ReadWriteSeeker
}

func NewHistory(f io.ReadWriteSeeker) *History {
	return &History{f: f, i: -1}
}

func (h *History) Seek(offset int64, whence int) (*rpg.State, error) {
	if whence == SeekCur && h.i < 0 {
		if offset > 0 {
			offset--
		} else {
			offset++
			whence = SeekEnd
		}
	}
	if whence == SeekStart || h.i < 0 {
		if _, err := h.f.Seek(0, SeekStart); err != nil {
			return nil, err
		}
		h.i, h.b = -1, nil
		if err := h.seekForward(); err != nil {
			return nil, err
		}
	}
	if whence == SeekEnd {
		for {
			err := h.seekForward()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
		}
	}
	for offset > 0 {
		if err := h.seekForward(); err != nil {
			return nil, err
		}
		offset--
	}
	for offset < 0 {
		if err := h.seekReverse(); err != nil {
			return nil, err
		}
		offset++
	}

	var s *rpg.State
	err := gob.NewDecoder(bytes.NewReader(h.b)).Decode(&s)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return s, err
}

func (h *History) seekForward() error {
	var size int64
	err := binary.Read(h.f, binary.LittleEndian, &size)
	if err != nil {
		return err
	}

	patch := make([]byte, size)
	_, err = io.ReadFull(h.f, patch)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	_, err = h.f.Seek(sizeof_int64, SeekCur)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	h.i++

	h.b, err = bindiff.Forward(h.b, patch)

	return err
}

func (h *History) seekReverse() error {
	if h.i == 0 {
		return io.EOF
	}

	_, err := h.f.Seek(-sizeof_int64, SeekCur)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	var size int64
	err = binary.Read(h.f, binary.LittleEndian, &size)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	_, err = h.f.Seek(-size-sizeof_int64*2, SeekCur)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	patch := make([]byte, sizeof_int64+size)
	_, err = io.ReadFull(h.f, patch)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	h.i--

	h.b, err = bindiff.Reverse(h.b, patch[sizeof_int64:])

	return err
}

func (h *History) Append(s *rpg.State) error {
	_, err := h.Seek(0, SeekEnd)
	if err == io.EOF && h.i == -1 {
		err = nil
	}
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(s)
	if err != nil {
		return err
	}

	patch := bindiff.Diff(h.b, buf.Bytes(), 10)

	h.i++
	h.b = buf.Bytes()

	err = binary.Write(h.f, binary.LittleEndian, int64(len(patch)))
	if err != nil {
		return err
	}

	n, err := h.f.Write(patch)
	if err != nil {
		return err
	}
	if n != len(patch) {
		return io.ErrShortWrite
	}

	err = binary.Write(h.f, binary.LittleEndian, int64(len(patch)))
	if err != nil {
		return err
	}

	return nil
}

func (h *History) Tell() int64 {
	return h.i
}

func (h *History) Reset() {
	h.i, h.b = -1, nil
}
