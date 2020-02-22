package network

import (
	//"bufio"
	"fmt"
	"log"
	"net"
	//"os"
	"strconv"
	"strings"
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
}

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

//InitConns start TCP server on application and initiation TCP connection to other nodes in network
func (n *Network) InitConns() (err error) {
	//Connect to each node in network
	for _, node := range n.Nodes {
		fmt.Printf("DialTCP to node %v\n", node)
		TCPconn, err := net.DialTCP("tcp", nil, node.TCPaddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
		if err != nil {
			log.Print(err)
		}
		n.Connections[node.ID] = TCPconn
		fmt.Printf("Connection to node %v: %v\n", node, n.Connections[node.ID])
	}
	fmt.Println(n.Connections)
	return err
}

//StartServer start the TCP listener on application host
func (n *Network) StartServer() (err error) {
	TCPln, err := net.ListenTCP("tcp", n.Myself.TCPaddr) //func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	fmt.Printf("Server listener for node %v: %v\n", n.Myself.TCPaddr, TCPln)
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
				n.Connections[node.ID] = TCPconn
			}
		}
		fmt.Println(n.Connections)
	}
	return err
}
