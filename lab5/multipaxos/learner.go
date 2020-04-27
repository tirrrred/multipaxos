// +build !solution

package multipaxos

//import "dat520/lab5/bank"

// Learner represents a learner as defined by the Multi-Paxos algorithm.
type Learner struct {
	// TODO(student)
	id       int
	n        int
	rnd      Round
	lrnSlots map[SlotID][]Learn
	lrnSent  map[SlotID]bool

	decidedChanOut chan<- DecidedValue
	learnChanIn    chan Learn
	stop           chan struct{}
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
		id:  id,
		n:   nrOfNodes,
		rnd: NoRound,

		lrnSlots: make(map[SlotID][]Learn),
		lrnSent:  make(map[SlotID]bool),

		decidedChanOut: decidedOut,
		learnChanIn:    make(chan Learn, 3000),
		stop:           make(chan struct{}),
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			// TODO(student)
			select {
			case lrn := <-l.learnChanIn:
				if val, slot, ok := l.handleLearn(lrn); ok == true {
					l.decidedChanOut <- DecidedValue{SlotID: slot, Value: val} //type DecidedValue struct{SlotID SlotID; Value Value}
				}
			case <-l.stop:
				break
			}
		}
	}()
}

// Stop stops l's main run loop.
func (l *Learner) Stop() {
	// TODO(student)
	l.stop <- struct{}{}
}

// DeliverLearn delivers learn lrn to learner l.
func (l *Learner) DeliverLearn(lrn Learn) {
	// TODO(student)
	l.learnChanIn <- lrn
}

// Internal: handleLearn processes learn lrn according to the Multi-Paxos
// algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true, sid the id for the
// slot that was decided and val contain the decided value. If handleLearn
// returns false as output, then val and sid will have their zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, sid SlotID, output bool) {
	// TODO(student)
	//type Learn struct{From int; Slot SlotID; Rnd Round; Val Value}
	//Rnd is lower than learned before = Ignore
	if learn.Rnd < l.rnd {
		return val, sid, false
	} else if learn.Rnd == l.rnd { //Rnd is equal to
		if learnsInSlot, ok := l.lrnSlots[learn.Slot]; ok { //learnsInSlot
			//for _, learnInSlot := range learnsIn
			//Check if we learned this before:
			for _, learnInSlot := range learnsInSlot {
				//If we learned from this node before with the same rnd = Ignore (cannot learn from same node with same rnd)
				if learnInSlot.From == learn.From && learnInSlot.Rnd == learn.Rnd {
					return val, sid, false
				}
				/*//If we learned from this node before, but now with a greater rnd: //Unnecessary?
				if learnInSlot.From == learn.From && learnInSlot.Rnd < learn.Rnd {
					//Replace learn.
					//a.accSlots = append(a.accSlots[:i], append([]PromiseSlot{slot}, a.accSlots[i+1:]...)...)
					learnsInSlot[i] = learn
					l.lrnSlots[learn.Slot] = learnsInSlot*/
			}
			l.lrnSlots[learn.Slot] = append(l.lrnSlots[learn.Slot], learn)
		} else {
			l.lrnSlots[learn.Slot] = append(l.lrnSlots[learn.Slot], learn)
		}
	} else { //Rnd is greater than learned before
		l.rnd = learn.Rnd
		l.lrnSlots[learn.Slot] = nil
		l.lrnSlots[learn.Slot] = append(l.lrnSlots[learn.Slot], learn)
	}

	//Check if learned sent for this SID and quroum already reach
	if l.lrnSent[learn.Slot] {
		return val, sid, false
	}
	//Check if quroum/Majority
	if len(l.lrnSlots[learn.Slot]) >= ((l.n / 2) + 1) {
		l.lrnSent[learn.Slot] = true
		return learn.Val, learn.Slot, true
	}
	return val, sid, false
}
