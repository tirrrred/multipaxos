package network

import (
	//"bufio"
	"dat520/lab5/multipaxos"
	"fmt"
	"log"
	"net"

	//"os"

	"encoding/json"
	"strconv"
	"strings"
	"sync"
	//"time"
)

//NetConfig struct has network information to be used for configuration of network. How you are in the network (node id) and how is every one else in the network
type NetConfig struct {
	Myself    int
	Nodes     []Node
	Proposers []int
	Acceptors []int
	Learners  []int
}

//Node struct has node spesific fields like node id, node IP and node server port
type Node struct {
	ID        int
	IP        string
	Port      int
	TCPaddr   *net.TCPAddr
	TCPListen *net.TCPListener //Only applicable for myself (application host)
}

//Network struct with spesific network details like network nodes and TCP connections to network nodes
type Network struct {
	Myself         Node
	Nodes          []Node
	Connections    map[int]*net.TCPConn
	ClientConns    []*net.TCPConn
	ReceiveChan    chan Message
	SendChan       chan Message
	ClientConnChan chan *net.TCPConn
}

//Message struct
type Message struct {
	Type         string     //Type of Message: Heartbeat, Prepare, Promise, Accept, Learn, Value, Redirect, Getinfo
	To           int        //Node ID
	From         int        //Node ID
	Msg          string     //Client message
	RedirectNode int        //Node ID to redirect connection to (if type == 'Redirect'). Used between clienthandler and client
	Request      bool       //true -> request, false -> reply
	ClientInfo   ClientInfo //Used to get clientinfo (like clientID) -> If type == 'Getinfo'
	Promise      multipaxos.Promise
	Accept       multipaxos.Accept
	Prepare      multipaxos.Prepare
	Learn        multipaxos.Learn
	Value        multipaxos.Value
	Response     multipaxos.Response
}

//ClientInfo is used to get ClientID from given TCPconn
type ClientInfo struct {
	ClientID string
	Conn     *net.TCPConn
	Addr     string
}

var mutex = &sync.Mutex{}

//InitNetwork defines necessary parameteres about the network
func InitNetwork(nodes []Node, myself int) (network Network, err error) {
	rC := make(chan Message, 65535) //Increased for TS
	sC := make(chan Message, 65535) //Increased for TS
	ccC := make(chan *net.TCPConn, 65535)
	network = Network{
		Nodes:          []Node{},
		Connections:    map[int]*net.TCPConn{},
		ClientConns:    []*net.TCPConn{},
		ReceiveChan:    rC,
		SendChan:       sC,
		ClientConnChan: ccC,
	}

	for _, node := range nodes {
		if node.ID == myself {
			fmt.Printf("You are number %v\n", myself)
			network.Myself = node
			TCPsocket := network.Myself.IP + ":" + strconv.Itoa(network.Myself.Port)
			network.Myself.TCPaddr, err = net.ResolveTCPAddr("tcp", TCPsocket)
		} else {
			TCPsocket := node.IP + ":" + strconv.Itoa(node.Port)
			addr, _ := net.ResolveTCPAddr("tcp", TCPsocket)
			network.Nodes = append(network.Nodes, Node{
				ID:      node.ID,
				IP:      node.IP,
				Port:    node.Port,
				TCPaddr: addr,
			})
		}
	}
	return network, err
}

//EstablishNetwork starts server, Init connections to nodes // NOT USED - REMOVE!
func (n *Network) EstablishNetwork() (err error) {
	err = n.InitConns()
	if err != nil {
		log.Print(err)
		return err
	}
	err = n.StartServer()
	if err != nil {
		log.Print(err)
		return err
	}
	return err
}

//InitConns start TCP server on application and initiation TCP connection to other nodes in network
func (n *Network) InitConns() (err error) {
	//Connect to each node in network
	for _, node := range n.Nodes {
		TCPconn, err := net.DialTCP("tcp", nil, node.TCPaddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
		if err != nil {
			log.Print(err)
			continue
		} else {
			n.Connections[node.ID] = TCPconn
			fmt.Printf("DialTCP to node %v\n", node.TCPaddr)
		}
		//Handle connections
		go n.ListenConns(TCPconn)
	}
	return err
}

//StartServer start the TCP listener on application host
func (n *Network) StartServer() (err error) {
	TCPln, err := net.ListenTCP("tcp", n.Myself.TCPaddr) //func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	fmt.Printf("TCP Server for node %v at %v\n", n.Myself.ID, n.Myself.TCPaddr)
	if err != nil {
		//log.Print(err)
		return err
	}
	n.Myself.TCPListen = TCPln

	go func() { //Fucntion to listen for TCP connections
		defer TCPln.Close()
		for {
			//Accept TCP connections to application server
			TCPconn, err := TCPln.AcceptTCP() //func() (*net.TCPConn, error)
			if err != nil {
				log.Print(err)
			}
			//Find node ID from the remote connection
			RemoteSocket := TCPconn.RemoteAddr()
			RemoteIPPort := strings.Split(RemoteSocket.String(), ":")
			RemoteIP := RemoteIPPort[0]
			//Add remote connection to n.Connection with node.ID as key and TCPconn as value
			clientNode := true
			//loop to match Connection IP woth Node IP in network. Adds the TCP connection to the correct Node ID in map "Connections"
			for _, node := range n.Nodes {
				if node.IP == RemoteIP { //This isn't bulletproof, was there can be several host behind one IP (e.g NATing). Can use Socket Address instead (IP:Port)?
					mutex.Lock()
					n.Connections[node.ID] = TCPconn
					mutex.Unlock()
					fmt.Printf("AcceptTCP from node %v\n", TCPconn.RemoteAddr())
					//Found matching IP address - This TCP connection is from a network node. Sets the clientNode to false
					clientNode = false
				}
			}
			if clientNode {
				n.ClientConns = append(n.ClientConns, TCPconn)
				n.ClientConnChan <- TCPconn
			}
			//Handle connections
			go n.ListenConns(TCPconn)
		}
	}()
	go func() { //Function to listen on sendChan for messages to send to other network nodes
		for {
			message := <-n.SendChan
			//fmt.Println(message)
			err := n.SendMessage(message)
			if err != nil {
				fmt.Println(err) //Spams error message when connection not established - how to fix?
			}
		}
	}()
	return err
}

//ListenConns handle TCP connections
func (n *Network) ListenConns(TCPconn *net.TCPConn) (err error) {
	defer n.CloseConn(TCPconn)
	buffer := make([]byte, 1024, 1024)
	//nodeID := n.findRemoteAddr(TCPconn)
	//fmt.Println("At HandleConns func now. Connection from: ", nodeID)
	n.printConnTable()
	for {
		len, _ := TCPconn.Read(buffer[0:])
		message := new(Message)
		err = json.Unmarshal(buffer[0:len], &message)
		//fmt.Printf("\nReceived message over conn %v: %v\n", TCPconn, *message)
		if message.Type != "Heartbeat" {
			//fmt.Printf("Network: Received message from %d with seqNum %d of size (bytes): %v\n", message.From, message.Value.ClientSeq, len)
		}
		n.ReceiveChan <- *message
	}
	return err
}

var syncMutex = &sync.Mutex{}

//SendMessage sends a message
func (n *Network) SendMessage(message Message) (err error) {
	//syncMutex.Lock()
	//defer syncMutex.Unlock()

	if message.To == n.Myself.ID {
		n.ReceiveChan <- message
		return nil
	}

	messageByte, err := json.Marshal(message) //func(v interface{}) ([]byte, error)
	if err != nil {
		log.Print(err)
		return err
	}
	remoteConn := n.Connections[message.To]
	if remoteConn == nil {
		//fmt.Printf("Connection to node %d isnt present in n.Connections\n", message.To)
		return fmt.Errorf("No TCP connection to node %d in Connection Table", message.To)
	}
	if message.Type != "Heartbeat" { //Remove 1 to block printing of HB messages
		//fmt.Printf("SendMessage: From: %v (%d) To: %v (%d) Message: %v\n", remoteConn.LocalAddr(), n.Myself.ID, remoteConn.RemoteAddr(), message.To, message)
	}
	_, err = n.Connections[message.To].Write(messageByte)
	if err != nil {
		log.Print(err)
		//Assumes the connection is broken (write: broken pipe)
		n.CloseConn(n.Connections[message.To])
		return err
	}
	return err
}

func (n *Network) findRemoteAddr(TCPconn *net.TCPConn) int {
	RemoteSocket := TCPconn.RemoteAddr()
	RemoteIPPort := strings.Split(RemoteSocket.String(), ":")
	RemoteIP := RemoteIPPort[0]

	for _, node := range n.Nodes {
		if node.IP == RemoteIP {
			return node.ID
		}
	}
	return -1
}

//SendMsgTo message to a set of destination hosts, based on node ID
func (n *Network) SendMsgTo(msg Message, dst []int) {
	for _, host := range dst {
		msg.To = host
		err := n.SendMessage(msg)
		if err != nil {
			log.Print(err) //Spams error message when connection not established - how to fix?
			continue
		}
	}
}

//printConnTable is intended as netstat is in windows - prints all connections
func (n *Network) printConnTable() {
	fmt.Printf("**Connection table for node: %d**\n\n", n.Myself.ID)
	fmt.Printf("Node ID \t Local Address \t\t Remote Address\n")
	for nodeID, TCPconn := range n.Connections {
		fmt.Printf("Node %d   \t %v \t %v\n", nodeID, TCPconn.LocalAddr(), TCPconn.RemoteAddr())
	}
	for i, TCPconn := range n.ClientConns {
		fmt.Printf("Client %d \t %v \t %v\n", i, TCPconn.LocalAddr(), TCPconn.RemoteAddr())
	}
	fmt.Printf("\n**Connection table for node: %d**\n", n.Myself.ID)

	//map[int]*net.TCPConn
}

//CloseConn closes *net.TCPConn connections and remove them from connection table
func (n *Network) CloseConn(c *net.TCPConn) error {
	//Node connection or client connection?
	c.Close()
	fmt.Println("Network: Closing connection from", c.RemoteAddr())
	nodeID := n.findRemoteAddr(c)
	if nodeID == -1 { //Client connection
		for i, cC := range n.ClientConns {
			if cC.RemoteAddr() == c.RemoteAddr() {
				n.ClientConns = append(n.ClientConns[:i], n.ClientConns[i+1:]...)
			}
		}
	}
	mutex.Lock()
	delete(n.Connections, nodeID)
	mutex.Unlock()
	n.printConnTable()
	return nil
}
