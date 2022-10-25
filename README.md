# Pertelian

Talking to Pertelian character LCD displays in Go

Note that only the Pertelian X2040 is supported at the moment. I don't know if there are any other Pertelian displays in existance. This is the one I had laying around in a drawer, so this is the one I wanted to play with.

## Underdocumented

This is just something I threw together for my own use. It's only public becasue I'm *also* playing around with putting my Go code on GitHub.

Originally part of a different project, but I figured I might as well toss it out there for general use.

Enjoy to the extent possible.

## Example

```go
package main

import (
	"log"
	"time"
	"github.com/DemmyDemon/pertelian"
	"github.com/google/gousb"
)

func main() {

    // First we need to get a hold of a USB context to work with.
    // Explaining how to get gousb up and running with the Pertelian X2040 is out of scope here.
	ctx := gousb.NewContext()
    // Suffice to say, read the gousb documentation.
    // Windows? Strongly consider https://zadig.akeo.ie/

    // Releasing the USB context after use is very important, or so I've heard.
	defer ctx.Close()

    // Once we have a USB context, it all gets a lot simpler!
    // Instantiate a display object:
	pert, err := pertelian.NewX2040(ctx)

    // Errors can be stuff like "Device not found" (if you forgot to plug it in)
    // or "Unsupported" (if you have the wrong driver)
	if err != nil {
		log.Fatal(err)
	}

    // You should always close USB devices. If not, the USB gods will eat your cookies!
	defer pert.Close()
    // Note that this only closes the connection to the display, it does not turn it off.
    // It doesn't even blank it or turn off the light. It leaves it in whatever state it's in.

    // Turning the display on is a good idea if you want to put information on it.
	err = pert.On()
	if err != nil {
		log.Fatal(err)
	}

    // Splash screen. Very fancy.
	pert.Splash()

    // Sleeep for a while to show off the very fancy splash screen.
	time.Sleep(5 * time.Second)

    // Clear out the very fancy splash screen.
	pert.Clear()

    // Print some text roughly centered on the second line (first line is 0)
    pert.Centered(1, "Some Text")

    // Print more text on the third line, 0 characters offset from the left edge.
    pert.PrintAt(2, 0, "More Text")

    // Sleep again, of nobody will see the text.
	time.Sleep(5 * time.Second)

    // Shut down the display, because electricity is very expensive right now.
	err = pert.Off()
	if err != nil {
		log.Fatal(err)
	}
}

```
