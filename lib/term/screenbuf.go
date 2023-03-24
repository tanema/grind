//go:build !windows
// +build !windows

package term

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/term"
)

const (
	defaultTermWidth = 80
)

// Println will print a formatted string out to a writer
func Println(in string, data interface{}, fs ...string) error {
	return NewScreenBuf(os.Stderr, fs...).Render(in, data)
}

// PrintlnTmpl will print a formatted string out to a writer
func PrintlnTmpl(tmpl string, data interface{}, fs ...string) error {
	return NewScreenBuf(os.Stderr, fs...).RenderTmpl(tmpl, data)
}

// ScreenBuf is a convenient way to write to terminal screens. It creates,
// clears and, moves up or down lines as needed to write the output to the
// terminal using ANSI escape codes.
type ScreenBuf struct {
	w    io.Writer
	buf  *bytes.Buffer
	mut  sync.Mutex
	tmpl *template.Template
}

// NewScreenBuf creates and initializes a new ScreenBuf.
func NewScreenBuf(w io.Writer, sources ...string) *ScreenBuf {
	tmpl := template.New("screenbuf").Funcs(funcMap)
	for _, src := range sources {
		template.Must(tmpl.Parse(src))
	}
	return &ScreenBuf{buf: &bytes.Buffer{}, w: w, tmpl: tmpl}
}

// Render will write a text/template out to the console, using a mutex so that
// only a single writer at a time can write. This prevents the buffer from losing
// sync with the newlines
func (s *ScreenBuf) Render(in string, data interface{}) error {
	if err := s.Reset(); err != nil {
		return err
	}
	if err := s.Write(in, data); err != nil {
		return err
	}
	return s.Flush()
}

// RenderTmpl will write an already parsed text/template out to the console, using
// a mutex so that only a single writer at a time can write. This prevents the
// buffer from losing sync with the newlines
func (s *ScreenBuf) RenderTmpl(tmpl string, data interface{}) error {
	if err := s.Reset(); err != nil {
		return err
	}
	if err := s.WriteTmpl(tmpl, data); err != nil {
		return err
	}
	return s.Flush()
}

// Reset will empty the buffer and refill it with control characters that will
// clear the previous data on the next flush call.
func (s *ScreenBuf) Reset() error {
	s.mut.Lock()
	defer s.mut.Unlock()
	linecount := bytes.Count(s.buf.Bytes(), []byte("\n"))
	s.buf.Reset()
	_, err := s.buf.Write([]byte(strings.Repeat(clearLastLine, linecount)))
	return err
}

// WriteTmpl will write an already parsed template to the buffer, this will not
// render to the screen without calling Flush. It will also not reset the screen,
// this is append only. Call reset first.
func (s *ScreenBuf) WriteTmpl(tmpl string, data interface{}) error {
	var buf bytes.Buffer
	fmt.Println(tmpl, s.tmpl.Lookup(tmpl))
	if err := s.tmpl.Lookup(tmpl).Execute(&buf, data); err != nil {
		return err
	}
	return s.write(buf.String(), data)
}

// Write will write to the buffer, this will not render to the screen without calling
// Flush. It will also not reset the screen, this is append only. Call reset first.
func (s *ScreenBuf) Write(in string, data interface{}) error {
	var buf bytes.Buffer
	if err := template.Must(s.tmpl.Parse(in)).Execute(&buf, data); err != nil {
		return err
	}
	return s.write(buf.String(), data)
}

func (s *ScreenBuf) write(in string, data interface{}) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width = defaultTermWidth
	}
	tmpl := wrapANSI(in, width)
	if !strings.HasSuffix(tmpl, "\n") {
		tmpl += "\n"
	}
	_, err = s.buf.WriteString(tmpl)
	return err
}

// Flush will flush the render buffer to the screen, this should be called after
// sever calls to Write
func (s *ScreenBuf) Flush() error {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, err := io.Copy(s.w, bytes.NewBuffer(s.buf.Bytes()))
	return err
}
