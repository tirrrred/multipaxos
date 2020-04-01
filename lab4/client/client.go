/*
Clients should be able to connect to the Paxos leader and issue
commands. Note that the single-decree version you have implemented for this lab
only decides on a single value. This means that only one client command can be
chosen (and stored) by your system. You will in the next lab extend your
application to handle multiple commands.

Clients should connect to _all_ Paxos nodes and send any command to _all_ of
them. The commands should be of type string, thereby matching the `Value` type
used by the Paxos roles.

The client needs:
1) Overview of proposer nodes
	- This includes TCP Socket (IP:Port) to establish TCPconn
2) Broadcast (should be a network function?)

*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/tirrrred/multipaxos/lab3/network"
	"github.com/tirrrred/multipaxos/lab4/singlepaxos"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	fmt.Println("Hello client app")

	//Create channels
	ReceiveChan := make(chan network.Message, 16)
	SendChan := make(chan network.Message, 16)

	//import network configuration based on netConfig.json file
	netconf, _ := importNetConf()

	//Based on network configuration, make TCP connections to each proposer and store it in a map, where proposer ID is the key and the TCP connection is the value
	conn2props := make(map[int]*net.TCPConn)
	conn2props, _ = connToProposers(netconf.Proposers, netconf.Nodes, conn2props)

	//Listen on each tcp connection (from different proposers)
	for _, tcpConn := range conn2props {
		go listenOnConn(tcpConn, ReceiveChan)
	}

	//Clients send command loop
	go func() {
		for {
			fmt.Println("\nEnter command: ")
			var input string
			fmt.Scanln(&input)
			msg := network.Message{
				Type:  "Value",
				Value: singlepaxos.Value(input),
			}
			SendChan <- msg
		}
	}()

	//Client main loop for handling incoming and outgoing messages
	for {
		select {
		case receiveMsg := <-ReceiveChan:
			if receiveMsg.Type != "Value" {
				continue
			}
			fmt.Printf("\nClient: Received value %v from node %d", receiveMsg.Value, receiveMsg.From)
		case sendMsg := <-SendChan:
			sendMessage(sendMsg, conn2props)
			fmt.Println("\nClient: Sends value: ", sendMsg.Value)
		}
	}
}

func importNetConf() (network.NetConfig, error) {
	//Open the network config file for the application
	netConfFile, err := os.Open("../app/netConf.json")
	if err != nil {
		log.Print(err)
	}
	defer netConfFile.Close()

	//Read the network config file as a byte array
	byteValue, _ := ioutil.ReadAll(netConfFile)

	//Init the netConfig struct to store the network config from file
	var netconf network.NetConfig

	err = json.Unmarshal(byteValue, &netconf)
	if err != nil {
		log.Print(err)
		return netconf, err
	}

	/*
		fmt.Println(string(byteValue))
		fmt.Printf("%v, %T\n", netconf, netconf)
		fmt.Printf("%v, %T\n", netconf.Nodes, netconf.Nodes)
		for _, node := range netconf.Nodes {
			fmt.Printf("Node ID is %v, node IP is '%v' and node serverport is %v\n", node.ID, node.IP, node.Port)
		}
	*/

	return netconf, nil
}

func connToProposers(props []int, nodes []network.Node, propConnMap map[int]*net.TCPConn) (map[int]*net.TCPConn, error) {
	for _, prop := range props {
		for _, node := range nodes {
			if prop == node.ID {
				rAddr, err := net.ResolveTCPAddr("tcp", node.IP+":"+strconv.Itoa(node.Port)) //ResolveTCPAddr func(network, address string) (*TCPAddr, error)
				if err != nil {
					log.Print(err)
					return nil, err
				}
				TCPconn, err := net.DialTCP("tcp", nil, rAddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
				if err != nil {
					log.Print(err)
					return nil, err
				}
				propConnMap[prop] = TCPconn
			}
		}
	}
	return propConnMap, nil
}

func listenOnConn(TCPconn *net.TCPConn, rChan chan network.Message) {
	defer TCPconn.Close()
	buffer := make([]byte, 1024, 1024)
	for {
		len, err := TCPconn.Read(buffer[0:])
		if err != nil {
			if err == io.EOF {
				fmt.Println(string(buffer[:len]))
			}
			fmt.Print("Client: listenOnConn error: ", err)
			fmt.Println("\tClosing TCP connection: ", TCPconn.RemoteAddr())
			TCPconn.Close()
		}
		message := new(network.Message)
		err = json.Unmarshal(buffer[0:len], &message)
		if err != nil {
			log.Print(err)
			fmt.Println(string(buffer[0:len]))
			return
		}
		rChan <- *message
	}
}

//SendMessage sends a message
func sendMessage(message network.Message, propConnMap map[int]*net.TCPConn) error {
	messageByte, err := json.Marshal(message) //func(v interface{}) ([]byte, error)
	if err != nil {
		log.Print(err)
		return err
	}

	for _, conn := range propConnMap {
		bytes, err := conn.Write(messageByte)
		fmt.Println("Client: Sends message of size (bytes): ", bytes)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}
