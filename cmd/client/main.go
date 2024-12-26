package main

import (
	"fmt"
	"log"
	"net"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
)

var srkReadChan = make(chan []byte)
var srkWriteChan = make(chan []byte)
var sessionReadChan = make(chan []byte)
var sessionWriteChan = make(chan []byte)

func main() {
	a := app.New()
	w := a.NewWindow("Sessioner")

	// Start SRK server
	_, port := startSRKServer()

	// Add label for SRK server address
	srkAddr := widget.NewLabel(fmt.Sprintf("SRK server address: %d", port))

	// Add input for td2 server address
	td2Addr := widget.NewEntry()
	td2Addr.SetText("127.0.0.1:7424")

	// Add input for session id
	sessionID := widget.NewEntry()
	sessionID.SetPlaceHolder("Session ID")

	// Add button to connect to td2
	connectButton := widget.NewButton("Connect", func() {
		// Connect to td2
		err := connectToTd2(td2Addr.Text)
		if err != nil {
			log.Printf("Failed to connect to td2: %v", err)
			return
		}
	})

	// Add button to connect to session
	sessionButton := widget.NewButton("Connect to session", func() {
		// Connect to session over ws
		conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:8080/ws/%s", sessionID.Text), nil)
		if err != nil {
			return
		}

		log.Printf("Connected to session %s", sessionID.Text)

		// Start reading from session
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}
				sessionReadChan <- message

				select {
				case data := <-sessionWriteChan:
					conn.WriteMessage(websocket.TextMessage, data)
				default:
				}
			}
		}()
	})

	// Add to window
	content := container.NewVBox(
		srkAddr,
		td2Addr,
		sessionID,
		connectButton,
		sessionButton,
	)

	w.SetContent(content)
	w.ShowAndRun()
}
