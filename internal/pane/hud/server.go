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

func NewTTYPaneServer() (state.StateWriter, error) {
	// TODO(dbentley): bad! should use wire to inject this, but DI and servers are hard
	store, err := states.NewStateStore(context.TODO())
	if err != nil {
		return nil, err
	}
	socketPath, err := proto.LocateSocket()
	if err != nil {
		return nil, err
	}

	l, err := UnixListen(socketPath)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	a := &PaneServerAdapter{}

	proto.RegisterPaneServer(grpcServer, a)

	// TODO(dbentley): deal with error
	go func() {
		err := grpcServer.Serve(l)
		if err != nil {
			log.Printf("hud server error: %v", err)
		}
	}()

	return store, nil
}

type PaneServerAdapter struct {
	stateReader state.StateReader
}

func (a *PaneServerAdapter) Connect(req *proto.ConnectRequest, stream proto.Pane_ConnectServer) error {
	ctx := stream.Context()

	sub, err := a.stateReader.Subscribe(ctx)
	if err != nil {
		return err
	}
	fdConn, err := net.DialUnix("unix", &net.UnixAddr{Name: "", Net: "unix"}, &net.UnixAddr{Name: req.FdSocketPath, Net: "unix"})
	if err != nil {
		return err
	}

	fdConnF, err := fdConn.File()
	if err != nil {
		return err
	}
	num := 5
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

	// TODO(dbentley)
	return NewByteTtyPane(ctx, sub, fs[0], fs[1], fs[2], fs[3], fs[4])
}

// func f() {
// 	screen, err := tcell.NewTerminfoScreenFromTty(fs[3], fs[4])
// 	if err != nil {
// 		log.Printf("can't start screen")
// 	}

// 	if err := screen.Init(); err != nil {
// 		log.Printf("can't init screen %v", err)
// 	}
// }
