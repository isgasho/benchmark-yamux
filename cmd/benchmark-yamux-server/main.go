package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/ttrpc"
	"github.com/fuweid/benchmark-yamux/api"
	"github.com/hashicorp/yamux"
)

const socket = "/tmp/benchmark-yamux-server.sock"

func main() {
	os.Remove(socket)

	s, err := ttrpc.NewServer()
	if err != nil {
		log.Fatalf("failed to new ttrpc server: %v", err)
	}
	defer s.Close()

	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalf("failed to create a pipe: %v", err)
	}

	go func() {
		cmd := exec.Command("benchmark-yamux-app")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGKILL,
		}
		cmd.Stdout = w
		cmd.Stderr = w

		if err := cmd.Run(); err != nil {
			log.Fatalf("unexpected error from app: %v", err)
		}
	}()

	api.RegisterUnknownHubService(s, newUnknownServer(r))

	l, err := net.Listen("unix", socket)
	if err != nil {
		log.Fatalf("failed to listen %s: %v", socket, err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}

		// setup yamux session
		cfg := yamux.DefaultConfig()
		cfg.EnableKeepAlive = false

		session, err := yamux.Server(conn, cfg)
		if err != nil {
			log.Fatalf("failed to setup yamux session: %v", err)
		}

		// loop to accept stream
		go func(session *yamux.Session) {
			defer session.Close()

			err := s.Serve(context.Background(), session)
			log.Printf("unexpected error %v", err)
		}(session)
	}

}

const maxReadLen = 8 * 1024

func newUnknownServer(data io.Reader) *unknownServer {
	return &unknownServer{
		data: data,
	}
}

type unknownServer struct {
	data io.Reader
}

func (s *unknownServer) Read(ctx context.Context, r *api.ReadRequest) (*api.ReadResponse, error) {
	min := minUint32(uint32(maxReadLen), r.GetLen())
	if min == 0 {
		min = uint32(maxReadLen)
	}

	b := make([]byte, min)
	n, err := s.data.Read(b)
	if err != nil {
		return nil, err
	}
	return &api.ReadResponse{
		Data: b[:n],
	}, nil
}

func minUint32(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}
