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
	id                  int                  //ID of the node running the clienthandler
	proposer            *multipaxos.Proposer //
	ClientConnsMap      map[string]*net.TCPConn
	ClientValueChanIn   chan multipaxos.Value
	ClientTimeoutChanIn chan multipaxos.Value
	LearnValueChan      chan multipaxos.Value
	responseChanIn      chan multipaxos.Response
	ClientConnChan      chan *net.TCPConn
	ClientConnsSlice    []*net.TCPConn
	ld                  detector.LeaderDetector
	leader              int
	ClientInfoMap       map[string]network.ClientInfo
}

//NewClientHandler creates a new ClientHandler
func NewClientHandler(id int, proposer *multipaxos.Proposer, ld detector.LeaderDetector, clConnChan chan *net.TCPConn) *ClientHandler {
	return &ClientHandler{
		id:                  id,
		proposer:            proposer,
		ClientValueChanIn:   make(chan multipaxos.Value, 3000),
		ClientTimeoutChanIn: make(chan multipaxos.Value, 3000),
		LearnValueChan:      make(chan multipaxos.Value, 3000),
		responseChanIn:      make(chan multipaxos.Response, 3000),
		ld:                  ld,
		leader:              ld.Leader(),
		ClientConnChan:      clConnChan,
		ClientConnsSlice:    []*net.TCPConn{},
		ClientConnsMap:      make(map[string]*net.TCPConn),
		ClientInfoMap:       make(map[string]network.ClientInfo),
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
				ch.ClientConnsMap[cConn.RemoteAddr().String()] = cConn
				ch.GetClientInfo(cConn)
			case cliVal := <-ch.ClientValueChanIn:
				fmt.Printf("Got clientValue with ClientSeq %d\n", cliVal.ClientSeq)
				if ch.id != ch.leader { //If this node is not the leader node
					ch.Redirect(cliVal)
				}
				ch.proposer.DeliverClientValue(cliVal)
			case cliTimeVal := <-ch.ClientTimeoutChanIn:
				ch.proposer.DeliverClientValue(cliTimeVal)
			case response := <-ch.responseChanIn:
				ch.SendResToCli(response)
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
func (ch *ClientHandler) DeliverResponse(res multipaxos.Response) {
	ch.responseChanIn <- res
}

//DeliverTimeoutMsg delivers a clients reconnect attempt due to timeout to correct handler
func (ch *ClientHandler) DeliverTimeoutMsg(timeVal multipaxos.Value) {
	ch.ClientTimeoutChanIn <- timeVal
}

//DeliverClientInfo delivers client information from client -> invoked when running ch.GetClientInfo() func
func (ch *ClientHandler) DeliverClientInfo(cliMsg network.Message) {
	//fmt.Printf("ClienterHandler %d: Got client info: ID = %v\n", ch.id, cliMsg.ClientInfo.ClientID)
	clientInfo := cliMsg.ClientInfo
	if cConn, ok := ch.ClientConnsMap[clientInfo.Addr]; ok {
		ch.ClientInfoMap[clientInfo.ClientID] = network.ClientInfo{
			ClientID: clientInfo.ClientID,
			Addr:     clientInfo.Addr,
			Conn:     cConn,
		}
	} else {
		fmt.Println("ClienterHandler: Err - Don't have any connection stored for that socket addr!")
	}
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

	cConn := ch.ClientInfoMap[val.ClientID].Conn

	redirMsg := network.Message{
		Type:         "Redirect",
		From:         ch.id,
		RedirectNode: ch.leader,
		Value:        val,
	}
	fmt.Println("ClientHandler: Redirect client. ClientSeq = ", val.ClientSeq)
	messageByte, err := json.Marshal(redirMsg)
	if err != nil {
		log.Print("ClientHandler - Redirect: json.Marshal: ", err)
	}
	//fmt.Printf("cConn: %v, RemoteAddr: %s", cConn, cConn.RemoteAddr().String)
	_, err = cConn.Write(messageByte)
	if err != nil {
		log.Print("ClientHandler - Redirect: cConn.Write: ", err)
	}
}

//SendValToCli sends decidedValueToClient
func (ch *ClientHandler) SendResToCli(res multipaxos.Response) {

	if cliInfo, ok := ch.ClientInfoMap[res.ClientID]; ok {
		cConn := cliInfo.Conn
		resMsg := network.Message{
			Type:         "Response",
			From:         ch.id,
			RedirectNode: ch.leader,
			Response:     res,
		}
		//fmt.Println("ClientHandler: Redirect client. ClientSeq = ", val.ClientSeq)
		messageByte, err := json.Marshal(resMsg)
		if err != nil {
			log.Print("ClientHandler - Redirect: json.Marshal: ", err)
		}
		//fmt.Printf("cConn: %v, RemoteAddr: %s", cConn, cConn.RemoteAddr().String)
		fmt.Printf("Sending Response to client: From: %d, To: %s, ClientSeq: %d", ch.id, cliInfo.Addr, res.ClientSeq)
		_, err = cConn.Write(messageByte)
		if err != nil {
			log.Print("ClientHandler - Redirect: cConn.Write: ", err)
		}
	}
}
