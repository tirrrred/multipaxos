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

	//**Get arguments from command line START**//
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Need one argument: Text to send")

	}
	//**Get arguments from command line END**//
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

	//subscribe to leader changes
	//ldSubscribtion := ld.Subscribe()

	//create failure detector
	hbChan := make(chan detector.Heartbeat, 16)
	fd := detector.NewEvtFailureDetector(appnet.Myself.ID, nodeIDs, ld, 1*time.Second, hbChan) //how to get things sent on the hbChan onto the network???

	fmt.Println(fd) //Remove

	fmt.Println("InitConns and StartServer done")

	appnet.InitConns()
	appnet.StartServer()

	err = appnet.SendMessage(1, arguments[1])
	if err != nil {
		log.Print(err)
	}

	//osSignalChan := make(chan os.Signal, 1)
	/*
		for {
			select {
			case NewLeader := <-ldSubscribtion:
				log.Printf("\nApplication %d: LEADER CHANGE - New leader is: %d \n", appnet.Myself.ID, NewLeader)
			case <-osSignalChan:
				appnet.Myself.TCPListen.Close()
				appnet.Myself.TCPListen = nil
				os.Exit(0)
			default:
				log.Printf("\nApplication %d: Default message....", appnet.Myself.ID)
			}
		} */
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
