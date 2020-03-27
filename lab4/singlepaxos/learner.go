// +build !solution

package singlepaxos

import "fmt"

// Learner represents a learner as defined by the single-decree Paxos
// algorithm.
type Learner struct {
	ID             int
	HighestRound   Round //Highest Round Seen
	ConsensusValue Value
	NumNodes       int
	LrnVals        []LearnedValues
	ValueOutChan   chan<- Value
	stopChan       chan int
	learnChan      chan Learn

	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	// Add needed fields
}

//LearnedValues is stores learned rounds and values pair from acceptors. Used to determine if a majority of the network has agreed on a value
type LearnedValues struct {
	From  int
	Round Round
	Value Value
}

// NewLearner returns a new single-decree Paxos learner. It takes the
// following arguments:
//
// id: The id of the node running this instance of a Paxos learner.
//
// nrOfNodes: The total number of Paxos nodes.
//
// valueOut: A send only channel used to send values that has been learned,
// i.e. decided by the Paxos nodes.
func NewLearner(id int, nrOfNodes int, valueOut chan<- Value) *Learner {
	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	return &Learner{
		ID:             id,
		HighestRound:   NoRound,
		ConsensusValue: ZeroValue,
		NumNodes:       nrOfNodes,
		ValueOutChan:   valueOut,
		stopChan:       make(chan int),
		learnChan:      make(chan Learn),
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			//TODO(student): Task 3 - distributed implementation
			select {
			case lrn := <-l.learnChan:
				if val, learned := l.handleLearn(lrn); learned == true {
					l.ValueOutChan <- val
				}
			case <-l.stopChan:
				break
			}
		}
	}()
}

// Stop stops l's main run loop.
func (l *Learner) Stop() {
	//TODO(student): Task 3 - distributed implementation
	l.stopChan <- 0
}

// DeliverLearn delivers learn lrn to learner l.
func (l *Learner) DeliverLearn(lrn Learn) {
	//TODO(student): Task 3 - distributed implementation
	l.learnChan <- lrn
}

// Internal: handleLearn processes learn lrn according to the single-decree
// Paxos algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true and val contain the
// decided value. If handleLearn returns false as output, then val will have
// its zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, output bool) {
	//TODO(student): Task 2 - algorithm implementation
	if learn.Rnd > l.HighestRound {
		l.HighestRound = learn.Rnd // New highest round seen
		l.LrnVals = nil            // Need to reset learned values slice, since the old learned values are from a lower/older round
	} else if learn.Rnd < l.HighestRound || learn.Rnd == NoRound { //If you get a learn with a lower (or not valid) Round, do nothing
		return "", false
	}
	//Only append valid learn messages:
	// 1) One acceptor cannot spam learned and "brute force" a learned value = Only learned value per acceptor ID
	for _, lv := range l.LrnVals {
		if lv.From == learn.From { //If the learner has already learned from this acceptor ID with the given Rnd (or lower)
			return "", false //Do nothing
		}
	}
	l.LrnVals = append(l.LrnVals, LearnedValues{From: learn.From, Round: learn.Rnd, Value: learn.Val})
	if len(l.LrnVals) >= l.NumNodes/2+1 { //If we learned values from a majority of acceptors
		//Check if any value gives a majority
		if verifiedValue, majorityFound := l.majorityValue(l.LrnVals); majorityFound == true {
			fmt.Printf("Learner: Consensus reached on value: %v\n", verifiedValue)
			return verifiedValue, true
		}
	}

	//l.LrnVals = append(l.LrnVals, LearnedValues{From: learn.From, Round: learn.Rnd, Value: learn.Val})
	//LrnVals: (Learn messages from acceptors roles)
	//From 0, Round 5, Foo
	//From 1, Round 5, Bar
	//From 2, Round 5, Foo
	fmt.Printf("Learner: Consensus not reached for value: %v\n", learn.Val)
	return "", false
}

//TODO(student): Add any other unexported methods needed.
func (l *Learner) majorityValue(lv []LearnedValues) (val Value, output bool) {
	valueCount := make(map[Value]int)
	//Count every value in slice
	for _, v := range lv {
		valueCount[v.Value]++
	}
	//Check if any value has majority
	for val, count := range valueCount {
		if count >= l.NumNodes/2+1 {
			return val, true
		}
	}
	return "", false
}
