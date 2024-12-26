package tcp

import (
	"fmt"
	"net"
)

type Client struct {
	addr   string
	conn   net.Conn
	inChan chan []byte
}

func NewClient(addr string) (*Client, error) {
	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	return &Client{
		addr:   addr,
		inChan: make(chan []byte),
	}, nil
}

func (c *Client) Start() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.read()

	return nil
}

func (c *Client) Stop() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.inChan)
}

func (c *Client) Receive() <-chan []byte {
	return c.inChan
}

func (c *Client) read() {
	for {
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			break
		}
		c.inChan <- buf[:n]
	}
}

func (c *Client) Send(msg []byte) error {
	_, err := c.conn.Write(msg)
	return err
}
