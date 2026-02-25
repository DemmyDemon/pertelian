package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/DemmyDemon/pertelian"
	"github.com/google/gousb"
)

const (
	// Where will the display daemon listen for lines?
	listenOn = "localhost:1984"
)

// doOrDie checks the given error, and if it's non-nil, it prints the message, and the error, to STDERR, before exiting with an error code.
func doOrDie(message string, err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err)
	os.Exit(1)
}

// shutdown clears the display, turns  it off, and closes the device handle/context.
func shutdown(ctx *gousb.Context, pert pertelian.PertelianX2040) {
	pert.Clear()
	doOrDie("Turnig off display", pert.Off())
	doOrDie("Closing device", pert.Close())
	err := ctx.Close()
	doOrDie("Closing GoUSB context", err)
}

// listenPlx opens a tcp port for listening to get lines. Failing to do so (port in use, etc) is fatal.
func listenPlx(chLines chan string) {
	listener, err := net.Listen("tcp", listenOn)
	doOrDie("Listening for lines", err)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection error:%s\n", err)
			continue
		}
		go handleClient(conn, chLines)
	}
}

// handleClient does the actual reading of lines from the connecting client, handing the string over to the main loop using the provided channel.
func handleClient(conn net.Conn, chLines chan string) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error:%s\n", err)
		return
	}
	chLines <- string(buffer[:n])
}

// sendLine dials the daemon, which is presumed to be already running, and sends the given string. It closes the connection after doing so. All errors (connection refused, etc) are fatal.
func sendLine(line string) {
	conn, err := net.Dial("tcp", listenOn)
	doOrDie("Connecting to display daemon", err)
	defer conn.Close()
	_, err = conn.Write([]byte(line))
	doOrDie("Sending to display daemon", err)
}

func main() {

	// If there are arguments at all, we presume they are meant for the display.
	// They are simply joined up and sent. All handling of the line being too long,
	// or whatever, is done on the daemon end, to keep the client simple.
	if len(os.Args) > 1 {
		sendLine(strings.Join(os.Args[1:], " "))
		os.Exit(0) // When we send stuff, we don't do anything else.
	}

	// First, we srt up the connection to the device.
	ctx := gousb.NewContext()
	pert, err := pertelian.NewX2040AutoDetach(ctx) // AutoDetach is useful for most USB devices where the kernel grabs them.

	doOrDie("Opening display comms", err)
	doOrDie("Turning on display", pert.On())

	defer shutdown(ctx, pert) // Shutting down properly is preferred, but not really critical.

	// Set up the channel to handle the interrupt signal
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, os.Interrupt)

	// Set up the channel to update the display several times a second
	ticker := time.NewTicker(250 * time.Millisecond)

	// Set up the two additional lines, and prepare to recieve new ones.
	chLines := make(chan string)
	lines := []string{"Display ready.", "-- !! --"}
	go listenPlx(chLines)

	for {
		select {
		case sig := <-notify:

			// This is just run-of-the-mill signal handling.

			fmt.Fprintf(os.Stderr, "Got signal %s\n", sig)
			pert.Clear()
			pert.Centered(1, "Shutting down display.")
			pert.Centered(2, "Bye!")
			time.Sleep(500 * time.Millisecond)
			shutdown(ctx, pert)
			os.Exit(0)
		case line := <-chLines:
			if len(line) > 20 {
				line = line[:20]
			}
			lines[1] = lines[0]
			lines[0] = line

			// If we don't clear the display here, there will likely be
			// partial lines left. If the new lines[0] is shorter than the old
			// one, the old one will be partially visible. Clearing the display
			// with every write, however, causes flickering. Clear as rarely
			// as you possibly can.
			doOrDie("Clearing display for new line", pert.Clear())
		case <-ticker.C:

			// Update the display
			// Note that errors are not checked here.
			// Write errors to the display can occur, and should probably be
			// checked here. It muddies the water of the example, so are left
			// out. You should totally do that, though. Right?

			now := time.Now()
			_, week := now.ISOWeek()
			pert.Centered(0, fmt.Sprintf("%s %d", now.Format("Monday"), week))
			if now.Second()%2 == 0 {
				pert.Centered(1, now.Format("2006-01-02 15:04"))
			} else {
				pert.Centered(1, now.Format("2006-01-02 15 04"))
			}
			if len(lines) >= 2 {
				pert.Centered(2, lines[0])
				pert.Centered(3, lines[1])
			}
		}
	}
}
