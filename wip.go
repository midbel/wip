package wip

import (
  "errors"
  "fmt"
  "os"
)

var (
  ErrKind = errors.New("wip: unsupported indicator")
  ErrWidth = errors.New("wip: invalid width")
)

type Color uint8

type IndicatorKind uint8

const (
	None IndicatorKind = iota
	Percent
	Size
	Rate
	Time
	Bounce
	Scroll
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
		b.label = label
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
		case None, Percent, Size, Rate, Time:
			b.indicator = kind
		default:
			return ErrIndicator
		}
		return nil
	}
}

func WithBackground(c Color) Option {
	return func(b *Bar) error {
		b.back = c
		return nil
	}
}

func WithForeground(c Color) Option {
	return func(b *Bar) error {
		b.fore = c
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

const (
	lsquare = '['
	rsquare = ']'
	space   = ' '
	pound   = '#'
	rangle  = '>'
)

const DefaultWidth = 50

type Bar struct {
	pre   byte
	post  byte
	char  byte
	space byte
	arrow byte

	width   int64
	current int64
	total   int64

	label     string
	indicator IndicatorKind
	back      Color
	fore      Color
}

func Zero(size int64) (*Bar, error) {
	b := Bar{
		pre:       lsquare,
		post:      rsquare,
		char:      pound,
		space:     space,
		indicator: Percent,
		width:     DefaultWidth,
    total:     size,
	}
	return &b, nil
}

func New(size int64, options ...Option) (*Bar, error) {
	b, err := Zero(size)
	if err != nil {
		return nil, err
	}
	for _, o := range options {
		if err := o(b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func DefaultProgress(label string, width, limit int64) *Bar {
	options := []Option{
		WithSpace('-'),
		WithFill('#'),
		WithWidth(width),
		WithLabel(label),
	}
	b, _ := New(limit, options...)
	return b
}

func (b *Bar) Reset() {
	b.current = 0
}

func (b *Bar) Incr(n int64) {
	b.current += n
	b.print()
}

func (b *Bar) Write(b []byte) (int, error) {
  b.Incr(int64(len(b)))
  return len(b), nil
}

const row = "%c%-*s%c %3d%%"

func (b *Bar) print() {
	var (
		frac  = float64(b.current) / float64(b.total)
		count = float64(b.width) * frac
		fill  = strings.Repeat(string(b.char), int(count))
	)
	if count > 0 && int64(count) < b.width && b.arrow != 0 {
		fill += string(b.arrow)
	}
	fmt.Fprint(os.Stdout, "\r")
	if b.label != "" {
		pat := "%-16s " + row
		fmt.Fprintf(os.Stdout, pat, b.label, b.pre, b.width, fill, b.post, int(frac*100))
	} else {
		fmt.Fprintf(os.Stdout, row, b.pre, b.width, fill, b.post, int(frac*100))
	}
}
