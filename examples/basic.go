package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/wip"
)

type Mode struct {
	Value wip.Mode
}

func (m *Mode) Set(str string) error {
	switch strings.ToLower(str) {
	case "regular", "":
		m.Value = wip.Regular
	case "bounce", "bouncing":
		m.Value = wip.Bouncing
	case "scroll", "scrolling":
		m.Value = wip.Scrolling
	default:
		return fmt.Errorf("%s: unknown mode", str)
	}
	return nil
}

func (m *Mode) String() string {
	return "mode"
}

type Color struct {
	Value wip.Color
}

func (c *Color) Set(str string) error {
	switch strings.ToLower(str) {
	case "black":
		c.Value = wip.Black
	case "red":
		c.Value = wip.DarkRed
	case "green":
		c.Value = wip.DarkGreen
	case "yellow":
		c.Value = wip.DarkYellow
	case "blue":
		c.Value = wip.DarkBlue
	case "purple":
		c.Value = wip.DarkPurple
	case "cyan":
		c.Value = wip.DarkCyan
	case "grey":
		c.Value = wip.LightGrey
	case "white":
		c.Value = wip.White
	default:
		return fmt.Errorf("%s: unknown color", str)
	}
	return nil
}

func (c *Color) String() string {
	return "color"
}

type Indicator struct {
	Value wip.IndicatorKind
}

func (i *Indicator) Set(str string) error {
	switch strings.ToLower(str) {
	case "percent", "":
		i.Value = wip.Percent
	case "elapsed":
		i.Value = wip.Elapsed
	case "remained":
		i.Value = wip.Remained
	case "rate":
		i.Value = wip.Rate
	case "none":
		i.Value = wip.None
	case "size":
		i.Value = wip.Size
	default:
		return fmt.Errorf("%s: unknown indicator", str)
	}
	return nil
}

func (i *Indicator) String() string {
	return "indicator"
}

const DefaultBufferSize = 1024

func main() {
	var (
		kind   = Indicator{Value: wip.Percent}
		fore   = Color{Value: wip.White}
		back   = Color{Value: wip.Black}
		width  = int64(wip.DefaultWidth)
		mode   = Mode{Value: wip.Regular}
		size   = DefaultBufferSize
		buffer bool
	)
	flag.Var(&kind, "k", "indicator type")
	flag.Var(&fore, "f", "foreground color")
	flag.Var(&back, "b", "background color")
	flag.Var(&mode, "m", "mode")
	flag.Int64Var(&width, "w", width, "bar width")
	flag.IntVar(&size, "z", size, "buffer size")
	flag.BoolVar(&buffer, "r", buffer, "bufferize reading")
	flag.Parse()

	filepath.Walk(flag.Arg(0), func(file string, i os.FileInfo, err error) error {
		if err != nil || i.IsDir() {
			return err
		}
		options := MakeOptions(file, kind, back, fore, mode, width)
		err = readFile(file, size, buffer, options)
		if err != nil {
			fmt.Println()
		}
		return err
	})
}

func readFile(file string, size int, buffer bool, options []wip.Option) error {
	if size <= 0 {
		size = DefaultBufferSize
	}
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	i, err := r.Stat()
	if err != nil {
		return err
	}

	bar, err := wip.New(i.Size(), options...)
	if err != nil {
		return nil
	}
	var rs io.Reader = r
	if buffer {
		rs = bufio.NewReader(rs)
	}
	if _, err := io.CopyBuffer(bar, rs, make([]byte, size)); err != nil {
		return err
	}
	bar.Complete()
	return nil
}

func MakeOptions(file string, kind Indicator, back, fore Color, mode Mode, width int64) []wip.Option {
	return []wip.Option{
		wip.WithLabel(filepath.Base(file)),
		wip.WithWidth(width),
		wip.WithSpace('-'),
		wip.WithFill('#'),
		wip.WithDelimiter('[', ']'),
		wip.WithIndicator(kind.Value),
		wip.WithForeground(fore.Value),
		wip.WithBackground(back.Value),
		wip.WithMode(mode.Value),
	}
}
