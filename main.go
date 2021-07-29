package main

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

type Msg struct {
	UserId		string 		`json:"user_id"`
	Text		string		`json:"text"`
	State		string		`json:"state"`
	NameSpace	string		`json:"name_space"`
	Rooms		[]string	`json:"rooms"`
}

func main()  {
	server := socketio.NewServer(nil)
	server.OnConnect("/", func(s socketio.Conn) error {
		msg := Msg{s.ID(), "on connect", "notice", "", nil}
		s.SetContext("")
		s.Emit("res", msg)
		fmt.Println("connected /:", s.ID())
		return nil
	})

	server.OnEvent("/", "join", func(s socketio.Conn, room string) {
		s.Join(room)
		msg := Msg{ s.ID(), "<= " + " join " + room, "state", s.Namespace(), s.Rooms()}
		fmt.Println("/:join", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "res", msg)
	})

	server.OnEvent("/", "leave", func(s socketio.Conn, room string) {
		s.Leave(room)
		msg := Msg{ s.ID(), "<= " + " leave " + room, "state", s.Namespace(), s.Rooms()}
		fmt.Println("/:lear", room, s.Namespace(), s.Rooms())
		server.BroadcastToRoom("/", room, "res", msg)
	})

	server.OnEvent("/", "chat", func(s socketio.Conn, msg string) {
		res := Msg{s.ID(), "<= ", "normal", s.Namespace(), s.Rooms()}
		s.SetContext(res)
		fmt.Println("/:chat received", msg, s.Namespace(), s.Rooms(), s.Rooms())
		rooms := s.Rooms()
		if len(rooms) > 0 {
			fmt.Println("broadcast to ", rooms)
			for i := range rooms {
				server.BroadcastToRoom("/", rooms[i], "res", res)
			}
		}
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("/:notice", msg)
		s.Emit("reply", "have "+ msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		fmt.Println("/chat:msg received", msg)
		return "recv" + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn, msg string) string {
		last := s.Context().(Msg)
		s.Emit("bye", last)
		res := Msg{s.ID(), "<= " + s.ID() +" leave", "state", s.Namespace(), s.Rooms()}
		rooms := s.Rooms()
		s.LeaveAll()
		s.Close()
		if len(rooms) > 0 {
			fmt.Println("broadcast to ", rooms)
			for i := range rooms {
				server.BroadcastToRoom("/", rooms[i], "res", res)
			}
		}
		fmt.Printf("/:bye last context: %+v \n", last)
		return last.Text
	})

	server.OnError("/", func(s socketio.Conn, err error) {
		fmt.Println("/:error", err)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("/:closed", s.ID(), reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", corsMiddleware(server))
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	log.Println("Severing at localhost:5000...")

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func hello(wr http.ResponseWriter, r *http.Request) {
	wr.Write([]byte("hello"))
}

func corsMiddleware(next http.Handler) http.Handler  {
	return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		wr.Header().Set("Access-Control-Allow-Origin", origin)
		wr.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(wr, r)
	})
}
