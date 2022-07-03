package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/ttrpc"
	"github.com/fuweid/benchmark-yamux/api"
	"github.com/hashicorp/yamux"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("required virtio-serial-port path")
	}

	portPath := os.Args[1]

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

	portChannel, err := os.OpenFile(portPath, os.O_RDWR, os.ModeDevice)
	if err != nil {
		log.Fatalf("unexpected error when open portPath %s: %v", portPath, err)
	}

	// setup yamux session
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = false

	session, err := yamux.Server(portChannel, cfg)
	if err != nil {
		log.Fatalf("failed to setup yamux session: %v", err)
	}

	err = s.Serve(context.Background(), session)
	log.Printf("unexpected error %v", err)
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
