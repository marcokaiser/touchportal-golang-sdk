package client

import (
	"bufio"
	"log"
	"net"
	"time"
)

type Socket struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	retries int
}

func NewSocket(c net.Conn) *Socket {
	return &Socket{
		conn:   c,
		reader: bufio.NewReader(c),
		writer: bufio.NewWriter(c),
	}
}

// GetMessage reads a single line from the connection. It blocks for a duration but plays nice and
// returns after the a 300ms deadline so you can handle graceful shutdowns.
func (s *Socket) GetMessage() ([]byte, error) {
	// handle retry backoff
	time.Sleep(time.Duration(300*time.Millisecond) * time.Duration(s.retries))

	err := s.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if err != nil {
		log.Printf("error whilst setting connection read deadline: %v", err)
	}

	bytes, err := s.reader.ReadBytes('\n')
	if err != nil {
		if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
			log.Printf("error whilst reading from TCP socket: %+v\n", err)

			if s.retries++; s.retries > 10 {
				return nil, err
			}
		}

		return nil, nil
	}

	s.retries = 0

	return bytes, nil
}

// SendMessage is a blocking call to send a message out via the socket
func (s *Socket) SendMessage(m []byte) error {
	m = append(m, '\n')

	i, err := s.writer.Write(m)
	if err != nil {
		return err
	}

	if i > 0 {
		return s.writer.Flush()
	}

	return nil
}

// Close will close the underlying communication socket
func (s *Socket) Close() {
	s.conn.Close()
}
