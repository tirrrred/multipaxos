// +build !solution

package multipaxos

import (
	"sort"
)

// Acceptor represents an acceptor as defined by the Multi-Paxos algorithm.
type Acceptor struct {
	// TODO(student)
	rnd      Round
	id       int
	accSlots []PromiseSlot
	maxSlot  SlotID
	//accSlot  map[SlotID]votedTuple

	promiseChanOut chan<- Promise
	learnChanOut   chan<- Learn
	prepareChanIn  chan Prepare
	acceptChanIn   chan Accept
	stop           chan struct{}
}

// NewAcceptor returns a new Multi-Paxos acceptor. It takes the following
// arguments:
//
// id: The id of the node running this instance of a Paxos acceptor.
//
// promiseOut: A send only channel used to send promises to other nodes.
//
// learnOut: A send only channel used to send learns to other nodes.
func NewAcceptor(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor {
	return &Acceptor{
		id:       id,
		rnd:      NoRound,
		accSlots: []PromiseSlot{},

		promiseChanOut: promiseOut,
		learnChanOut:   learnOut,
		prepareChanIn:  make(chan Prepare, 3000),
		acceptChanIn:   make(chan Accept, 3000),
		stop:           make(chan struct{}),
	}
}

// Start starts a's main run loop as a separate goroutine. The main run loop
// handles incoming prepare and accept messages.
func (a *Acceptor) Start() {
	go func() {
		for {
			// TODO(student)
			select {
			case prp := <-a.prepareChanIn:
				if prm, ok := a.handlePrepare(prp); ok {
					a.promiseChanOut <- prm
				}
			case acc := <-a.acceptChanIn:
				if lrn, ok := a.handleAccept(acc); ok {
					a.learnChanOut <- lrn
				}
			case <-a.stop:
				break
			}
		}
	}()
}

// Stop stops a's main run loop.
func (a *Acceptor) Stop() {
	// TODO(student)
	a.stop <- struct{}{}
}

// DeliverPrepare delivers prepare prp to acceptor a.
func (a *Acceptor) DeliverPrepare(prp Prepare) {
	// TODO(student)
	a.prepareChanIn <- prp
}

// DeliverAccept delivers accept acc to acceptor a.
func (a *Acceptor) DeliverAccept(acc Accept) {
	// TODO(student)
	a.acceptChanIn <- acc
}

// Internal: handlePrepare processes prepare prp according to the Multi-Paxos
// algorithm. If handling the prepare results in acceptor a emitting a
// corresponding promise, then output will be true and prm contain the promise.
// If handlePrepare returns false as output, then prm will be a zero-valued
// struct.
func (a *Acceptor) handlePrepare(prp Prepare) (prm Promise, output bool) {
	// TODO(student)
	//type Prepare struct{From int; Slot SlotID; Crnd Round}
	if prp.Crnd > a.rnd {
		a.rnd = prp.Crnd
		var prmSlots []PromiseSlot
		for _, slot := range a.accSlots {
			if slot.ID >= prp.Slot {
				prmSlots = append(prmSlots, slot)
			}
		}
		return Promise{To: prp.From, From: a.id, Rnd: a.rnd, Slots: prmSlots}, true
	}
	//type Promise struct{To int; From int; Rnd Round; Slots []PromiseSlot}
	return prm, false
}

// Internal: handleAccept processes accept acc according to the Multi-Paxos
// algorithm. If handling the accept results in acceptor a emitting a
// corresponding learn, then output will be true and lrn contain the learn.  If
// handleAccept returns false as output, then lrn will be a zero-valued struct.
func (a *Acceptor) handleAccept(acc Accept) (lrn Learn, output bool) {
	// TODO(student)
	//type PromiseSlot struct{ID SlotID; Vrnd Round; Vval Value}
	//fmt.Printf("Accept rnd = %v, acceptors rnd = %v", acc.Rnd, a.rnd)
	if acc.Rnd >= a.rnd {
		a.rnd = acc.Rnd
		for i, slot := range a.accSlots {
			if slot.ID == acc.Slot {
				slot.Vrnd = acc.Rnd
				slot.Vval = acc.Val
				a.accSlots = append(a.accSlots[:i], append([]PromiseSlot{slot}, a.accSlots[i+1:]...)...)
				return Learn{From: a.id, Slot: slot.ID, Rnd: slot.Vrnd, Val: slot.Vval}, true
			}
		}
		a.accSlots = append(a.accSlots, PromiseSlot{ID: acc.Slot, Vrnd: acc.Rnd, Vval: acc.Val})

		sort.SliceStable(a.accSlots, func(i, j int) bool {
			return a.accSlots[i].ID < a.accSlots[j].ID
		})

		//type Learn struct{From int; Slot SlotID; Rnd Round; Val Value}
		return Learn{From: a.id, Slot: acc.Slot, Rnd: acc.Rnd, Val: acc.Val}, true
	}
	return lrn, false
}
