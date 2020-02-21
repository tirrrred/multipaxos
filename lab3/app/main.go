package main

import (
	"encoding/json"
	"fmt"
	"github.com/tirrrred/multipaxos/lab3/network"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	fmt.Println("Hello main app")
	netconf, _ := importNetConf()
	appnet, err := network.InitNetwork(netconf.Nodes, netconf.Myself)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(appnet)
	appnet.StartServer()
	appnet.InitConns()

}

func importNetConf() (network.NetConfig, error) {
	//Open the network config file for the application
	netConfFile, err := os.Open("netConf.json")
	if err != nil {
		log.Fatal(err)
	}
	defer netConfFile.Close()

	//Read the network config file as a byte array
	byteValue, _ := ioutil.ReadAll(netConfFile)

	//Init the netConfig struct to store the network config from file
	var netconf network.NetConfig

	err = json.Unmarshal(byteValue, &netconf)
	if err != nil {
		log.Fatal(err)
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
