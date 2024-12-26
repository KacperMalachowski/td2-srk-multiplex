package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kacpermalachowski/td2-srk-multiplex/internal/hub"
)

var sessions = make(map[string]*hub.Hub)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	router := mux.NewRouter()
	router.HandleFunc("/ws/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if _, ok := sessions[id]; !ok {
			sessions[id] = hub.New()
			go sessions[id].Run()
		}

		log.Printf("Session %s: %v\n", id, sessions[id])
		hub.Serve(sessions[id], w, r)
	})

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", router)
}
