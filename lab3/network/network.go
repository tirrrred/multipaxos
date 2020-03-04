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
	//"time"
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
	ReceiveChan chan Message
	SendChan    chan Message
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
	rC := make(chan Message, 16)
	sC := make(chan Message, 16)
	network = Network{
		Nodes:       []Node{},
		Connections: map[int]*net.TCPConn{},
		ReceiveChan: rC,
		SendChan:    sC,
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

	go func() { //Fucntion to listen for TCP connections
		defer TCPln.Close()
		for {
			//Accept TCP connections to application server
			TCPconn, err := TCPln.AcceptTCP() //func() (*net.TCPConn, error)
			if err != nil {
				log.Print(err) //This fails: accept tcp 10.0.0.102:5002: use of closed network connection
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
			go n.ListenConns(TCPconn)
		}
	}()
	go func() { //Function to listen on sendChan for messages to send to other network nodes
		for {
			message := <-n.SendChan
			err := n.SendMessage(message)
			if err != nil {
				fmt.Println("Error: Sending message to network node failed")
			}
		}
	}()
	return err
}

//ListenConns handle TCP connections
func (n *Network) ListenConns(TCPconn *net.TCPConn) (err error) {
	buffer := make([]byte, 1024, 1024)
	nodeID := n.findRemoteAddr(TCPconn)
	fmt.Println("At HandleConns func now. Connection from: ", nodeID)
	for {
		len, _ := TCPconn.Read(buffer[0:])
		message := new(Message)
		err = json.Unmarshal(buffer[0:len], &message)
		n.ReceiveChan <- *message
	}

	/*for {
		len, _ := TCPconn.Read(buffer[0:])
		clientInputstr := string(buffer[0:len])
		//fmt.Printf("Message from node%d: %v of type %T", nodeID, clientInputstr, clientInputstr)
		if clientInputstr == "Hei\n" {
			n.SendMessage(nodeID, "done")
			fmt.Printf("Message from node%d: %v", nodeID, clientInputstr)
		} else if clientInputstr == "done" {
			fmt.Println("Sent hei, and now I recived done. BREAK!")
			break
		}
		err = n.SendMessage(n.Myself.ID+1, "Hei")
		if err != nil {
			fmt.Printf("Error sending from node%d to node%d", n.Myself.ID, (n.Myself.ID + 1))
			return err
		}
	}*/
	return err
}

//SendMessage sends a message
func (n *Network) SendMessage(message Message) (err error) {
	messageByte, err := json.Marshal(message) //func(v interface{}) ([]byte, error)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = n.Connections[message.To].Write(messageByte)
	if err != nil {
		log.Print(err)
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
