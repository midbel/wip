package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/midbel/wip"
)

func main() {
	var (
		wait   = time.Millisecond * 100
		total  = time.Minute
		bounce = false
	)
	flag.DurationVar(&wait, "w", wait, "wait time")
	flag.DurationVar(&total, "t", total, "total time")
	flag.BoolVar(&bounce, "b", false, "bounce")
	flag.Parse()

	var (
		now  = time.Now()
		tick = time.NewTicker(wait)
		bar  *wip.Bar
	)
	if options := MakeOptions("working..."); bounce {
		bar, _ = wip.Bounce(options...)
	} else {
		bar, _ = wip.Scroll(options...)
	}
	defer tick.Stop()
	rand.Seed(now.Unix())
	for range tick.C {
		n := rand.Intn(10)
		bar.Incr(int64(n))
		if time.Since(now) >= total {
			bar.Complete()
			break
		}
	}
}

func MakeOptions(label string) []wip.Option {
	return []wip.Option{
		wip.WithLabel(label),
		wip.WithWidth(16),
		wip.WithSpace('-'),
		wip.WithFill('#'),
		wip.WithDelimiter('[', ']'),
		wip.WithIndicator(wip.Elapsed),
		wip.WithForeground(wip.DarkGreen),
		wip.WithBackground(wip.Black),
	}
}
