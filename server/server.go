package main

import (
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/j-haj/heartbeat/heartbeat"
)

type heartbeatServer struct {
	registeredClients map[int64]int64 // Tracks ClientID and timestamp of last received hearbeat
	mu                sync.Mutex      // protects registeredClients
	lifetimeThreshold int64           // Threshold that determines when a client is down
	clientCount       int64           // Generates new ClientIDs
}

// Heartbeat is called when the heartbeat server receives a heartbeat from a client
func (s *heartbeatServer) Heartbeat(ctx context.Context, msg *pb.HeartbeatMessage) (*pb.Status, error) {
	// Register heartbeat
	log.Printf("Received heartbeat from %d with timestamp %d\n", msg.Id, msg.Time)
	if _, ok := s.registeredClients[msg.Id]; ok {
		s.mu.Lock()
		s.registeredClients[msg.Id] = msg.Time
		s.mu.Unlock()
		log.Printf("Client %d checked in successfully\n")
		return &pb.Status{}, nil
	}
	log.Printf("Error - client ID %d unrecognized\n", msg.Id)
	return nil, nil
}

func (s *heartbeatServer) JoinRequest(ctx context.Context, requestInfo *pb.RequestInfo) (*pb.ClientId, error) {
	id := &pb.ClientId{Id: s.clientCount + 1, HeartbeatInterval: s.lifetimeThreshold}
	log.Printf("Received join request - sending ID %d\n", id.Id)
	s.mu.Lock()
	s.clientCount++
	s.mu.Unlock()
	return id, nil
}

func newServer(threshold int64) *heartbeatServer {
	log.Printf("Creating a new heartbeat server with heartbeat threshold of %d\n", threshold)
	s := &heartbeatServer{registeredClients: make(map[int64]int64), lifetimeThreshold: threshold}
	return s
}

func main() {
	lis, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		log.Printf("Failed to open port 1000 - %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterHeartbeatServer(grpcServer, newServer(int64(3*time.Second)))
	log.Printf("Starting heartbeat server\n")
	grpcServer.Serve(lis)
}
