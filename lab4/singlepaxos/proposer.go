// +build !solution

package singlepaxos

import (
	"fmt"
	"github.com/tirrrred/multipaxos/lab3/detector"
)

// Proposer represents a proposer as defined by the single-decree Paxos
// algorithm.
type Proposer struct {
	crnd             Round
	clientValue      Value
	ConstrainedValue Value
	ID               int
	NumNodes         int
	PromiseRequests  []Promise
	LeaderDetector   detector.LeaderDetector
	PrepareOutChan   chan<- Prepare
	AcceptOutChan    chan<- Accept
	promiseChan      chan Promise
	clientValueChan  chan Value
	stopChan         chan int

	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	// Add other needed fields
}

// NewProposer returns a new single-decree Paxos proposer.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos proposer.
//
// nrOfNodes: The total number of Paxos nodes.
//
// ld: A leader detector implementing the detector.LeaderDetector interface.
//
// prepareOut: A send only channel used to send prepares to other nodes.
//
// The proposer's internal crnd field should initially be set to the value of
// its id.
func NewProposer(id int, nrOfNodes int, ld detector.LeaderDetector, prepareOut chan<- Prepare, acceptOut chan<- Accept) *Proposer {
	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	return &Proposer{
		crnd:             Round(id),
		clientValue:      ZeroValue,
		ConstrainedValue: ZeroValue,
		ID:               id,
		NumNodes:         nrOfNodes,
		PromiseRequests:  []Promise{},
		LeaderDetector:   ld,
		PrepareOutChan:   prepareOut,
		AcceptOutChan:    acceptOut,
		promiseChan:      make(chan Promise, 16),
		clientValueChan:  make(chan Value, 16),
		stopChan:         make(chan int),
	}
}

// Start starts p's main run loop as a separate goroutine. The main run loop
// handles incoming promise messages and leader detector trust messages.
func (p *Proposer) Start() {
	go func() {
		for {
			//TODO(student): Task 3 - distributed implementation
			select {
			case prm := <-p.promiseChan:
				if accMsg, sendMsg := p.handlePromise(prm); sendMsg == true {
					p.AcceptOutChan <- accMsg
				}
			case cVal := <-p.clientValueChan:
				if prpMsg, sendMsg := p.clientHandler(cVal); sendMsg == true {
					p.PrepareOutChan <- prpMsg
				}
			case <-p.stopChan:
				break
			}
		}
	}()
}

// Stop stops p's main run loop.
func (p *Proposer) Stop() {
	//TODO(student): Task 3 - distributed implementation
	p.stopChan <- 0
}

// DeliverPromise delivers promise prm to proposer p.
func (p *Proposer) DeliverPromise(prm Promise) {
	//TODO(student): Task 3 - distributed implementation
	p.promiseChan <- prm
}

// DeliverClientValue delivers client value val from to proposer p.
func (p *Proposer) DeliverClientValue(val Value) {
	//TODO(student): Task 3 - distributed implementation
	p.clientValueChan <- val
}

// Internal: handlePromise processes promise prm according to the single-decree
// Paxos algorithm. If handling the promise results in proposer p emitting a
// corresponding accept, then output will be true and acc contain the promise.
// If handlePromise returns false as output, then acc will be a zero-valued
// struct.
func (p *Proposer) handlePromise(prm Promise) (acc Accept, output bool) {
	//TODO(student): Task 2 - algorithm implementation
	//v = Desired Consensus Value
	//crnd = Current Round (unique)
	//cval = Constrained Consensus Value, i.e not freely choosen by the client/proposer.
	//vrnd = Round in which a Value was Last Accepted
	//vval = Value Last Accepted
	fmt.Println("Promiser: handlePromise(prm) Start")

	if prm.Rnd > p.crnd {
		p.crnd = prm.Rnd
		p.PromiseRequests = nil
		//p.PromiseRequests = append(p.PromiseRequests, prm)
		//return Accept{}, false
	} else if prm.Rnd < p.crnd || prm.Rnd == NoRound {
		return Accept{}, false
	}

	//p.crnd = prm.rnd at this stage
	if len(p.PromiseRequests) > 0 {
		for i, prmReq := range p.PromiseRequests {
			if prmReq.From == prm.From { //If we already got a promise request from this node
				if prmReq.Rnd < prm.Rnd { //Check if the new promise request has a bigger Rnd than the current (old) promise request
					p.PromiseRequests[i] = prm //If yes, replace the old request with the new one
				}
			} else {
				p.PromiseRequests = append(p.PromiseRequests, prm) //If we don't have a promise request from this node before, append it.
			}

		}
	}

	p.PromiseRequests = append(p.PromiseRequests, prm)
	//If the proposer has gotten promise messages from a majority of acceptors
	if len(p.PromiseRequests) >= p.NumNodes/2+1 {
		//Check if the Promise messages have a enough "last voted values (vval)" to form a majority
		if cstrVal, status := p.constrainedValue(p.PromiseRequests); status == true { //re-name these variables, bad names
			//If yes, set the constrainedValue with the majority values from the acceptors
			p.ConstrainedValue = cstrVal
			fmt.Println("Promiser: handlePromise(prm) return constraintedValue, true")
			return Accept{From: p.ID, Rnd: p.crnd, Val: p.ConstrainedValue}, true
		}
		//If there is no majority values out there, use my clientValue (Due to the test, set a dummy value first)
		//p.clientValue = Value("defaultValue")
		fmt.Println("Promiser: handlePromise(prm) return clientvalue, true")
		return Accept{From: p.ID, Rnd: p.crnd, Val: p.clientValue}, true
	}
	return Accept{}, false
}

// Internal: increaseCrnd increases proposer p's crnd field by the total number
// of Paxos nodes.
func (p *Proposer) increaseCrnd() {
	//TODO(student): Task 2 - algorithm implementation
	p.crnd += Round(p.NumNodes)
	return
}

//TODO(student): Add any other unexported methods needed.
func (p *Proposer) constrainedValue(prmReqs []Promise) (val Value, output bool) {
	rndValue := make(map[Round]Value)
	//Count every value in slice
	for _, prmReq := range prmReqs {
		if prmReq.Vval == ZeroValue {
			continue
		}
		rndValue[prmReq.Vrnd] = prmReq.Vval
	}
	rndIndex := NoRound
	for rnd := range rndValue {
		if rnd > rndIndex {
			rndIndex = rnd
		}
		return rndValue[rndIndex], true
	}
	return "", false
}

func (p *Proposer) clientHandler(cVal Value) (prp Prepare, output bool) {
	leaderNode := p.LeaderDetector.Leader()
	if leaderNode != p.ID {
		return Prepare{}, false
	}
	p.clientValue = cVal
	p.increaseCrnd()
	return Prepare{From: p.ID, Crnd: p.crnd}, true
}
