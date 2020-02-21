package network

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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
		TCPconn, err := net.DialTCP("tcp", nil, node.TCPaddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
		if err != nil {
			log.Fatal(err)
		}
		n.Connections[node.ID] = TCPconn
	}
	return err
}

//StartServer start the TCP listener on application host
func (n *Network) StartServer() error {
	TCPln, err := net.ListenTCP("tcp", n.Myself.TCPaddr) //func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	if err != nil {
		log.Fatal(err)
		return err
	}
	n.Myself.TCPListen = TCPln

	go func(TCPln *net.TCPListener) {
		defer TCPln.Close()

		// run loop forever (or until ctrl-c)
		for {
			TCPconn, err := TCPln.AcceptTCP() //func() (*net.TCPConn, error)
			if err != nil {
				log.Fatal(err)
				continue
			}
			//RemoteHost := TCPconn.RemoteAddr() //Do I need this RemoteHost (TCP Scoket)?

			buffer := make([]byte, 1024, 1024) //Channel??
			n, err := TCPconn.Read(buffer[0:])

			recData := string(buffer[0:n])
			fmt.Println(recData)

			//Creates a terminal chat to test TCP connectivity and Docker setup
			reader := bufio.NewReader(os.Stdin) //func(rd io.Reader) *bufio.Reader - NewReader returns a new Reader whose buffer has the default size.
			sendData, _ := reader.ReadString('\n')
			sendByte := []byte(sendData)
			TCPconn.Write(sendByte)

		}
	}(TCPln)

	return err
}

/*
func newServer(server Server) (*net.TCPConn, error) {
	lSocket := server.ip + ":" + strconv.Itoa(server.port)
	lAddr, err := net.ResolveTCPAddr("tcp", lSocket) //func(network string, address string) (*net.TCPAddr, error)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	TCPln, err := net.ListenTCP("tcp", lAddr) //func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	if err != nil {
		log.Fatal(err)
		return TCPln, err
	}
	return TCPln, nil
}

func startServer() {

}

func startTcpServer() error {
	lnTCP, err := net.ListenTCP("tcp", "localhost:5000") //Func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	return error
}

func tcpConnect() error {
	connTCP, err := net.DialTCP("tcp", "localhost:5001") //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(clientConn)
	writer := bufio.NewWriter(clientConn)

	return error
}
*/
