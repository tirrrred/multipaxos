package main

import (
	"bufio"
	"dat520/lab3/network"
	"dat520/lab5/bank"
	"dat520/lab5/multipaxos"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	myID           string
	seqNum         = 0                              //Clients Sequence Number for client commands
	reqSeq         = 0                              //Request Sequence Number: Stores which request is currently sent, and waiting for response for (sync communication)
	seqCmd         = make(map[int]multipaxos.Value) //Maps sequence numbers (seqNum) with commands(values)
	responseBuffer = make(map[int]network.Message)
	currentConn    = -1                         //Which node are we connected to now
	connTable      = make(map[int]*net.TCPConn) //Map with TCPconnection to different nodes
	networkNodes   = []network.Node{}
	ReceiveChan    = make(chan network.Message, 3000) //Create channels
	SendChan       = make(chan network.Message, 3000) //Create channels
	responseOK     = true
	responseTimer  *time.Ticker
	mu             sync.Mutex
)

func main() {
	//import network configuration based on netConfig.json file
	netconf, _ := importNetConf()
	connectToNodes(netconf.Nodes)

	myID := generateID(10)

	responseTimer = time.NewTicker(10 * time.Second)
	responseTimer.Stop()
	//Clients send command loop - Manual mode
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Println("\nEnter command ('Operation' + 'Amount' + 'Account'): ")
			fmt.Println("Valid operations: 'balance', 'deposit', 'withdraw")
			scanner.Scan()
			input := scanner.Text()
			val, err := inputToValue(input, myID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			//fmt.Println(val)
			//fmt.Println("for loop (after inputToValue) seqNum: ", seqNum)
			seqCmd[seqNum] = val //Add Command/Value to sequence map
			//fmt.Println(seqCmd)
			//status := syncTxRx(msg) //A client should always send requests synchronously, i.e. wait for a response to the previous request before sending a new one
			//fmt.Println("networkNodes[]: ", networkNodes)
			//fmt.Println("connTable map[int]TCPconn: ", connTable)
			if responseOK {
				syncTxRx(seqCmd[reqSeq+1], "Value")
			}
		}
	}()

	for {
		select {
		case rMsg := <-ReceiveChan:
			switch rMsg.Type {
			case "Redirect":
				fmt.Println("Client: Redirected to node " + strconv.Itoa(rMsg.RedirectNode))
				reconnect(rMsg, seqCmd[rMsg.Value.ClientSeq], false)
			case "Getinfo":
				deliverClientInfo(rMsg, myID, connTable[rMsg.From])
			case "Response":
				responseTimer.Stop()
				//fmt.Printf("Got value form node %d: ClientSeq = %d and reqSeq = %d\n", rMsg.From, rMsg.Value.ClientSeq, reqSeq)
				if rMsg.Response.ClientSeq == reqSeq {
					mu.Lock()
					responseOK = true
					mu.Unlock()
					fmt.Printf("Client: Got response from %d: \nAccount %d = %d\n", rMsg.From, rMsg.Response.TxnRes.AccountNum, rMsg.Response.TxnRes.Balance)
				} else {
					responseBuffer[rMsg.Value.ClientSeq] = rMsg
				}
				//handleIncValue(rMsg)
			}
		case sMsg := <-SendChan:
			switch sMsg.Type {
			case "Value":
				fmt.Printf("Client: Sending transaction to node %d: \nAccount: %d | %v: %d\n", sMsg.To, sMsg.Value.AccountNum, sMsg.Value.Txn.Op, sMsg.Value.Txn.Amount)
				err := sendMessage(currentConn, sMsg)
				if err != nil {
					log.Print(err)
				}
			}
		case <-responseTimer.C:
			fmt.Printf("Client command request timed out...\n Reconnecting to another node\n")
			reVal := seqCmd[reqSeq]
			reMsg := network.Message{
				To:    currentConn,
				Type:  "Timeout",
				Value: reVal,
			}
			reconnect(reMsg, reVal, true)
		}
	}
}

func generateID(n int) string {
	t := time.Now().UTC()
	timestamp := t.Format(time.RFC3339)

	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	entropy := string(b)
	id := timestamp + entropy
	return id
}

func inputToValue(input string, id string) (val multipaxos.Value, err error) {
	mu.Lock()
	defer mu.Unlock()
	commands := strings.Fields(input)
	//fmt.Println(commands, len(commands))

	if len(commands) == 2 { //If numbers of commands equals two the operation should be balance followed by account number (i.e no amount)
		accountStr := commands[1]
		accountNum, _ := strconv.Atoi(accountStr)
		if commands[0] == "balance" {
			seqNum++
			//fmt.Println("inputToValue seqNum: ", seqNum)
			return multipaxos.Value{
				ClientID:   id,
				ClientSeq:  seqNum,
				Noop:       false,
				AccountNum: accountNum,
				Txn: bank.Transaction{
					Op:     bank.Operation(0),
					Amount: 0,
				},
			}, nil
		}
		return val, fmt.Errorf("you fucked up, enter a valid command (op + amount + account)")
	}
	if len(commands) == 3 {
		amountStr := commands[1]
		amountNum, _ := strconv.Atoi(amountStr)
		accountStr := commands[2]
		accountNum, _ := strconv.Atoi(accountStr)
		if commands[0] == "deposit" || commands[0] == "withdraw" {
			seqNum++
			//fmt.Println("inputToValue seqNum: ", seqNum)
			op := 1 //default is deposit = 1
			if commands[0] == "withdraw" {
				op = 2
			}
			return multipaxos.Value{
				ClientID:   id,
				ClientSeq:  seqNum,
				Noop:       false,
				AccountNum: accountNum,
				Txn: bank.Transaction{
					Op:     bank.Operation(op),
					Amount: amountNum,
				},
			}, nil
		} else {
			return val, fmt.Errorf("you fucked up, enter a valid command (op + amount + account)")
		}
	} else {
		return val, fmt.Errorf("you fucked up, enter a valid command (op + amount + account)")
	}
}

func syncTxRx(val multipaxos.Value, msgType string) {
	//Send message (request) to correct connection (currentConn)
	//Wait for reply (reponse) from connection
	// 1) If correct -> proceed with next command
	// 2) If fail - timeout -> reconnect(currentConn+1)
	// 3) If fail - Redirect -> reconnect(addr)
	// 4) If fail - CloseConn -> Reconnect(currentConn+1)
	// 5) If fail - err -> Reconnect(currentConn+1)
	mu.Lock()
	defer mu.Unlock()
	responseOK = false
	reqSeq = val.ClientSeq
	//fmt.Println("syncTxRx  reSeq = seqNum: ", reqSeq, seqNum)
	msg := network.Message{
		To:    currentConn,
		Type:  msgType,
		Value: val,
	}
	SendChan <- msg
	responseTimer = time.NewTicker(10 * time.Second)
	//responseTimer := time.NewTicker(5 * time.Second)
	//msgSent = true
}

//connectToNodes establish TCP Connections to all nodes from the netConf.json file and stores them in connTable map
func connectToNodes(nodes []network.Node) {
	for _, node := range nodes {
		go func(n network.Node) {
			rAddr, err := net.ResolveTCPAddr("tcp", n.IP+":"+strconv.Itoa(n.Port)) //ResolveTCPAddr func(network, address string) (*TCPAddr, error))
			if err != nil {
				log.Print(err)
				//continue
			}
			TCPconn, err := net.DialTCP("tcp", nil, rAddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
			if err != nil {
				log.Print(err)
				//continue
			}
			connTable[n.ID] = TCPconn
			listenOnConn(TCPconn, ReceiveChan)
		}(node)
		networkNodes = append(networkNodes, node)
	}

	currentConn = nodes[0].ID //Starts to use node 0 as a starting point for sending values
}

func reconnect(rMsg network.Message, val multipaxos.Value, timeout bool) {
	//fmt.Printf(" Received value seq: \t%d\n Expected value seq: \t%d\n Current Conn: \t\t%d\n New Conn: \t\t%d\n", rMsg.Value.ClientSeq, reqSeq, currentConn, rMsg.RedirectNode)
	nodeID := rMsg.RedirectNode
	cSeq := rMsg.Value.ClientSeq
	if timeout == false {
		//Check if we got a active connection for given nodeID
		if _, ok := connTable[nodeID]; ok {
			currentConn = nodeID
			//fmt.Println("Current Connection is to Node: ", currentConn)
		} else {
			fmt.Println("Don't have any active TCP connection for given NodeID, cheking netConf.json file again to verify")
			for _, node := range networkNodes {
				if node.ID == nodeID {
					fmt.Println("Found a corresponding network Node in netConf.json. Trying to connect...")
					rAddr, err := net.ResolveTCPAddr("tcp", node.IP+":"+strconv.Itoa(node.Port)) //ResolveTCPAddr func(network, address string) (*TCPAddr, error))
					if err != nil {
						log.Print(err)
					}
					TCPconn, err := net.DialTCP("tcp", nil, rAddr) //func(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
					if err != nil {
						log.Print(err)
						fmt.Println("Unable to connect to network node...")
						continue
					}
					connTable[node.ID] = TCPconn
					fmt.Println(connTable)
					currentConn = node.ID
				}
			}
		}
	} else if timeout {
		testID := currentConn - 1
		if testID < 0 {
			currentConn = len(networkNodes) - 1
		} else {
			currentConn = testID
		}
		if cSeq == reqSeq {
			fmt.Printf("Timeout, reconnecting to new node %d\n", currentConn)
			syncTxRx(val, "Timeout")
			return
		}
	}

	//fmt.Println("Failed to find or connect to a network node with the given nodeID: ", nodeID)
	if cSeq == reqSeq {
		fmt.Printf("Reconnected to new node %d - Resending message\n", currentConn)
		syncTxRx(val, "Value")
		return
	}

}

func sendMessage(nodeID int, message network.Message) error {
	messageByte, err := json.Marshal(message) //func(v interface{}) ([]byte, error)
	if err != nil {
		log.Print(err)
		return err
	}
	if conn, ok := connTable[nodeID]; ok {
		bytes, err := conn.Write(messageByte)
		fmt.Printf("Client: Sends message to node %d with seqNum %d of size (bytes): %v \n", nodeID, message.Value.ClientSeq, bytes)
		if err != nil {
			log.Print(err)
			return err
		}
	} else {
		return fmt.Errorf("No active TCP conn for given nodeID: %d", nodeID)
	}
	return nil
}

func importNetConf() (network.NetConfig, error) {
	//Open the network config file for the application
	netConfFile, err := os.Open("../netConf.json")
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

	return netconf, nil
}

func deliverClientInfo(msg network.Message, id string, conn *net.TCPConn) {
	cliInfo := network.ClientInfo{
		ClientID: id,
		Addr:     conn.LocalAddr().String(),
	}

	cliMsg := network.Message{
		Type:       "ClientInfo",
		To:         msg.From,
		ClientInfo: cliInfo,
	}
	//fmt.Println("Sending client Info ", cliMsg)
	err := sendMessage(msg.From, cliMsg)
	if err != nil {
		log.Print(err)
	}
}

func listenOnConn(TCPconn *net.TCPConn, rChan chan network.Message) {
	defer TCPconn.Close()
	buffer := make([]byte, 10240, 10240)
	for {
		len, err := TCPconn.Read(buffer[0:])
		if err != nil {
			if err == io.EOF {
				fmt.Println(string(buffer[:len]))
			}
			fmt.Print("Client: listenOnConn error: ", err)
			fmt.Println("\tClosing TCP connection: ", TCPconn.RemoteAddr())
			TCPconn.Close()
			break
		}
		message := new(network.Message)
		err = json.Unmarshal(buffer[0:len], &message)
		if err != nil {
			log.Print(err)
			fmt.Println(string(buffer[0:len]))
			continue
		}
		rChan <- *message
	}
}

func handleIncValue(rMsg network.Message) {

}
