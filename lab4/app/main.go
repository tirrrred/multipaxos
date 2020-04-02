package main

//TO add based on LAB4:
//Unicast and brodcast. Only "Promise" messsages is unicast (has a 'to' field). Learn, Accept and Prepare are all broadcast
//Use it own NodeID as unique round number (rnd). Current round = (Crnd) and Voted round = (Vrnd) -> Type conversion. IncreaseCrnd adds the number of Paxos nodes to this unique round number (NodeID)

import (
	"encoding/json"
	"fmt"
	"github.com/dat520-2020/assignments/lab3/detector"
	"github.com/dat520-2020/assignments/lab3/network"
	"github.com/dat520-2020/assignments/lab4/clienthandler"
	"github.com/dat520-2020/assignments/lab4/singlepaxos"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	fmt.Println("Hello main app")

	//import network configuration based on netConfig.json file
	netconf, _ := importNetConf()

	//Initiate this applications network
	appnet, err := network.InitNetwork(netconf.Nodes, netconf.Myself)
	if err != nil {
		log.Print(err)
	}

	//Creates a slice of node IDs - TO be used in newMonLeaderDetector
	nodeIDs := []int{appnet.Myself.ID}
	for _, node := range appnet.Nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}
	fmt.Printf("Nodes in Network: %v \n", nodeIDs)
	//create leader detector, based on proposers ID in the network
	ld := detector.NewMonLeaderDetector(netconf.Proposers) //*MonLeaderDetector

	//create failure detector
	hbSend := make(chan detector.Heartbeat, 24)
	fd := detector.NewEvtFailureDetector(appnet.Myself.ID, nodeIDs, ld, 2*time.Second, hbSend) //2 second timeout for failure detector - Increase/Decrease?

	//subscribe to leader changes
	ldChan := ld.Subscribe()
	fmt.Printf("\nLeader after ld init and subsribtion: %d\n", ld.Leader())

	//Declare PAXOS variables:
	var proposer *singlepaxos.Proposer
	var prepareOutChan chan singlepaxos.Prepare
	var acceptOutChan chan singlepaxos.Accept
	var acceptor *singlepaxos.Acceptor
	var promiseOutChan chan singlepaxos.Promise
	var learnOutChan chan singlepaxos.Learn
	var learner *singlepaxos.Learner
	var valueOutChan chan singlepaxos.Value

	//Implement PAXOS roles:
	if iAMa(netconf.Proposers, appnet.Myself.ID) { //If I'am a proposer
		fmt.Println("I am a PROPOSER!")
		prepareOutChan = make(chan singlepaxos.Prepare, 16)
		acceptOutChan = make(chan singlepaxos.Accept, 16)
		proposer = singlepaxos.NewProposer(appnet.Myself.ID, len(netconf.Acceptors), ld, prepareOutChan, acceptOutChan) //NewProposer func(id int, nrOfNodes int, ld detector.LeaderDetector, prepareOut chan<- Prepare, acceptOut chan<- Accept) *Proposer
		proposer.Start()
	}
	if iAMa(netconf.Acceptors, appnet.Myself.ID) { //If I'am a acceptor
		fmt.Println("I am a ACCEPTOR!")
		promiseOutChan = make(chan singlepaxos.Promise, 16)
		learnOutChan = make(chan singlepaxos.Learn, 16)
		acceptor = singlepaxos.NewAcceptor(appnet.Myself.ID, promiseOutChan, learnOutChan) //NewAcceptor func(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor
		acceptor.Start()
	}
	if iAMa(netconf.Learners, appnet.Myself.ID) { //If I'am a learner
		fmt.Println("I am a LEARNER!")
		valueOutChan = make(chan singlepaxos.Value, 16)
		learner = singlepaxos.NewLearner(appnet.Myself.ID, len(netconf.Acceptors), valueOutChan) //NewLearner func(id int, nrOfNodes int, valueOut chan<- Value) *Learner
		learner.Start()
	}

	//Try to connect (TCP) to each node in the network, before the application starts it's own TCP server
	fmt.Println("InitConns and StartServer")
	appnet.InitConns()   //Try to initiate TCP connections to nodes in network (peer to peer topology) - net.DialTCP()
	appnet.StartServer() //Starts it own TCP server after initating TCP conns to other nodes - net.ListenTCP()
	fmt.Println("InitConns and StartServer done")

	fd.Start()

	//Implement Client Handler:
	cliConnChan := appnet.ClientConnChan
	clihandler := clienthandler.NewClientHandler(appnet.Myself.ID, proposer, cliConnChan)
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
			//fmt.Printf("\nappnet.SendChan <- sendHBmsg: %v\n", sendHBmsg)
			appnet.SendChan <- sendHBmsg //Send sendHBmsg on sendChan
		/*//Receive Heartbeat
		case receivedHBmsg := <-appnet.ReceiveChan: //If recivedHBmsg from receiveChan -> CAN BE REMOVE, REPLACE by rMSG
		hb := detector.Heartbeat{
			To:      receivedHBmsg.To,
			From:    receivedHBmsg.From,
			Request: receivedHBmsg.Request,
		}
		//fmt.Printf("\n{From: %v, To: %v, Request: %v}\n", hb.From, hb.To, hb.Request)
		//fmt.Printf("\nfd.DeliverHeartbeat(hb): %v\n", hb)
		fd.DeliverHeartbeat(hb) //Deliver hearbeat to fd*/
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
		//Send value from learner to client(handler)
		case val := <-valueOutChan:
			clihandler.SendValToCli(val)
		//Receive message - switch to decide which message type received and how to handle it accordingly. May this be a bottleneck when scaled, since it is a serial switch, and a node can receive alot of messages?
		case rMsg := <-appnet.ReceiveChan:
			switch {
			case rMsg.Type == "Heartbeat":
				hb := detector.Heartbeat{
					To:      rMsg.To,
					From:    rMsg.From,
					Request: rMsg.Request,
				}
				//fmt.Printf("Main: (FD) %d got heartbeat from %d\n", appnet.Myself.ID, hb.From)
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
				clihandler.DeliverValue(rMsg.Value)
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
	netConfFile, err := os.Open("netConf.json")
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
