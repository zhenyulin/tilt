package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/windmilleng/tilt/internal/pane/proto"
)

type paneCmd struct{}

func (c *paneCmd) register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pane",
		Short: "connect a pane to the app",
	}

	return cmd
}

func (c *paneCmd) run(ctx context.Context, args []string) error {
	fmt.Printf("hello pane\n")

	cmd := exec.Command("tty")
	cmd.Stdin = os.Stdin
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	outputString := strings.TrimRight(string(output), "\n")
	log.Printf("hrm %q %s", outputString, err)

	ttyOut, err := os.OpenFile(outputString, syscall.O_WRONLY, 0)
	if err != nil {
		return err
	}
	ttyIn, err := os.OpenFile(outputString, syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}

	files := []*os.File{os.Stdin, os.Stdout, os.Stderr, ttyIn, ttyOut}

	socketPath, err := proto.LocateSocket()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(
		socketPath,
		grpc.WithInsecure(),
		grpc.WithDialer(UnixDial),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	cl := proto.NewPaneClient(conn)

	dir, err := ioutil.TempDir("", "tilt-pane-")
	if err != nil {
		return err
	}

	fdSocketPath := filepath.Join(dir, "socket")

	fdListener, err := net.ListenUnix("unix", &net.UnixAddr{Name: fdSocketPath, Net: "unix"})
	if err != nil {
		return err
	}

	stream, err := cl.ConnectPane(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(
		&proto.PaneControl{
			Control: &proto.PaneControl_Connect{Connect: &proto.ConnectRequest{
				FdSocketPath: fdSocketPath,
				TtyPath:      outputString,
			}},
		}); err != nil {
		return err
	}

	// now we need to wait for one of two things to happen:
	// 1) tilt daemon connects to us (and we send it our fds)
	// 2) stream closes

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return sendFDs(fdListener, files)
	})

	g.Go(func() error {
		return stream.RecvMsg(nil)
	})

	g.Go(func() error {
		winchCh := make(chan os.Signal, 10) // 10 is enough that we don't overflow the buffer
		signal.Notify(winchCh, syscall.SIGWINCH)
		defer signal.Stop(winchCh)
		for {
			select {
			case <-winchCh:
				log.Printf("sending winch")
				if err := stream.Send(
					&proto.PaneControl{
						Control: &proto.PaneControl_WindowChange{WindowChange: &proto.WindowChange{}},
					}); err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	return g.Wait()
}

func UnixDial(addr string, timeout time.Duration) (net.Conn, error) {
	// TODO(dbentley): do timeouts right
	return net.DialTimeout("unix", addr, 100*time.Millisecond)
}

func sendFDs(fdListener *net.UnixListener, fs []*os.File) error {
	defer fdListener.Close()
	fdConn, err := fdListener.AcceptUnix()
	log.Printf("got a client who wants fds %v", fdConn)
	if err != nil {
		return err
	}
	connFd, err := fdConn.File()
	if err != nil {
		return err
	}
	defer fdConn.Close()

	fds := make([]int, len(fs))
	for i, f := range fs {
		fds[i] = int(f.Fd())
	}

	rights := syscall.UnixRights(fds...)
	return syscall.Sendmsg(int(connFd.Fd()), nil, rights, nil, 0)
}
