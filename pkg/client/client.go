package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/context"
)

const (
	tpPort = 12136
	tpHost = "127.0.0.1"
)

type Client struct {
	socket      *Socket
	incoming    chan []byte
	fetchStop   chan bool
	processStop chan bool
	ready       chan bool

	handlers   map[ClientMessageType][]func(e interface{})
	processors map[ClientMessageType]func(msg json.RawMessage) (interface{}, error)
}

func NewClient() *Client {
	c := &Client{
		incoming:    make(chan []byte, 5),
		fetchStop:   make(chan bool),
		processStop: make(chan bool),
		ready:       make(chan bool),
		handlers:    make(map[ClientMessageType][]func(event interface{})),
		processors:  make(map[ClientMessageType]func(msg json.RawMessage) (interface{}, error)),
	}

	c.registerDefaultMessageProcessors()

	return c
}

func (c *Client) AddMessageHandler(msgType ClientMessageType, handler func(e interface{})) {
	if _, contains := c.handlers[msgType]; !contains {
		c.handlers[msgType] = []func(event interface{}){}
	}

	c.handlers[msgType] = append(c.handlers[msgType], handler)
}

func (c *Client) Ready() <-chan bool {
	return c.ready
}

func (c *Client) Run(ctx context.Context) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tpHost, tpPort))
	if err != nil {
		log.Fatalf("unable to connect to touchportal: %v. exiting...", err)
	}
	defer conn.Close()

	c.socket = NewSocket(conn)

	// by closing the ready channel we're telling any observers that enough of
	// this client has started that they can begin using it
	close(c.ready)

	// start the message handling stack
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go c.fetchIncomingMessage(wg)
	go c.processIncomingMessages(wg)

	// Watch for the context cancellation so we can ask our
	// goroutines to exit
	go func() {
		<-ctx.Done()
		c.Close()
	}()

	// wait for goroutines to exit
	wg.Wait()
}

func (c *Client) Close() {
	close(c.fetchStop)
	close(c.processStop)
	close(c.incoming)
}

func (c *Client) Dispatch(mType ClientMessageType, event interface{}) {
	for _, handler := range c.handlers[mType] {
		handler(event)
	}
}

func (c *Client) SendMessage(m interface{}) error {
	return c.socket.SendMessage(toJson(m))
}

func (c *Client) fetchIncomingMessage(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-c.fetchStop:
			return
		default:
			msg, err := c.socket.GetMessage()
			if err != nil {
				log.Fatalf("the connection has been lost with touchportal. exiting...")
			}

			if msg != nil {
				c.incoming <- msg
			}
		}
	}
}

func (c *Client) processIncomingMessages(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-c.processStop:
			return
		case msg := <-c.incoming:
			c.processMessage(msg)
		}
	}
}

func (c *Client) processMessage(msg json.RawMessage) {
	var m Message
	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Printf("unable to marshall message and discern type: %v\n", err)
		return
	}

	mType := ClientMessageType(m.Type)
	processor, ok := c.processors[mType]
	if !ok {
		log.Printf("type of message \"%s\" not currently handled\n", mType)
		return
	}

	pm, err := processor(msg)
	if err != nil {
		log.Printf("unable to marshall message into type %s: %v\n", mType, err)
		return
	}

	c.Dispatch(mType, pm)
}

func toJson(msg interface{}) []byte {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("unable to marshal message struct to string %v", msg)
	}

	return msgBytes
}
