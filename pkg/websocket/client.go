package websocket

import "github.com/gorilla/websocket"

type Client struct {
	addr   string
	conn   *websocket.Conn
	inChan chan []byte
}

func NewClient(addr string) (*Client, error) {
	return &Client{
		addr:   addr,
		inChan: make(chan []byte),
	}, nil
}

func (c *Client) Start() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.addr, nil)
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
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.inChan <- msg
	}
}

func (c *Client) Send(msg []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}
