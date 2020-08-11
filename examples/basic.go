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
    i.Kind = wip.None
  case "size":
    i.Value = wip.Size
  default:
    return fmt.Errorf("%s: unknown value", str)
  }
  return nil
}

func (i *Indicator) String() string {
  return "indicator"
}

func main() {
  var k Indicator
  flag.Var(&k, "k", "indicator type")
	flag.Parse()

	for _, a := range flag.Args() {
		readFile(a, k.Kind)
    fmt.Println()
	}
}

func readFile(a string, kind wip.IndicatorKind) {
	r, err := os.Open(a)
	if err != nil {
		return
	}
	defer r.Close()

	i, err := r.Stat()
	if err != nil {
		return
	}

	bar := Create(i.Size(), a, kind)
	io.CopyBuffer(bar, r, make([]byte, 1024))
}

func Create(size int64, file string, kind wip.IndicatorKind) *wip.Bar {
	options := []wip.Option{
		wip.WithLabel(filepath.Base(file)),
		wip.WithWidth(32),
		wip.WithSpace('-'),
		wip.WithFill('#'),
		wip.WithDelimiter('[', ']'),
		wip.WithIndicator(kind),
		wip.WithForeground(wip.LightGreen),
	}
	bar, _ := wip.New(size, options...)
	return bar
}
