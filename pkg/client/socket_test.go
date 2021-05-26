package client

import (
	"net"
	"net/textproto"
	"reflect"
	"testing"
)

func TestNewSocket(t *testing.T) {
	_, client := net.Pipe()
	defer client.Close()

	want := &Socket{
		conn: textproto.NewConn(client),
	}

	t.Run("can create socket", func(t *testing.T) {
		got := NewSocket(client)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewSocket() = %v, want %v", got, want)
		}
	})
}

func TestSocket_GetMessage(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	socket := NewSocket(client)

	want := "test"

	t.Run("can fetch message from socket", func(t *testing.T) {
		go func() {
			server.Write([]byte(want + "\n"))
		}()

		msg := socket.GetMessage()
		if string(msg) != want {
			t.Errorf("GetMessage() = %v, want %v", msg, want)
		}
	})
}
