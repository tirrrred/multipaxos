package network

import (
	//"bufio"
	"fmt"
	"log"
	"net"
	//"os"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
)

//NetConfig struct has network information to be used for configuration of network. How you are in the network (node id) and how is every one else in the network
type NetConfig struct {
	Myself int
	Nodes  []Node
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
	Myself      Node
	Nodes       []Node
	Connections map[int]*net.TCPConn
}

//Message struct
type Message struct {
	To      int //Node ID
	From    int // Node ID
	Msg     string
	Request bool // true -> request, false -> reply
}

var mutex = &sync.Mutex{}

//InitNetwork defines necessary parameteres about the network
func InitNetwork(nodes []Node, myself int) (network Network, err error) {
	network = Network{
		Nodes:       []Node{},
		Connections: map[int]*net.TCPConn{},
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

//EstablishNetwork starts server, Init connections to nodes
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
		} else {
			n.Connections[node.ID] = TCPconn
			fmt.Printf("DialTCP to node %v\n", node.TCPaddr)
		}
	}
	fmt.Println(n.Connections)
	return err
}

//StartServer start the TCP listener on application host
func (n *Network) StartServer() (err error) {
	TCPln, err := net.ListenTCP("tcp", n.Myself.TCPaddr) //func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	fmt.Printf("TCP Server for node %v at %v\n", n.Myself.ID, n.Myself.TCPaddr)
	if err != nil {
		log.Print(err)
		return err
	}
	n.Myself.TCPListen = TCPln
	defer TCPln.Close()

	// run loop forever (or until ctrl-c)
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
		for _, node := range n.Nodes {
			if node.IP == RemoteIP {
				mutex.Lock()
				n.Connections[node.ID] = TCPconn
				mutex.Unlock()
				fmt.Printf("AcceptTCP from node %v\n", TCPconn.RemoteAddr())
			}
		}

		//Handle connections
		go n.HandleConns(TCPconn)
	}

	return err
}

//SendToNode sends messages to nodeID
func (n *Network) SendToNode(nodeID int) (err error) {
	msg := "Test Message From Node " + strconv.Itoa(nodeID)
	/*message := Message{
		To:      nodeID,
		From:    n.Myself.ID,
		Msg:     msg,
		Request: true,
	}*/
	msgByte := []byte(msg)
	for i := 0; i < 60; i++ {
		time.Sleep(2 * time.Second)
		n.Connections[nodeID].Write(msgByte)
	}
	return err
}

//ReceiveMessage receives messages from other nodes in network
func (n *Network) ReceiveMessage(nodeID int) (message Message, err error) {
	for {
		buffer := make([]byte, 1024, 1024)
		len, err := n.Connections[nodeID].Read(buffer[0:]) //func(b []byte) (int, error)
		//Prints the recived output for testing
		for _, v := range buffer[0:len] {
			fmt.Println(v)
		}
		message := new(Message)
		err = json.Unmarshal(buffer[0:len], &message)
		return *message, err
	}
}

//HandleConns handle TCP connections
func (n *Network) HandleConns(TCPconn *net.TCPConn) (err error) {
	buffer := make([]byte, 1024, 1024)
	for {
		len, _ := TCPconn.Read(buffer[0:])
		//Prints the recived output for testing
		for _, v := range buffer[0:len] {
			fmt.Println(v)
		}
		clientInputstr := string(buffer[0:len])
		if clientInputstr == "Hei" {
			nodeID := n.findRemoteAddr(TCPconn)
			n.SendMessage(nodeID, "done")
			fmt.Printf("Message from node%d: %v", nodeID, clientInputstr)
		} else if clientInputstr == "done" {
			fmt.Println("Sent hei, and now I recived done")
			break
		}
	}

	return err
}

//SendMessage sends a message
func (n *Network) SendMessage(NodeID int, message string) (err error) {
	n.Connections[NodeID].Write(([]byte(message)))
	return nil
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
