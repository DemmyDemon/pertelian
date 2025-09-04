package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/DemmyDemon/pertelian"
	"github.com/google/gousb"
)

func doOrDie(message string, err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err)
	os.Exit(1)
}

func shutdown(ctx *gousb.Context, pert pertelian.PertelianX2040) {
	pert.Clear()
	doOrDie("Turnig off display", pert.Off())
	doOrDie("Closing device", pert.Close())
	err := ctx.Close()
	doOrDie("Closing GoUSB context", err)
}

func main() {
	ctx := gousb.NewContext()
	pert, err := pertelian.NewX2040(ctx)
	doOrDie("Opening display comms", err)
	doOrDie("Turning on display", pert.On())

	defer shutdown(ctx, pert)

	notify := make(chan os.Signal, 1)
	signal.Notify(notify, os.Interrupt)
	ticker := time.NewTicker(250 * time.Millisecond)

	// TODO: Listen somewhere for what to put on line 2 and 3

	for {
		select {
		case sig := <-notify:
			fmt.Fprintf(os.Stderr, "Got signal %s\n", sig)
			pert.Centered(2, "Bye!")
			time.Sleep(500 * time.Millisecond)
			shutdown(ctx, pert)
			os.Exit(0)
		case <-ticker.C:
			now := time.Now()
			_, week := now.ISOWeek()
			pert.Centered(0, fmt.Sprintf("%s %d", now.Format("Monday"), week))
			pert.Centered(1, now.Format("2006-01-02 15:04"))
		}
	}
}
