package bufio

import (
	"bytes"
	"errors"
	"io"
)

const (
	defaultBufSize           = 4096 // 默认4Kb
	minReadBufferSize        = 16   // 最小缓冲区
	maxConsecutiveEmptyReads = 100  // 最大连续空读
)

var (
	errNegativeRead  = errors.New("bufio: reader returned negative count from Read")
	ErrNegativeCount = errors.New("bufio: negative count")
	ErrBufferFull    = errors.New("bufio: buffer full")
)

// 自定义reader bufio
type Reader struct {
	buf  []byte
	rd   io.Reader
	r, w int
	err  error
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

func NewReaderSize(rd io.Reader, size int) *Reader {
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

func (b *Reader) Buffered() int {
	return b.w - b.r
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{buf: buf, rd: r}
}

// 从缓冲区读取元素
func (b *Reader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, b.readErr()
	}
	// 当缓冲区无有效长度
	if b.r == b.w {
		// 缓冲区数据没有,有可能err不为空
		if b.err != nil {
			return 0, b.err
		}
		// 如果传入的字节切片大于缓冲区的长度, 且缓冲区没有有效数据
		if len(p) >= len(b.buf) {
			// 直接从底层文件读取数据, 写入到p中
			n, b.err = b.rd.Read(p)
			if n < 0 {
				panic(errNegativeRead)
			}
			return n, b.err
		}
		// 到这里说明缓冲区为空, 并且 len(p) < len(b.buf)
		b.fill()
		if b.r == b.w {
			return 0, b.readErr()
		}
	}
	n += copy(p, b.buf[b.r:b.w])
	b.r += n
	return n, nil
}

// 读取一个字节
func (b *Reader) ReadByte() (c byte, err error) {
	for b.r == b.w {
		if b.err != nil {
			return 0, b.readErr()
		}
		b.fill()
	}
	c = b.buf[b.r]
	b.r++
	return c, nil
}

func (b *Reader) Pop(n int) ([]byte, error) {
	d, err := b.Peek(n)
	if err == nil {
		b.r += n
		return d, err
	}
	return nil, err
}

// 丢弃缓冲区的n个字节
func (b *Reader) Discard(n int) (discarded int, err error) {
	if n < 0 {
		return 0, ErrNegativeCount
	}
	if n == 0 {
		return
	}
	remain := n
	for {
		skip := b.Buffered()
		if skip == 0 {
			b.fill()
			skip = b.Buffered()
		}
		if skip > remain {
			skip = remain
		}
		b.r += skip
		remain -= skip
		if remain == 0 {
			return n, nil
		}
		if b.err != nil {
			return n - remain, b.readErr()
		}
	}
}

func (b *Reader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrNegativeCount
	}
	if n > len(b.buf) {
		return nil, ErrBufferFull
	}
	for b.w-b.r < n && b.err == nil {
		b.fill()
	}
	var err error
	if avail := b.w - b.r; avail < n {
		n = avail
		if err = b.readErr(); err == nil {
			err = ErrBufferFull
		}
	}
	return b.buf[b.r : b.r+n], err
}

// 填充缓冲区
func (b *Reader) fill() {
	// 说明缓冲区有被读过, 无效数据丢弃
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}
	if b.w >= len(b.buf) {
		panic("bufio: tried to fill full buffer")
	}
	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := b.rd.Read(b.buf[b.w:])
		if n < 0 {
			panic(errNegativeRead)
		}
		b.w += n
		if err != nil {
			b.err = err
			return
		}
		if n > 0 {
			return
		}
	}
	b.err = io.ErrNoProgress
}

func (b *Reader) ReadSlice(delim byte) (line []byte, err error) {
	for {
		// Search buffer.
		if i := bytes.IndexByte(b.buf[b.r:b.w], delim); i >= 0 {
			line = b.buf[b.r : b.r+i+1]
			b.r += i + 1
			break
		}

		// Pending error?
		if b.err != nil {
			line = b.buf[b.r:b.w]
			b.r = b.w
			err = b.readErr()
			break
		}

		// Buffer full?
		if b.Buffered() >= len(b.buf) {
			b.r = b.w
			line = b.buf
			err = ErrBufferFull
			break
		}

		b.fill() // buffer is not full
	}
	return
}

func (b *Reader) ReadLine() (line []byte, isPrefix bool, err error) {
	line, err = b.ReadSlice('\n')
	if err == ErrBufferFull {
		// Handle the case where "\r\n" straddles the buffer.
		if len(line) > 0 && line[len(line)-1] == '\r' {
			// Put the '\r' back on buf and drop it from line.
			// Let the next call to ReadLine check for "\r\n".
			if b.r == 0 {
				// should be unreachable
				panic("bufio: tried to rewind past start of buffer")
			}
			b.r--
			line = line[:len(line)-1]
		}
		return line, true, nil
	}

	if len(line) == 0 {
		if err != nil {
			line = nil
		}
		return
	}
	err = nil

	if line[len(line)-1] == '\n' {
		drop := 1
		if len(line) > 1 && line[len(line)-2] == '\r' {
			drop = 2
		}
		line = line[:len(line)-drop]
	}
	return
}

type Writer struct {
	err error
	buf []byte
	n   int
	wr  io.Writer
}

func NewWriterSize(w io.Writer, size int) *Writer {
	b, ok := w.(*Writer)
	if ok && len(b.buf) >= size {
		return b
	}
	if size <= 0 {
		size = defaultBufSize
	}
	return &Writer{buf: make([]byte, size), wr: w}
}

func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufSize)
}

func (b *Writer) Reset(w io.Writer) {
	b.err = nil
	b.n = 0
	b.wr = w
}

func (b *Writer) ResetBuffer(w io.Writer, buf []byte) {
	b.buf = buf
	b.err = nil
	b.n = 0
	b.wr = w
}

func (b *Writer) Flush() error {
	err := b.flush()
	return err
}

func (b *Writer) flush() error {
	if b.err != nil {
		return b.err
	}
	if b.n == 0 {
		return nil
	}
	n, err := b.wr.Write(b.buf[0:b.n])
	if n < b.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < b.n {
			copy(b.buf[0:b.n-n], b.buf[n:b.n])
		}
		b.n -= n
		b.err = err
		return err
	}
	b.n = 0
	return nil
}

func (b *Writer) Available() int {
	return len(b.buf) - b.n
}

func (b *Writer) Buffered() int {
	return b.n
}

func (b *Writer) Write(p []byte) (nn int, err error) {
	for len(p) > b.Available() && b.err == nil {
		var n int
		if b.Buffered() == 0 {
			n, b.err = b.wr.Write(p)
		} else {
			n = copy(b.buf[b.n:], p)
			b.n += n
			b.flush()
		}
		nn += n
		p = p[n:]
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], p)
	b.n += n
	nn += n
	return nn, nil
}

func (b *Writer) WriteRaw(p []byte) (nn int, err error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.Buffered() == 0 {
		nn, err = b.wr.Write(p)
		b.err = err
	} else {
		nn, err = b.Write(p)
	}
	return
}

func (b *Writer) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrNegativeCount
	}
	if n > len(b.buf) {
		return nil, ErrBufferFull
	}
	for b.Available() < n && b.err == nil {
		b.flush()
	}
	if b.err != nil {
		return nil, b.err
	}
	d := b.buf[b.n : b.n+n]
	b.n += n
	return d, nil
}

func (b *Writer) WriteString(s string) (int, error) {
	nn := 0
	for len(s) > b.Available() && b.err == nil {
		n := copy(b.buf[b.n:], s)
		b.n += n
		nn += n
		s = s[n:]
		b.flush()
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], s)
	b.n += n
	nn += n
	return nn, nil
}
