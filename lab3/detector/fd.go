// +build !solution

package detector

import (
	"time"
)

// EvtFailureDetector represents a Eventually Perfect Failure Detector as
// described at page 53 in:
// Christian Cachin, Rachid Guerraoui, and Luís Rodrigues: "Introduction to
// Reliable and Secure Distributed Programming" Springer, 2nd edition, 2011.
type EvtFailureDetector struct {
	id        int          // this node's id
	nodeIDs   []int        // node ids for every node in cluster
	alive     map[int]bool // map of node ids considered alive
	suspected map[int]bool // map of node ids  considered suspected

	sr SuspectRestorer // Provided SuspectRestorer implementation

	delay         time.Duration // the current delay for the timeout procedure
	delta         time.Duration // the delta value to be used when increasing delay
	timeoutSignal *time.Ticker  // the timeout procedure ticker

	hbSend chan<- Heartbeat // channel for sending outgoing heartbeat messages
	hbIn   chan Heartbeat   // channel for receiving incoming heartbeat messages
	stop   chan struct{}    // channel for signaling a stop request to the main run loop

	testingHook func() // DO NOT REMOVE THIS LINE. A no-op when not testing.
}

// NewEvtFailureDetector returns a new Eventual Failure Detector. It takes the
// following arguments:
//
// id: The id of the node running this instance of the failure detector.
//
// nodeIDs: A list of ids for every node in the cluster (including the node
// running this instance of the failure detector).
//
// ld: A leader detector implementing the SuspectRestorer interface.
//
// delta: The initial value for the timeout interval. Also the value to be used
// when increasing delay.
//
// hbSend: A send only channel used to send heartbeats to other nodes.
func NewEvtFailureDetector(id int, nodeIDs []int, sr SuspectRestorer, delta time.Duration, hbSend chan<- Heartbeat) *EvtFailureDetector {
	suspected := make(map[int]bool)
	alive := make(map[int]bool)

	// TODO(student): perform any initialization necessary
	for _, node := range nodeIDs {
		alive[node] = true //assumes all provided node IDs is alive (when init fd)
	}

	return &EvtFailureDetector{
		id:        id,
		nodeIDs:   nodeIDs,
		alive:     alive,
		suspected: suspected,

		sr: sr,

		delay: delta,
		delta: delta,

		hbSend: hbSend,
		hbIn:   make(chan Heartbeat, 8),
		stop:   make(chan struct{}),

		testingHook: func() {}, // DO NOT REMOVE THIS LINE. A no-op when not testing.
	}
}

// Start starts e's main run loop as a separate goroutine. The main run loop
// handles incoming heartbeat requests and responses. The loop also trigger e's
// timeout procedure at an interval corresponding to e's internal delay
// duration variable.
func (e *EvtFailureDetector) Start() {
	e.timeoutSignal = time.NewTicker(e.delay)
	go func() {
		for {
			e.testingHook() // DO NOT REMOVE THIS LINE. A no-op when not testing.
			select {
			case incHB := <-e.hbIn: //heartbeath comming in
				// TODO(student): Handle incoming heartbeat
				if incHB.Request && incHB.To == e.id {
					//If heartbeat is actual meant for us and it is a request, send reply back to sender
					hbReply := Heartbeat{To: incHB.From, From: e.id, Request: false}
					//fmt.Printf("\nfd.go recived incoming HB and creates reply: {To: %d From: %d Request: %v}\n", incHB.From, e.id, false)
					e.ReplyHeartbeat(hbReply)
				} else if incHB.Request == false && incHB.To == e.id {
					//incHB is a reply (incHB.Request == false)
					//incHB is addressed to us (incHB.To == e.id)
					//Set incHB.From to alive
					e.alive[incHB.From] = true
				}
			case <-e.timeoutSignal.C:
				e.timeout()
			case <-e.stop:
				return
			}
		}
	}()
}

// DeliverHeartbeat delivers heartbeat hb to failure detector e.
func (e *EvtFailureDetector) DeliverHeartbeat(hb Heartbeat) {
	e.hbIn <- hb
}

// Stop stops e's main run loop.
func (e *EvtFailureDetector) Stop() {
	e.stop <- struct{}{}
}

// Internal: timeout runs e's timeout procedure.
func (e *EvtFailureDetector) timeout() {
	// TODO(student): Implement timeout procedure
	// Based on Algorithm 2.7: Increasing Timeout at page 55
	//if alive ∩ suspected != ∅ then:
	if len(e.intersection(e.alive, e.suspected)) > 0 {
		//delay := delay +Δ;
		e.delay = e.delay + e.delta
	}
	// forall p ∈ Π do
	for _, nodeID := range e.nodeIDs {
		// if (p !∈ alive) ∧ (p !∈ suspected) then
		if e.inAlive(nodeID) == false && e.inSuspected(nodeID) == false {
			//suspected := suspected ∪{p};
			e.suspected[nodeID] = true
			//trigger P, Suspect | p;
			e.sr.Suspect(nodeID)
			//else if (p ∈ alive) ∧ (p ∈ suspected) then
		} else if e.inAlive(nodeID) && e.inSuspected(nodeID) {
			//suspected := suspected \{p};
			delete(e.suspected, nodeID)
			//e.suspected[nodeID] = false
			//trigger P, Restore | p;
			e.sr.Restore(nodeID)
		}
		//trigger pl, Send | p, [HEARTBEATREQUEST];
		hbReq := Heartbeat{From: e.id, To: nodeID, Request: true}
		e.hbSend <- hbReq
	}
	//alive := ∅;
	emptyAlive := make(map[int]bool)
	e.alive = emptyAlive
	//starttimer(delay);
	e.timeoutSignal.Stop()
	e.timeoutSignal = time.NewTicker(e.delay)
}

// TODO(student): Add other unexported functions or methods if needed.

// ReplyHeartbeat replies to incoming heartbeats hb to failure detector e.
func (e *EvtFailureDetector) ReplyHeartbeat(hb Heartbeat) {
	e.hbSend <- hb
}

func (e *EvtFailureDetector) intersection(a, b map[int]bool) []int {
	var elements []int
	for keyA := range a {
		for keyB := range b {
			if keyA == keyB {
				elements = append(elements, keyA)
			}
		}
	}
	return elements
}

//used?
func (e *EvtFailureDetector) isAlive(nodeID int) bool {
	return e.alive[nodeID]
}

//used?
func (e *EvtFailureDetector) isSuspected(nodeID int) bool {
	return e.suspected[nodeID]
}

// if (p !∈ alive)
func (e *EvtFailureDetector) inAlive(nodeID int) bool {
	for node := range e.alive {
		if nodeID == node {
			return true
		}
	}
	return false
}

// if (p !∈ suspected)
func (e *EvtFailureDetector) inSuspected(nodeID int) bool {
	for node := range e.suspected {
		if nodeID == node {
			return true
		}
	}
	return false
}
