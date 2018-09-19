package hud

import (
	"context"
	"fmt"
	"os"

	"github.com/windmilleng/tilt/internal/state"
)

// Byte-oriented output

func NewByteTtyPane(ctx context.Context, st state.Subscription, stdin, stdout, stderr, readTty, writeTty *os.File) error {
	defer stdin.Close()
	defer stdout.Close()
	defer stderr.Close()
	defer readTty.Close()
	defer writeTty.Close()

	fmt.Fprintf(stdout, "Hello pane\n")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-st.Ch():
			fmt.Fprintf(stdout, "Event!")
		}
	}
}
