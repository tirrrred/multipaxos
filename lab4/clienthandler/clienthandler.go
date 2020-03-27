package clienthandler

/*
The client handling module should listen for and handle client connections.

The client handling module should forward any command
to the proposer using the `DeliverClientCommand` method. The Proposer should in
turn propose the command if it consider itself the current leader.  Your
application should listen for decided values from the learner using the
`valueOut` channel. Any decided value should be given to the client handling
module, which in turn should forward it to _every_ connected client.

*/

import (
	"encoding/json"
	"fmt"
	"github.com/tirrrred/multipaxos/lab3/network"
	"github.com/tirrrred/multipaxos/lab4/singlepaxos"
	"log"
	"net"
)

//ClientHandler struct for neccessary fields
type ClientHandler struct {
	ID              int
	proposer        *singlepaxos.Proposer
	ClientValueChan chan singlepaxos.Value
	LearnValueChan  chan singlepaxos.Value
	ClientConns     []*net.TCPConn
	ClientConnChan  chan *net.TCPConn
}

//NewClientHandler Init the ClientHandler.
//ID is the ID of the Node running the clienthandler,
//proposer is the proposer on the node,
//clConnChan a channel from the "network package" which send TCP connections from clients (i.e not node connections)
func NewClientHandler(id int, proposer *singlepaxos.Proposer, clConnChan chan *net.TCPConn) *ClientHandler {
	return &ClientHandler{
		ID:              id,
		proposer:        proposer,
		ClientValueChan: make(chan singlepaxos.Value),
		LearnValueChan:  make(chan singlepaxos.Value),
		ClientConns:     []*net.TCPConn{},
		ClientConnChan:  clConnChan,
	}
}

//Start starts the clienthandler main loop for handling client connections and messages
func (ch *ClientHandler) Start() {
	go func() {
		for {
			select {
			case cConn := <-ch.ClientConnChan:
				ch.ClientConns = append(ch.ClientConns, cConn)
				fmt.Println("ClientHandler: Got a new client connection!")
			case clientVal := <-ch.ClientValueChan:
				ch.proposer.DeliverClientValue(clientVal)
			case lrnVal := <-ch.LearnValueChan:
				lrnMsg := network.Message{
					Type:  "Value",
					From:  ch.ID,
					Value: lrnVal,
				}
				for _, cConn := range ch.ClientConns {
					//Send learned value to client conns
					messageByte, err := json.Marshal(lrnMsg) //func(v interface{}) ([]byte, error)
					if err != nil {
						log.Print(err)
					}
					_, err = cConn.Write(messageByte)
					if err != nil {
						log.Print(err)
					}
				}
			}
		}
	}()
}

//DeliverValue sends message to the proposer
func (ch *ClientHandler) DeliverValue(val singlepaxos.Value) {
	ch.ClientValueChan <- val
}

//SendValToCli sends value to all the clients known
func (ch *ClientHandler) SendValToCli(val singlepaxos.Value) {
	ch.LearnValueChan <- val
}
