package hud

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"

	"log"

	"google.golang.org/grpc"

	"github.com/windmilleng/tilt/internal/pane/proto"
	"github.com/windmilleng/tilt/internal/state"
	"github.com/windmilleng/tilt/internal/state/states"
)

func NewTTYPaneServer() (*PaneServerProvider, error) {
	// TODO(dbentley): bad! should use wire to inject this, but DI and servers are hard
	store, err := states.NewStateStore(context.TODO())
	if err != nil {
		return nil, err
	}
	controlCh := make(chan state.ControlEvent)
	socketPath, err := proto.LocateSocket()
	if err != nil {
		return nil, err
	}

	l, err := UnixListen(socketPath)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	a := &PaneServerAdapter{stateReader: store, controlCh: controlCh}

	proto.RegisterPaneServer(grpcServer, a)

	// TODO(dbentley): deal with error
	go func() {
		err := grpcServer.Serve(l)
		if err != nil {
			log.Printf("hud server error: %v", err)
		}
	}()

	return &PaneServerProvider{store, controlCh}, nil
}

type PaneServerProvider struct {
	state.StateWriter
	ch chan state.ControlEvent
}

func (p *PaneServerProvider) Ch() <-chan state.ControlEvent {
	return p.ch
}

type PaneServerAdapter struct {
	stateReader state.StateReader
	controlCh   chan state.ControlEvent
}

func (a *PaneServerAdapter) ConnectPane(stream proto.Pane_ConnectPaneServer) error {
	ctx := stream.Context()

	evs, err := a.stateReader.Subscribe(ctx)
	if err != nil {
		return err
	}

	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	connectMsg := msg.GetConnect()
	if connectMsg == nil {
		return fmt.Errorf("expected a connect msg; got %T %v", msg, msg)
	}

	fdConn, err := net.DialUnix("unix", &net.UnixAddr{Name: "", Net: "unix"}, &net.UnixAddr{Name: connectMsg.FdSocketPath, Net: "unix"})
	if err != nil {
		return err
	}

	fdConnF, err := fdConn.File()
	if err != nil {
		return err
	}
	num := 5 // stdin, stdout, stderr, readTTY, writeTTY
	buf := make([]byte, syscall.CmsgSpace(num*4))
	_, _, _, _, err = syscall.Recvmsg(int(fdConnF.Fd()), nil, buf, 0)
	if err != nil {
		return err
	}

	msgs, err := syscall.ParseSocketControlMessage(buf)
	if err != nil {
		return err
	}

	var fs []*os.File
	for _, msg := range msgs {
		fds, err := syscall.ParseUnixRights(&msg)
		if err != nil {
			return err
		}
		for _, fd := range fds {
			fs = append(fs, os.NewFile(uintptr(fd), "/dev/null"))
		}
	}

	if len(fs) != 5 {
		return fmt.Errorf("expected 5 files; got %v", len(fs))
	}

	_, err = fmt.Fprintf(fs[1], "Hullo\n")
	if err != nil {
		log.Printf("whoops %v", err)
	}

	fs[0].Close()
	fs[1].Close()
	fs[2].Close()

	winchCh := make(chan os.Signal)
	streamErrCh := make(chan error)

	hud, err := NewHud(evs, a.controlCh)
	if err != nil {
		return err
	}
	hudDoneCh := make(chan error)
	hudCtx, cancelHud := context.WithCancel(ctx)
	go func() {
		hudDoneCh <- hud.Run(hudCtx, fs[3], fs[4], winchCh)
	}()

	go func() {
		for {
			_, err := stream.Recv() // assume it's a window change message
			if err != nil {
				streamErrCh <- err
				return
			}
			winchCh <- syscall.SIGWINCH
		}
	}()

	select {
	case err := <-hudDoneCh:
		log.Printf("hud is done")
		return err
	case err := <-streamErrCh:
		cancelHud()
		<-hudDoneCh
		return err
	}
}
