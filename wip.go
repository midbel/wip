package wip

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	ErrKind  = errors.New("wip: unknown indicator")
	ErrWidth = errors.New("wip: invalid width")
	ErrColor = errors.New("wup: unknown color")
)

const (
	lsquare  = '['
	rsquare  = ']'
	space    = ' '
	pound    = '#'
	percent  = '%'
	carriage = '\r'
)

const (
	DefaultWidth      = 50
	defaultPrologSize = 24
	defaultEpilogSize = 16
)

type IndicatorKind uint8

const (
	None IndicatorKind = iota
	Percent
	Size
	Rate
	Elapsed
	Remained
	// Bounce
	// Scroll
)

type Option func(*Bar) error

func WithFill(c byte) Option {
	return func(b *Bar) error {
		b.char = c
		return nil
	}
}

func WithSpace(c byte) Option {
	return func(b *Bar) error {
		b.space = c
		return nil
	}
}

func WithLabel(label string) Option {
	return func(b *Bar) error {
		b.prolog = makeSlice(defaultPrologSize, space)
		copy(b.prolog, label)
		return nil
	}
}

func WithDelimiter(pre, post byte) Option {
	return func(b *Bar) error {
		b.pre, b.post = pre, post
		return nil
	}
}

func WithArrow(c byte) Option {
	return func(b *Bar) error {
		b.arrow = c
		return nil
	}
}

func WithIndicator(kind IndicatorKind) Option {
	return func(b *Bar) error {
		switch kind {
		case None, Percent, Size, Rate, Elapsed, Remained:
			b.indicator = kind
			if kind != None {
				b.epilog = makeSlice(defaultEpilogSize, space)
			}
		default:
			return ErrKind
		}
		return nil
	}
}

func WithWidth(width int64) Option {
	return func(b *Bar) error {
		if width <= 0 {
			return ErrWidth
		}
		b.width = width
		return nil
	}
}

type Color uint8

const (
	foreground = 38
	background = 48
)

func (c Color) background() string {
	return c.sequence(background)
}

func (c Color) foreground() string {
	return c.sequence(foreground)
}

func (c Color) sequence(n int) string {
	return fmt.Sprintf("\033[%d;5;%dm", n, c)
}

const (
	Black Color = iota
	DarkRed
	DarkGreen
	DarkYellow
	DarkBlue
	DarkPurple
	DarkCyan
	LightGrey
	DarkGrey
	LightRed
	LightGreen
	LightYellow
	LigthBlue
	LightPurple
	LightCyan
	White
)

func WithBackground(c Color) Option {
	return func(b *Bar) error {
		if c > White {
			return ErrColor
		}
		b.back.color = c
		b.back.set = true
		return nil
	}
}

func WithForeground(c Color) Option {
	return func(b *Bar) error {
		if c > White {
			return ErrColor
		}
		b.fore.color = c
		b.fore.set = true
		return nil
	}
}

type Bar struct {
	pre   byte
	post  byte
	char  byte
	space byte
	arrow byte

	prolog []byte
	epilog []byte

	width int64
	ui    *widget

	indicator IndicatorKind
	back      struct {
		color Color
		set   bool
	}
	fore struct {
		color Color
		set   bool
	}

	tcn state
}

func New(size int64, options ...Option) (*Bar, error) {
	var b Bar
	b.init()
	for _, o := range options {
		if err := o(&b); err != nil {
			return nil, err
		}
	}
	b.Reset(size)
	b.ui = makeWidget(int(b.width), b.space)
	return &b, nil
}

func (b *Bar) init() {
	b.pre = lsquare
	b.post = rsquare
	b.char = pound
	b.space = space
	b.indicator = Percent
	b.width = DefaultWidth
}

func (b *Bar) Reset(size int64) {
	b.tcn.Reset(size)
	b.ui.reset(b.space)
}

func (b *Bar) Complete() {
	b.tcn.Complete()
	b.print()
}

func (b *Bar) Incr(n int64) {
	b.tcn.Incr(n)
	b.print()
}

func (b *Bar) Update(n int64) {
	b.tcn.Set(n)
	b.print()
}

func (b *Bar) Write(bs []byte) (int, error) {
	b.Incr(int64(len(bs)))
	return len(bs), nil
}

func (b *Bar) print() {
	var (
		frac  = b.tcn.Fraction()
		count = float64(b.width) * frac
	)

	var line bytes.Buffer
	if count > 0 {
		line.WriteByte(carriage)
	}
	b.writeProlog(&line)
	b.writeProgress(&line, int(count))
	b.writeEpilog(&line)
	os.Stdout.Write(line.Bytes())
}

func (b *Bar) writeProlog(line *bytes.Buffer) {
	if len(b.prolog) == 0 {
		return
	}
	line.Write(b.prolog)
}

func (b *Bar) writeProgress(line *bytes.Buffer, count int) {
	if b.back.set {
		line.WriteString(b.back.color.background())
	}
	if b.fore.set {
		line.WriteString(b.fore.color.foreground())
	}
	if b.pre != 0 {
		line.WriteByte(b.pre)
	}
	line.Write(b.ui.update(b.char, b.arrow, count))
	if b.post != 0 {
		line.WriteByte(b.post)
	}
	if b.back.set || b.fore.set {
		line.WriteString("\x1b[0m")
	}
}

func (b *Bar) writeEpilog(line *bytes.Buffer) {
	if len(b.epilog) == 0 {
		return
	}
	var tmp []byte
	switch b.indicator {
	case None:
	case Percent:
		tmp = strconv.AppendFloat(tmp, b.tcn.Fraction()*100, 'f', 2, 64)
		tmp = append(tmp, percent)
	case Size:
		tmp = strconv.AppendInt(tmp, b.tcn.Current(), 10)
	case Rate:
		tmp = strconv.AppendFloat(tmp, b.tcn.Rate(), 'f', 2, 64)
	case Elapsed:
		e := b.tcn.Elapsed()
		tmp = []byte(e.String())
	case Remained:
		e := b.tcn.Remained()
		tmp = []byte(e.String())
	}
	defer fillSlice(b.epilog, space)

	copy(b.epilog[len(b.epilog)-len(tmp):], tmp)
	line.Write(b.epilog)
}

func makeSlice(size int, fill byte) []byte {
	b := make([]byte, size)
	fillSlice(b, fill)
	return b
}

func fillSlice(b []byte, fill byte) {
	for i := range b {
		b[i] = fill
	}
}

type widget struct {
	buffer []byte
	offset int
}

func makeWidget(width int, fill byte) *widget {
	w := widget{
		offset: 0,
		buffer: makeSlice(width, fill),
	}
	return &w
}

func (w *widget) reset(fill byte) {
	if w == nil {
		return
	}
	fillSlice(w.buffer, fill)
	w.offset = 0
}

func (w *widget) update(fill, arrow byte, count int) []byte {
	for i := w.offset; i < count; i++ {
		w.buffer[i] = fill
	}
	w.offset = count

	if count > 0 && count < len(w.buffer) && arrow != 0 {
		w.buffer[w.offset] = arrow
	}
	return w.buffer
}

func (w *widget) bytes() []byte {
	return w.buffer
}

type state struct {
	current int64
	total   int64
	now     time.Time
}

func (s *state) Indeterminate() bool {
	return s.total <= 0
}

func (s *state) Reset(total int64) {
	s.total = total
	s.current = 0
	s.now = time.Now()
}

func (s *state) Complete() {
	if s == nil {
		return
	}
	s.current = s.total
}

func (s *state) Set(n int64) {
	if s == nil {
		return
	}
	s.current = n
}

func (s *state) Incr(n int64) {
	if s == nil {
		return
	}
	s.current += n
}

func (s *state) Current() int64 {
	return s.current
}

func (s *state) Elapsed() time.Duration {
	return time.Since(s.now)
}

func (s *state) Estimated() time.Duration {
	r := float64(s.total) / s.Rate()
	return time.Duration(r) * time.Second
}

func (s *state) Remained() time.Duration {
	r := s.Estimated() - s.Elapsed()
	if r < 0 {
		r = 0
	}
	return r
}

func (s *state) Rate() float64 {
	e := s.Elapsed()
	if e.Seconds() == 0 {
		return float64(s.total)
	}
	return float64(s.current) / e.Seconds()
}

func (s *state) Fraction() float64 {
	return float64(s.current) / float64(s.total)
}
