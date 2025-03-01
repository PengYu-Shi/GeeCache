package gee

import (
	pb "GeeCacheNode/gee/geecachepb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedGeeServerServer
}

func (s *server) Get(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	group, err := GetGroup(req.Group)
	if err != nil {
		res := &pb.Response{
			Value: make([]byte, 1),
			Err:   err.Error(),
		}
		return res, err
	}
	value, err := group.Get(req.Key)
	if err != nil {
		res := &pb.Response{
			Value: value.ByteSlice(),
			Err:   err.Error(),
		}
		return res, err
	} else {
		res := &pb.Response{
			Value: value.ByteSlice(),
			Err:   " ",
		}
		return res, nil
	}

}

type loggingListener struct {
	net.Listener
}

func (l *loggingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	log.Printf("Accepted connection from: %v", conn.RemoteAddr())
	return conn, nil
}

func Init(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen port: %v, error: %v", address, err)
	}
	loggingListener := &loggingListener{listener}
	s := grpc.NewServer()
	pb.RegisterGeeServerServer(s, &server{})
	fmt.Printf("Server is listening on port: %v", address)
	if err := s.Serve(loggingListener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return nil
}
