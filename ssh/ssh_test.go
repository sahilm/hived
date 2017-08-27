package ssh_test

import (
	"testing"

	"fmt"

	"time"

	"github.com/sahilm/hived/ssh"
)

func TestListen(t *testing.T) {
	server, err := ssh.NewServer("0.0.0.0:2222")
	if err != nil {
		t.Fatal(err)
	}

	connections, errors, err := server.Serve()

	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for conn := range connections {
			fmt.Println(conn)
		}
	}()

	go func() {
		for e := range errors {
			fmt.Println(e)
		}
	}()

	time.Sleep(5 * time.Minute)
	server.Stop()
}
