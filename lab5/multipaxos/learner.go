// +build !solution

package multipaxos

import "dat520/lab5/bank"

// Learner represents a learner as defined by the Multi-Paxos algorithm.
type Learner struct {
	// TODO(student)
}

// NewLearner returns a new Multi-Paxos learner. It takes the following
// arguments:
//
// id: The id of the node running this instance of a Paxos learner.
//
// nrOfNodes: The total number of Paxos nodes.
//
// decidedOut: A send only channel used to send values that has been learned,
// i.e. decided by the Paxos nodes.
func NewLearner(id int, nrOfNodes int, decidedOut chan<- DecidedValue) *Learner {
	return &Learner{
		// TODO(student)
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			// TODO(student)
		}
	}()
}

// Stop stops l's main run loop.
func (l *Learner) Stop() {
	// TODO(student)
}

// DeliverLearn delivers learn lrn to learner l.
func (l *Learner) DeliverLearn(lrn Learn) {
	// TODO(student)
}

// Internal: handleLearn processes learn lrn according to the Multi-Paxos
// algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true, sid the id for the
// slot that was decided and val contain the decided value. If handleLearn
// returns false as output, then val and sid will have their zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, sid SlotID, output bool) {
	// TODO(student)
	return Value{ClientID: "-1", ClientSeq: -1, AccountNum: -1, Txn: bank.Transaction{}}, -1, true
}
