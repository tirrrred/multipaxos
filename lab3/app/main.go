package main

import (
	"encoding/json"
	"fmt"
	"github.com/tirrrred/multipaxos/lab3/detector"
	"github.com/tirrrred/multipaxos/lab3/network"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Hello main app")

	netconf, _ := importNetConf()
	appnet, err := network.InitNetwork(netconf.Nodes, netconf.Myself)
	if err != nil {
		log.Print(err)
	}

	nodeIDs := []int{}
	for _, node := range appnet.Nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}
	//create leader detector
	ld := detector.NewMonLeaderDetector(nodeIDs) //*MonLeaderDetector

	//create failure detector
	hbSend := make(chan detector.Heartbeat, 16)
	fd := detector.NewEvtFailureDetector(appnet.Myself.ID, nodeIDs, ld, 5*time.Second, hbSend) //how to get things sent on the hbChan onto the network???

	//fmt.Println(fd) //Remove

	fmt.Println("InitConns and StartServer")
	appnet.InitConns()
	appnet.StartServer()
	fmt.Println("InitConns and StartServer done")
	//subscribe to leader changes
	ldChan := ld.Subscribe()
	fd.Start()

	for {
		select {
		case newLeader := <-ldChan: //If ld publish a new leader
			fmt.Printf("\nNew leader: %d \n", newLeader)
		case hb := <-hbSend: //If hb reply
			//fmt.Printf("\n{From: %v, To: %v, Request: %v}\n", hb.From, hb.To, hb.Request)
			//Send hearbeat
			sendHBmsg := network.Message{
				To:      hb.To,
				From:    hb.From,
				Msg:     "",
				Request: hb.Request,
			}
			fmt.Printf("\nappnet.SendChan <- sendHBmsg: %v\n", sendHBmsg)
			appnet.SendChan <- sendHBmsg //Send sendHBmsg on sendChan
		case receivedHBmsg := <-appnet.ReceiveChan: //If recivedHBmsg from receiveChan
			hb := detector.Heartbeat{
				To:      receivedHBmsg.To,
				From:    receivedHBmsg.From,
				Request: receivedHBmsg.Request,
			}
			//fmt.Printf("\n{From: %v, To: %v, Request: %v}\n", hb.From, hb.To, hb.Request)
			fmt.Printf("\nfd.DeliverHeartbeat(hb): %v\n", hb)
			fd.DeliverHeartbeat(hb) //Deliver hearbeat to fd
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
