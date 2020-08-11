package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/wip"
)

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

func main() {
	var (
		kind   = Indicator{Value: wip.Percent}
		fore   = Color{Value: wip.White}
		back   = Color{Value: wip.DarkGreen}
		width  = int64(wip.DefaultWidth)
		scroll = false
	)
	flag.Var(&kind, "k", "indicator type")
	flag.Var(&fore, "f", "foreground color")
	flag.Var(&back, "b", "background color")
	flag.Int64Var(&width, "w", width, "bar width")
	flag.BoolVar(&scroll, "i", scroll, "indeterminate")
	flag.Parse()

	for _, a := range flag.Args() {
		options := MakeOptions(a, kind, back, fore, width)
		readFile(a, scroll, options)
		fmt.Println()
	}
}

func readFile(a string, scroll bool, options []wip.Option) {
	r, err := os.Open(a)
	if err != nil {
		return
	}
	defer r.Close()

	i, err := r.Stat()
	if err != nil {
		return
	}

	var bar *wip.Bar
	if scroll {
		bar = Indeterminate(options)
	} else {
		bar = Create(i.Size(), options)
	}
	io.CopyBuffer(bar, r, make([]byte, 1024))
	bar.Complete()
}

func Create(size int64, options []wip.Option) *wip.Bar {
	bar, _ := wip.New(size, options...)
	return bar
}

func Indeterminate(options []wip.Option) *wip.Bar {
	bar, _ := wip.Scroll(options...)
	return bar
}

func MakeOptions(file string, kind Indicator, back, fore Color, width int64) []wip.Option {
	return []wip.Option{
		wip.WithLabel(filepath.Base(file)),
		wip.WithWidth(width),
		wip.WithSpace('-'),
		wip.WithFill('#'),
		wip.WithDelimiter('[', ']'),
		wip.WithIndicator(kind.Value),
		wip.WithForeground(fore.Value),
		wip.WithBackground(back.Value),
	}
}
