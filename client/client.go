package main

import (
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/j-haj/heartbeat/heartbeat"
)

type Client struct {
	ID                int64
	HeartbeatInterval int64
}

// Contacts heartbeat server and request to join network
func (c *Client) RequestID(client pb.HeartbeatClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rqstInfo := &pb.RequestInfo{}
	if msg, err := client.JoinRequest(ctx, rqstInfo); err == nil {
		c.ID = msg.Id
		c.HeartbeatInterval = msg.HeartbeatInterval
		log.Printf("Received ID %d from heartbeat server\n", c.ID)
	} else {
		log.Fatalf("error in join request!")
	}
}

func (c *Client) beginHeartbeat(client pb.HeartbeatClient) {
	for {
		time.Sleep(time.Duration(c.HeartbeatInterval) * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		msg := &pb.HeartbeatMessage{Id: c.ID, Time: time.Now().Unix()}
		status, err := client.Heartbeat(ctx, msg)
		if status != nil {
			log.Printf("Client %d Received OK from master\n", c.ID)
		}
		if err != nil {
			log.Fatalf("failed to send heartbeat - %v", err)
		}
	}
}

func main() {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("localhost:10000", opts...)
	if err != nil {
		log.Printf("error - failed to establish connection: %v", err)
		return
	}
	client := pb.NewHeartbeatClient(conn)

	N := 5
	clients := make([]Client, N)
	for _, c := range clients {
		c.RequestID(client)
		go c.beginHeartbeat(client)
	}
}
