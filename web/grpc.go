package web

import (
	"fmt"
	"net"
	"time"

	"github.com/im-kulikov/helium/internal"
	"google.golang.org/grpc"
)

type (
	gRPC struct {
		skipErrors      bool
		address         string
		network         string
		server          *grpc.Server
		shutdownTimeout time.Duration
	}

	// GRPCOption allows changing default gRPC
	// service settings.
	GRPCOption func(g *gRPC)
)

const (
	// ErrEmptyGRPCServer is raised when called NewGRPCService
	// or gRPC methods with empty grpc.Server.
	ErrEmptyGRPCServer = internal.Error("empty gRPC server")

	// ErrEmptyGRPCAddress is raised when passed empty address to NewGRPCService.
	ErrEmptyGRPCAddress = internal.Error("empty gRPC address")
)

// GRPCSkipErrors allows to skip any errors
func GRPCSkipErrors() GRPCOption {
	return func(g *gRPC) {
		g.skipErrors = true
	}
}

// GRPCListenAddress allows to change network for net.Listener.
func GRPCListenAddress(addr string) GRPCOption {
	return func(g *gRPC) {
		g.address = addr
	}
}

// GRPCListenNetwork allows to change default (tcp)
// network for net.Listener.
func GRPCListenNetwork(network string) GRPCOption {
	return func(g *gRPC) {
		g.network = network
	}
}

// GRPCShutdownTimeout changes default shutdown timeout.
func GRPCShutdownTimeout(v time.Duration) GRPCOption {
	return func(g *gRPC) {
		g.shutdownTimeout = v
	}
}

// NewGRPCService creates gRPC service with passed gRPC options.
// If something went wrong it returns an error.
func NewGRPCService(serve *grpc.Server, opts ...GRPCOption) (Service, error) {
	if serve == nil {
		return nil, ErrEmptyGRPCServer
	}

	s := &gRPC{
		server:          serve,
		network:         "tcp",
		skipErrors:      false,
		shutdownTimeout: time.Second * 30,
	}

	for i := range opts {
		opts[i](s)
	}

	if s.address == "" {
		return nil, ErrEmptyGRPCAddress
	}

	return s, nil
}

// Start tries to start gRPC service.
// If something went wrong it returns an error.
// If could not start server panics.
func (g *gRPC) Start() error {
	var (
		err error
		lis net.Listener
	)

	if g.server == nil {
		return g.catch(ErrEmptyGRPCServer)
	} else if lis, err = net.Listen(g.network, g.address); err != nil {
		return g.catch(err)
	}

	go func() {
		if err := g.catch(g.server.Serve(lis)); err != nil {
			fmt.Printf("could not start grpc.Server: %v\n", err)
			fatal(2)
		}
	}()

	return nil
}

// Stop tries to stop gRPC service.
func (g *gRPC) Stop() error {
	if g.server == nil {
		return g.catch(ErrEmptyGRPCServer)
	}

	g.server.GracefulStop()
	return nil
}

func (g *gRPC) catch(err error) error {
	if g.skipErrors || err == grpc.ErrServerStopped {
		return nil
	}

	return err
}
