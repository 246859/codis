package test

import (
	"bufio"
	"context"
	"github.com/246859/codis/coco"
	"github.com/246859/codis/pkg/logger"
	"net"
	"testing"
	"time"
)

func TestCocoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	server := coco.NewServer(ctx)
	listen, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Error(err)
	}
	go server.Serve(listen, &coco.CocoHandler{})
	for i := 0; i < 20; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			logger.Error("client dial error: ", err)
			continue
		}
		conn.Write([]byte("hello world" + "\n"))
		reader := bufio.NewReader(conn)
		readString, err := reader.ReadString('\n')
		if err != nil {
			logger.Error("client read error: ", err)
			continue
		}
		logger.Info(readString)
	}
	server.Shutdown()
	time.Sleep(2 * time.Second)
}
