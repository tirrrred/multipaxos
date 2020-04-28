package clienthandler2

import (
	"dat520/lab3/detector"
	"dat520/lab3/network"
	"dat520/lab5/multipaxos"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

//ClientHandler struct for necessary fields
type ClientHandler struct {
	id                int                  //ID of the node running the clienthandler
	proposer          *multipaxos.Proposer //
	ClientConnsMap    map[string]*net.TCPConn
	ClientValueChanIn chan multipaxos.Value
	LearnValueChan    chan multipaxos.Value
	ClientConnChan    chan *net.TCPConn
	ClientConnsSlice  []*net.TCPConn
	ld                detector.LeaderDetector
	leader            int
}

//NewClientHandler creates a new ClientHandler
func NewClientHandler(id int, proposer *multipaxos.Proposer, ld detector.LeaderDetector, clConnChan chan *net.TCPConn) *ClientHandler {
	return &ClientHandler{
		id:                id,
		proposer:          proposer,
		ClientValueChanIn: make(chan multipaxos.Value, 3000),
		LearnValueChan:    make(chan multipaxos.Value, 3000),
		ld:                ld,
		leader:            ld.Leader(),
		ClientConnChan:    clConnChan,
		ClientConnsSlice:  []*net.TCPConn{},
		ClientConnsMap:    make(map[string]*net.TCPConn),
	}
}

//Start starts the Clienthandler main loop
func (ch *ClientHandler) Start() {
	ldCHG := ch.ld.Subscribe()
	go func() {
		for {
			select {
			case cConn := <-ch.ClientConnChan:
				ch.ClientConnsSlice = append(ch.ClientConnsSlice, cConn)
				ch.GetClientInfo(cConn)
			case cliVal := <-ch.ClientValueChanIn:
				if ch.id != ch.leader { //If this node is not the leader node
					ch.Redirect(cliVal)
				}
				ch.proposer.DeliverClientValue(cliVal)
			case newLeader := <-ldCHG:
				ch.leader = newLeader
			}
		}
	}()
}

//DeliverClientValue send client value to the channel clientValueChanIn so the clienthandler can handle the value
func (ch *ClientHandler) DeliverClientValue(val multipaxos.Value) {
	ch.ClientValueChanIn <- val
}

//DeliverResponse send resplonse based on clienter handler algorithm
//func (ch *ClientHandler) DeliverResponse(res multipaxos.Response) {
//	ch.responseChanIn <- res
//}

//DeliverClientInfo delivers client information from client -> invoked when running ch.GetClientInfo() func
func (ch *ClientHandler) DeliverClientInfo(cliMsg network.Message) {
	fmt.Printf("ClienterHandler %d: Got client info from %v", ch.id, cliMsg.ClientInfo.ClientID)
	clientInfo := cliMsg.ClientInfo
	ch.ClientConnsMap[clientInfo.ClientID] = clientInfo.Conn
}

//GetClientInfo is sent to clients to get ClientID and information to be stored and uses later
func (ch *ClientHandler) GetClientInfo(cConn *net.TCPConn) {
	message := network.Message{
		Type: "Getinfo",
		From: ch.id,
	}
	messageByte, err := json.Marshal(message)
	if err != nil {
		log.Print(err)
	}
	_, err = cConn.Write(messageByte)
	if err != nil {
		log.Print(err)
	}
}

//Redirect : if Clients tryies to send value to a node that's not the leader, it should be redirected to the networks/clusters leader ID
func (ch *ClientHandler) Redirect(val multipaxos.Value) {
	cConn := ch.ClientConnsMap[val.ClientID]

	redirMsg := network.Message{
		Type:         "Redirect",
		From:         ch.id,
		RedirectNode: ch.leader,
	}
	messageByte, err := json.Marshal(redirMsg)
	if err != nil {
		log.Print("ClientHandler - Redirect: json.Marshal: ", err)
	}
	fmt.Printf("cConn: %v, RemoteAddr: %v", cConn, cConn.RemoteAddr().String)
	_, err = cConn.Write(messageByte)
	if err != nil {
		log.Print("ClientHandler - Redirect: cConn.Write: ", err)
	}
}

//SendValToCli sends decidedValueToClient
func (ch *ClientHandler) SendValToCli(dVal multipaxos.DecidedValue) {
	return
}
