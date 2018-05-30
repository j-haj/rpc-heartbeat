package main

import (
	"log"
	"sync"
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
		log.Printf("Received ID %d from heartbeat server - using heartbeat interval %d\n", c.ID, c.HeartbeatInterval)
	} else {
		log.Fatalf("error in join request: %v", err)
	}
}

func (c *Client) beginHeartbeat(client pb.HeartbeatClient) {
	for {
		log.Printf("Client: %v\n", *c)
		time.Sleep(time.Duration(c.HeartbeatInterval) * time.Nanosecond)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		msg := &pb.HeartbeatMessage{Id: c.ID, Time: time.Now().Unix()}
		response, err := client.Heartbeat(ctx, msg)
		if err != nil {
			log.Fatalf("failed to send heartbeat - %v", err)
		}
		log.Printf("Client %d received token %d from master\n", c.ID, response.Token)
	}
}

func main() {
	N := 50
	clients := make([]*Client, N)
	var wg sync.WaitGroup
	for _, c := range clients {
		opts := []grpc.DialOption{grpc.WithInsecure()}
		conn, err := grpc.Dial("localhost:10000", opts...)
		if err != nil {
			log.Printf("error - failed to establish connection: %v", err)
			return
		}
		client := pb.NewHeartbeatClient(conn)
		c = &Client{}
		c.RequestID(client)
		wg.Add(1)
		go c.beginHeartbeat(client)
	}
	wg.Wait()
}
