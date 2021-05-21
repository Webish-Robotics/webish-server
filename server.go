/**
======================================================================================================
EV3 server-side (socket-brain) handles all the sockets requests
======================================================================================================
* 	Written in GoLang
*	@author Vrinceanu Radu-Tudor, student @ National College "Vasile Alecsandri", Galati, Romania
*
*	THIS SOFTWARE IS AS IT IS ANY MODIFICATION WITHOUT THE CONSENT OF THE @author WILL BE DISPUTED IN
*	TERMS WITH THE PROJECT LICENSE.
*/
package main

import (
	"net/http"
	"log"
	"time"
	"encoding/json"

	serverish "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Architecture struct {
	Code string `json:"code"`
	User string `json:"user"`
	Room string `json:"room"`
}

type Check struct {
	Room string `json:"room"`
}

type Feedback struct {
	Message string `json:"message"`
	Username string `json:"username"`
	Room string `json:"room"`
	Code int `json:"code"`
}

type Data struct {
	Message  string `json:"message"`
	Username string `json:"username"`
	Code     int    `json:"code"`
}

// @todo
//type UserInfo struct {
//	Name string `json:"name"`
//}

const found string = "\"found\"" // processor architecture of arm (the robot don't strip off the quotes)

func main() {
	server := serverish.NewServer(transport.GetDefaultWebsocketTransport())
	server.On(serverish.OnConnection, func(client *serverish.Channel, args interface{}) {
		log.Println("Client connected, client id is", client.Id())
	})

	server.On("/execution", func(client *serverish.Channel, arch Architecture) {
		for _, robot := range client.List(arch.Room) {
			result, err := robot.Ack("/check", Check{arch.Room}, time.Second)

			if result == found && err == nil {
				robot.Emit("/serve", arch)
				return
			}
		}
	})

	server.On("/feedback", func(client *serverish.Channel, data string) {
		var feedback Feedback

		json.Unmarshal([]byte(data), &feedback)
		server.BroadcastTo(feedback.Room, "/results", Data{feedback.Message, feedback.Username, feedback.Code})
	})

	/**
		==============================================================================
		If an EV3 connects on the socket server it must be redirected to its own room.
		==============================================================================
	 */
	server.On("/joinable", func(client *serverish.Channel, room string) {
		client.Join(room)

		//@todo
		//server.Broadcast(room, "/joined", UserInfo{})
	})

	/**
		=========================================
		Initialize the Mux Server for the socket.
		=========================================
	 */
	mux := http.NewServeMux()
	mux.Handle("/socket.io/", server)

	/**
		===============================
		Log the action with the logger.
		===============================
	 */
	log.Println("Starting socket.io main server at http://localhost:8080/")
	log.Panic(http.ListenAndServe(":8080", mux))
}
