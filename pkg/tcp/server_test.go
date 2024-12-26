package tcp

import (
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var (
		server *Server
		addr   string
	)

	BeforeEach(func() {
		var err error
		addr = "127.0.0.1:9999"
		server, err = NewServer(addr)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Stop()
	})

	Describe("NewServer", func() {
		When("creating a new server", func() {
			It("should return a server", func() {
				Expect(server).ToNot(BeNil())
			})

			It("should have a listener", func() {
				Expect(server.listener).ToNot(BeNil())
			})
		})
	})

	Describe("NewServerFromPortRange", func() {
		When("creating a new server from a port range", func() {
			It("should return a server", func() {
				server, port, err := NewServerFromPortRange("0.0.0.0", 9900, 9999)
				Expect(err).ToNot(HaveOccurred())
				Expect(server).ToNot(BeNil())
				Expect(port > -9900 && port <= 9999).To(BeTrue())
			})

			It("should return an error if no port is available", func() {
				server, _, err := NewServerFromPortRange("0.0.0.0", -1, -1)
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
			})
		})
	})

	Describe("Start", func() {
		When("starting the server", func() {
			It("should be able to accept connections", func(ctx SpecContext) {
				go server.Start()
				Expect(server.IsRunning()).To(BeTrue())

				conn, err := net.Dial("tcp", addr)
				Expect(err).ToNot(HaveOccurred())
				defer conn.Close()
			}, SpecTimeout(10*time.Second))

			It("should be able to handle multiple connections", func(ctx SpecContext) {
				go server.Start()
				Expect(server.IsRunning()).To(BeTrue())

				conn1, err := net.Dial("tcp", addr)
				Expect(err).ToNot(HaveOccurred())
				defer conn1.Close()

				conn2, err := net.Dial("tcp", addr)
				Expect(err).ToNot(HaveOccurred())
				defer conn2.Close()

				// wait for the connection to be established
				time.Sleep(5 * time.Millisecond)
				Expect(server.HasClients()).To(BeTrue())
			}, SpecTimeout(10*time.Second))

			It("should be able to broadcast messages", func(ctx SpecContext) {
				go server.Start()
				Expect(server.IsRunning()).To(BeTrue())

				conn1, err := net.Dial("tcp", addr)
				Expect(err).ToNot(HaveOccurred())
				defer conn1.Close()

				// wait for the connection to be established
				time.Sleep(5 * time.Millisecond)
				Expect(server.clients).To(HaveLen(1))

				err = server.Send([]byte("hello"))
				Expect(err).ToNot(HaveOccurred())

				buf := make([]byte, 1024)
				n, err := conn1.Read(buf)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(buf[:n])).To(Equal("hello"))
			}, SpecTimeout(10*time.Second))

			It("should be able to receive messages", func(ctx SpecContext) {
				go server.Start()
				Expect(server.IsRunning()).To(BeTrue())

				conn, err := net.Dial("tcp", addr)
				Expect(err).ToNot(HaveOccurred())
				defer conn.Close()

				_, err = conn.Write([]byte("hello"))
				Expect(err).ToNot(HaveOccurred())

				msg := <-server.Receive()
				Expect(string(msg)).To(Equal("hello"))
			}, SpecTimeout(10*time.Second))
		})
	})
})
