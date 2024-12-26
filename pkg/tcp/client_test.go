package tcp

import (
	"github.com/kacpermalachowski/td2-srk-multiplex/internal/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		client     *Client
		testServer *testserver.TestTCPServer
		addr       string
	)

	BeforeEach(func() {
		var err error
		addr = "127.0.0.1:9000"
		testServer = testserver.New(addr)
		srvAddr, err := testServer.Start()
		Expect(err).ToNot(HaveOccurred())
		client, err = NewClient(srvAddr)
		Expect(err).ToNot(HaveOccurred())
		go client.Start()
	})

	AfterEach(func() {
		client.Stop()
		testServer.Stop()
	})

	Describe("NewClient", func() {
		When("creating a new client", func() {
			It("should return a client", func() {
				Expect(client).ToNot(BeNil())
			})

			It("should have an address", func() {
				Expect(client.addr).To(Equal(addr))
			})
		})

		When("creating a new client with an invalid address", func() {
			It("should return an error", func() {
				_, err := NewClient("invalid")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Start", func() {
		When("starting the client", func() {
			It("should be able to connect to the server", func() {
				err := client.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(client.conn).ToNot(BeNil())
			})

			It("should be able to write to the server", func() {
				err := client.Start()
				Expect(err).ToNot(HaveOccurred())
				err = client.Send([]byte("test"))
				Expect(err).ToNot(HaveOccurred())

				Expect(testServer.ReceivedDataChan).To(Receive(Equal("test")))
			})
		})
	})
})
