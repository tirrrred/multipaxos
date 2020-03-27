// +build !solution

package singlepaxos

// Acceptor represents an acceptor as defined by the single-decree Paxos
// algorithm.
type Acceptor struct {
	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	// Add needed fields
	ID             int
	HighestRound   Round //Highest Round Seen
	VotedRound     Round //Round in which a Value was Last Accepted
	VotedValue     Value //Value Last Accepted
	PromiseOutChan chan<- Promise
	LearnOutChan   chan<- Learn
	stopChan       chan int
	prepareChan    chan Prepare
	acceptChan     chan Accept
}

// NewAcceptor returns a new single-decree Paxos acceptor.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos acceptor.
//
// promiseOut: A send only channel used to send promises to other nodes.
//
// learnOut: A send only channel used to send learns to other nodes.
func NewAcceptor(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor {
	//TODO(student): Task 2 and 3 - algorithm and distributed implementation
	//Set of Proposers
	//Proposers := []NodeIDs
	//Set of Learners
	//Learners := []NodeIDs
	return &Acceptor{
		ID:             id,
		HighestRound:   NoRound,
		VotedRound:     NoRound,
		VotedValue:     ZeroValue,
		PromiseOutChan: promiseOut,
		LearnOutChan:   learnOut,
		stopChan:       make(chan int),
		prepareChan:    make(chan Prepare),
		acceptChan:     make(chan Accept),
	}
}

// Start starts a's main run loop as a separate goroutine. The main run loop
// handles incoming prepare and accept messages.
func (a *Acceptor) Start() {
	go func() {
		for {
			//TODO(student): Task 3 - distributed implementation
			select {
			case prpMsg := <-a.prepareChan: //TODO: Create a channel to recvie pepare messages from other nodes
				if prmMsg, sendMsg := a.handlePrepare(prpMsg); sendMsg == true {
					a.PromiseOutChan <- prmMsg
				}
			case accMsg := <-a.acceptChan: //TODO: Create a channel to receive accept messages from other nodes
				if lrnMsg, sendMsg := a.handleAccept(accMsg); sendMsg == true {
					a.LearnOutChan <- lrnMsg
				}
			case <-a.stopChan:
				break
			}
		}
	}()
}

// Stop stops a's main run loop.
func (a *Acceptor) Stop() {
	//TODO(student): Task 3 - distributed implementation
	a.stopChan <- 0
}

// DeliverPrepare delivers prepare prp to acceptor a.
func (a *Acceptor) DeliverPrepare(prp Prepare) {
	//TODO(student): Task 3 - distributed implementation
	a.prepareChan <- prp
}

// DeliverAccept delivers accept acc to acceptor a.
func (a *Acceptor) DeliverAccept(acc Accept) {
	//TODO(student): Task 3 - distributed implementation
	a.acceptChan <- acc
}

// Internal: handlePrepare processes prepare prp according to the single-decree
// Paxos algorithm. If handling the prepare results in acceptor a emitting a
// corresponding promise, then output will be true and prm contain the promise.
// If handlePrepare returns false as output, then prm will be a zero-valued
// struct.
func (a *Acceptor) handlePrepare(prp Prepare) (prm Promise, output bool) {
	//TODO(student): Task 2 - algorithm implementation
	if prp.Crnd > a.HighestRound {
		a.HighestRound = prp.Crnd
		return Promise{To: prp.From, From: a.ID, Rnd: a.HighestRound, Vrnd: a.VotedRound, Vval: a.VotedValue}, true //promise(rnd,vrnd,vval)
	}
	return Promise{}, false

}

// Internal: handleAccept processes accept acc according to the single-decree
// Paxos algorithm. If handling the accept results in acceptor a emitting a
// corresponding learn, then output will be true and lrn contain the learn.  If
// handleAccept returns false as output, then lrn will be a zero-valued struct.
func (a *Acceptor) handleAccept(acc Accept) (lrn Learn, output bool) {
	//TODO(student): Task 2 - algorithm implementation

	if acc.Rnd >= a.HighestRound && acc.Rnd != a.VotedRound { //on <ACCEPT, n, v> with n >= rnd ^ n != vrnd from proposer c
		a.HighestRound = acc.Rnd
		a.VotedRound = acc.Rnd
		a.VotedValue = acc.Val
		return Learn{From: a.ID, Rnd: a.VotedRound, Val: a.VotedValue}, true //learn(rnd,vval)
	}
	return Learn{}, false
}

//TODO(student): Add any other unexported methods needed.
