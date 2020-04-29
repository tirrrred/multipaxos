package main

import (
	//"dat520/lab3/detector"
	"dat520/lab3/detector"
	"dat520/lab3/network"

	"dat520/lab5/clienthandler2"
	"dat520/lab5/multipaxos"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	netconf, _ := importNetConf()
	appnet, err := network.InitNetwork(netconf.Nodes, netconf.Myself)
	if err != nil {
		log.Print(err)
	}
	//Creates a slice of node IDs - TO be used in newMonLeaderDetector
	nodeIDs := []int{appnet.Myself.ID}
	for _, node := range appnet.Nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}
	ld := detector.NewMonLeaderDetector(netconf.Proposers) //*MonLeaderDetector
	//create failure detector
	hbSend := make(chan detector.Heartbeat, 3000)
	fd := detector.NewEvtFailureDetector(appnet.Myself.ID, nodeIDs, ld, 2*time.Second, hbSend) //2 second timeout for failure detector - Increase/Decrease?
	//subscribe to leader changes
	ldChan := ld.Subscribe()
	fmt.Printf("\nLeader after ld init and subsribtion: %d\n", ld.Leader())
	fmt.Printf("Proposers: %v, ld: %v, ldChan: %v\n", netconf.Proposers, ld, ldChan)

	//Declare PAXOS variables:
	var proposer *multipaxos.Proposer
	var prepareOutChan chan multipaxos.Prepare
	var acceptOutChan chan multipaxos.Accept
	var acceptor *multipaxos.Acceptor
	var promiseOutChan chan multipaxos.Promise
	var learnOutChan chan multipaxos.Learn
	var learner *multipaxos.Learner
	var decidedOutChan chan multipaxos.DecidedValue
	//var valueOutChan chan multipaxos.Value

	//Implement PAXOS roles:
	if iAMa(netconf.Proposers, appnet.Myself.ID) { //If I'am a proposer
		fmt.Println("I am a PROPOSER!")
		prepareOutChan = make(chan multipaxos.Prepare, 3000)
		acceptOutChan = make(chan multipaxos.Accept, 3000)
		proposer = multipaxos.NewProposer(appnet.Myself.ID, len(netconf.Acceptors), -1, ld, prepareOutChan, acceptOutChan) //NewProposer func(id int, nrOfNodes int, ld detector.LeaderDetector, prepareOut chan<- Prepare, acceptOut chan<- Accept) *Proposer
		proposer.Start()
	}
	if iAMa(netconf.Acceptors, appnet.Myself.ID) { //If I'am a acceptor
		fmt.Println("I am a ACCEPTOR!")
		promiseOutChan = make(chan multipaxos.Promise, 3000)
		learnOutChan = make(chan multipaxos.Learn, 3000)
		acceptor = multipaxos.NewAcceptor(appnet.Myself.ID, promiseOutChan, learnOutChan) //NewAcceptor func(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor
		acceptor.Start()
	}
	if iAMa(netconf.Learners, appnet.Myself.ID) { //If I'am a learner
		fmt.Println("I am a LEARNER!")
		decidedOutChan = make(chan multipaxos.DecidedValue, 3000)
		//valueOutChan = make(chan multipaxos.Value, 3000)
		learner = multipaxos.NewLearner(appnet.Myself.ID, len(netconf.Acceptors), decidedOutChan) //NewLearner func(id int, nrOfNodes int, valueOut chan<- Value) *Learner
		learner.Start()
	}

	//Try to connect (TCP) to each node in the network, before the application starts it's own TCP server
	fmt.Println("InitConns and StartServer")
	appnet.InitConns()   //Try to initiate TCP connections to nodes in network (peer to peer topology) - net.DialTCP()
	appnet.StartServer() //Starts it own TCP server after initating TCP conns to other nodes - net.ListenTCP()
	fmt.Println("InitConns and StartServer done")

	fd.Start()

	//Implement Client Handler: Rewrite
	cliConnChan := appnet.ClientConnChan
	clihandler := clienthandler2.NewClientHandler(appnet.Myself.ID, proposer, ld, cliConnChan)
	clihandler.Start()

	//done channel to receive os signal (like ctrl+c). Should close TCP connections correctly
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)

	for {
		select {
		//New leader notifications
		case newLeader := <-ldChan: //If ld publish a new leader
			fmt.Printf("\nNew leader: %d \n", newLeader)
		//Sending Heartbeat
		case hb := <-hbSend: //If hb reply
			//fmt.Printf("\n{From: %v, To: %v, Request: %v}\n", hb.From, hb.To, hb.Request)
			//Send hearbeat
			sendHBmsg := network.Message{
				Type:    "Heartbeat",
				To:      hb.To,
				From:    hb.From,
				Msg:     "random message",
				Request: hb.Request,
			}
			//fmt.Println("Sending Heartbeat")
			appnet.SendChan <- sendHBmsg //Send sendHBmsg on sendChan
		//Send Prepare from proposer to acceptors
		case prp := <-prepareOutChan:
			prpMsg := network.Message{ // Type, To, From, string, Request, Prepare, Promise, Accept, Learn, Value
				Type:    "Prepare",
				From:    prp.From,
				Prepare: prp,
			}
			fmt.Printf("Main: (Proposer) %d sent prepare to acceptors %v: %v\n", appnet.Myself.ID, netconf.Acceptors, prp)
			appnet.SendMsgTo(prpMsg, netconf.Acceptors)
		//Send promise from acceptor to proposer
		case prm := <-promiseOutChan:
			prmMsg := network.Message{
				Type:    "Promise",
				To:      prm.To,
				From:    prm.From,
				Promise: prm,
			}
			fmt.Printf("Main: (Acceptor) %d sent promise to proposer %d: %v\n", appnet.Myself.ID, prm.To, prm)
			appnet.SendMessage(prmMsg)
		//Send accept from proposer to acceptors
		case acc := <-acceptOutChan:
			accMsg := network.Message{
				Type:   "Accept",
				From:   acc.From,
				Accept: acc,
			}
			fmt.Printf("Main: (Proposer) %d sent accept to acceptors %v: %v\n", appnet.Myself.ID, netconf.Acceptors, acc)
			appnet.SendMsgTo(accMsg, netconf.Acceptors)
		//Send learn from acceptor to learners
		case lrn := <-learnOutChan:
			lrnMsg := network.Message{
				Type:  "Learn",
				From:  lrn.From,
				Learn: lrn,
			}
			fmt.Printf("Main: (Acceptor) %d sent learn to learners %v: %v\n", appnet.Myself.ID, netconf.Learners, lrn)
			appnet.SendMsgTo(lrnMsg, netconf.Learners)
		case decidedValue := <-decidedOutChan:
			fmt.Printf("Main: (Learner) %d sent decided value to client\n", appnet.Myself.ID)
			clihandler.SendValToCli(decidedValue)
			handleDecidedValue(decidedValue)
		case rMsg := <-appnet.ReceiveChan:
			switch {
			case rMsg.Type == "Heartbeat":
				hb := detector.Heartbeat{
					To:      rMsg.To,
					From:    rMsg.From,
					Request: rMsg.Request,
				}
				fd.DeliverHeartbeat(hb) //Deliver hearbeat to fd
			case rMsg.Type == "Prepare":
				fmt.Printf("Main: (Acceptor) %d got prepare from %d: %v\n", appnet.Myself.ID, rMsg.From, rMsg.Prepare)
				acceptor.DeliverPrepare(rMsg.Prepare)
			case rMsg.Type == "Promise":
				fmt.Printf("Main: (Proposer) %d got promise from %d: %v\n", appnet.Myself.ID, rMsg.From, rMsg.Promise)
				proposer.DeliverPromise(rMsg.Promise)
			case rMsg.Type == "Accept":
				fmt.Printf("Main: (Acceptor) %d got accept from %d: %v\n", appnet.Myself.ID, rMsg.From, rMsg.Accept)
				acceptor.DeliverAccept(rMsg.Accept)
			case rMsg.Type == "Learn":
				fmt.Printf("Main: (Learner) %d got learn from %d: %v\n", appnet.Myself.ID, rMsg.From, rMsg.Learn)
				learner.DeliverLearn(rMsg.Learn)
			case rMsg.Type == "Value":
				fmt.Printf("\nMain: (Clienthandler) %d got value from %d: %v\n", appnet.Myself.ID, rMsg.From, rMsg.Value)
				clihandler.DeliverClientValue(rMsg.Value)
			case rMsg.Type == "ClientInfo":
				fmt.Printf("\nMain: (Clienthandler) %d got client info from %v\n", appnet.Myself.ID, rMsg.ClientInfo.ClientID)
				clihandler.DeliverClientInfo(rMsg)
			}
		case <-done:
			proposer.Stop()
			acceptor.Stop()
			learner.Stop()
			for _, tcpConn := range appnet.Connections {
				appnet.CloseConn(tcpConn)
			}
			for _, tcpConn := range appnet.ClientConns {
				appnet.CloseConn(tcpConn)
			}
			os.Exit(0)
		}
	}
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

func iAMa(roleIDs []int, myself int) bool {
	for _, ID := range roleIDs {
		if ID == myself {
			return true
		}
	}
	return false
}
