package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/kacpermalachowski/td2-srk-multiplex/pkg/tcp"
	"github.com/kacpermalachowski/td2-srk-multiplex/pkg/websocket"
)

type UIEvent struct {
	EventType string
	Element   string
}

const (
	remoteAddr = "localhost:8080"
)

var (
	srkServer     *tcp.Server
	td2Client     *tcp.Client
	sessionClient *websocket.Client
	uiEventChan   = make(chan UIEvent)
)

func main() {
	a := app.New()
	w := a.NewWindow("Sessioner")

	// Start SRK server
	server, port, err := tcp.NewServerFromPortRange("0.0.0.0", 7424, 7524)
	if err != nil {
		log.Fatalf("error starting SRK server: %v", err)
	}
	srkServer = server
	go srkServer.Start()
	defer srkServer.Stop()

	// Add label for SRK server address
	srkAddr := widget.NewLabel(fmt.Sprintf("SRK server address: %d", port))

	// Add input for td2 server address
	td2Addr := widget.NewEntry()
	td2Addr.SetText("127.0.0.1:7424")

	// Add input for session id
	sessionID := widget.NewEntry()
	sessionID.SetPlaceHolder("Session ID")

	// Add button to connect to td2
	connectButton := widget.NewButton("Connect to TD2", func() {
		uiEventChan <- UIEvent{
			EventType: "tapped",
			Element:   "connectButton",
		}
	})

	// Add button to connect to session
	sessionButton := widget.NewButton("Connect to session", func() {
		uiEventChan <- UIEvent{
			EventType: "tapped",
			Element:   "sessionButton",
		}
	})

	// Add to window
	content := container.New(
		layout.NewVBoxLayout(),
		srkAddr,
		td2Addr,
		layout.NewSpacer(),
		sessionID,
		layout.NewSpacer(),
		connectButton,
		sessionButton,
	)

	go func() {
		for {
			var err error
			select {
			case event := <-uiEventChan:
				switch event.Element {
				case "connectButton":
					if td2Client != nil {
						td2Client.Stop()
						td2Client = nil
						connectButton.SetText("Connect to TD2")
					} else {
						addr := td2Addr.Text
						td2Client, err = tcp.NewClient(addr)
						if err != nil {
							log.Printf("error connecting to TD2: %v", err)
							continue
						}
						err = td2Client.Start()
						if err != nil {
							log.Printf("error starting TD2 client: %v", err)
							continue
						}
						connectButton.SetText("Disconnect from TD2")
					}
				case "sessionButton":
					if sessionClient != nil {
						sessionClient.Stop()
						sessionClient = nil
						sessionButton.SetText("Connect to session")
					} else {
						sessionClient, err = websocket.NewClient(fmt.Sprintf("ws://%s/ws/%s", remoteAddr, sessionID.Text))
						if err != nil {
							log.Printf("error connecting to session: %v", err)
							continue
						}

						err = sessionClient.Start()
						if err != nil {
							log.Printf("error starting session client: %v", err)
							continue
						}

						sessionButton.SetText("Disconnect from session")
					}
				}
			case msg := <-srkServer.Receive():
				// Forward message to TD2
				if td2Client != nil {
					td2Client.Send(msg)
				}
			default:
				if sessionClient != nil {
					select {
					case msg := <-sessionClient.Receive():
						// Forward message to td2
						if td2Client != nil {
							err := td2Client.Send(msg)
							if err != nil {
								log.Printf("error sending message to td2: %v", err)
							}
						}

						// Forward message to SRK
						if srkServer != nil {
							srkServer.Send(msg)
						}
					default:
					}
				}
				if td2Client != nil {
					select {
					case msg := <-td2Client.Receive():
						// Forward message to SRK
						if srkServer != nil {
							srkServer.Send(msg)
						}

						// Forward message to session
						if sessionClient != nil {
							err := sessionClient.Send(msg)
							if err != nil {
								log.Printf("error sending message to session: %v", err)
							}
						}
					default:
					}
				}
			}
		}
	}()

	w.SetContent(container.New(layout.NewCustomPaddedLayout(10, 10, 20, 20), content))
	w.ShowAndRun()
}
