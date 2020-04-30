// +build !solution

package multipaxos

import (
	"container/list"
	"dat520/lab3/detector"
	"sort"
	"time"
)

// Proposer represents a proposer as defined by the Multi-Paxos algorithm.
type Proposer struct {
	id     int
	quorum int
	n      int

	crnd     Round
	adu      SlotID
	nextSlot SlotID

	promises     []*Promise
	promiseCount int
	nonNILprm    []*Promise

	phaseOneDone           bool
	phaseOneProgressTicker *time.Ticker

	acceptsOut *list.List
	requestsIn *list.List

	ld     detector.LeaderDetector
	leader int

	prepareOut chan<- Prepare
	acceptOut  chan<- Accept
	promiseIn  chan Promise
	cvalIn     chan Value

	incDcd chan struct{}
	stop   chan struct{}
}

//Used to map PromiseSlot and deceide the which Vval should be in []Accept slice if there are multiple PromiseSlot from same SlotID
type votedTuple struct {
	Vrnd Round
	Vval Value
}

// NewProposer returns a new Multi-Paxos proposer. It takes the following
// arguments:
//
// id: The id of the node running this instance of a Paxos proposer.
//
// nrOfNodes: The total number of Paxos nodes.
//
// adu: all-decided-up-to. The initial id of the highest _consecutive_ slot
// that has been decided. Should normally be set to -1 initially, but for
// testing purposes it is passed in the constructor.
//
// ld: A leader detector implementing the detector.LeaderDetector interface.
//
// prepareOut: A send only channel used to send prepares to other nodes.
//
// The proposer's internal crnd field should initially be set to the value of
// its id.
func NewProposer(id, nrOfNodes, adu int, ld detector.LeaderDetector, prepareOut chan<- Prepare, acceptOut chan<- Accept) *Proposer {
	return &Proposer{
		id:     id,
		quorum: (nrOfNodes / 2) + 1,
		n:      nrOfNodes,

		crnd:     Round(id),
		adu:      SlotID(adu),
		nextSlot: 0,

		promises:  make([]*Promise, nrOfNodes),
		nonNILprm: make([]*Promise, 0),

		phaseOneProgressTicker: time.NewTicker(time.Second),

		acceptsOut: list.New(),
		requestsIn: list.New(),

		ld:     ld,
		leader: ld.Leader(),

		prepareOut: prepareOut,
		acceptOut:  acceptOut,
		promiseIn:  make(chan Promise, 3000000000000000),
		cvalIn:     make(chan Value, 3000000000000000),

		incDcd: make(chan struct{}),
		stop:   make(chan struct{}),
	}

}

// Start starts p's main run loop as a separate goroutine.
func (p *Proposer) Start() {
	trustMsgs := p.ld.Subscribe()
	go func() {
		for {
			select {
			case prm := <-p.promiseIn:
				accepts, output := p.handlePromise(prm)
				if !output {
					continue
				}
				p.nextSlot = p.adu + 1
				p.acceptsOut.Init()
				p.phaseOneDone = true
				for _, acc := range accepts {
					p.acceptsOut.PushBack(acc)
				}
				p.sendAccept()
			case cval := <-p.cvalIn:
				if p.id != p.leader {
					continue
				}
				p.requestsIn.PushBack(cval)
				if !p.phaseOneDone {
					continue
				}
				p.sendAccept()
			case <-p.incDcd:
				p.adu++
				if p.id != p.leader {
					continue
				}
				if !p.phaseOneDone {
					continue
				}
				p.sendAccept()
			case <-p.phaseOneProgressTicker.C:
				if p.id == p.leader && !p.phaseOneDone {
					p.startPhaseOne()
				}
			case leader := <-trustMsgs:
				p.leader = leader
				if leader == p.id {
					p.startPhaseOne()
				}
			case <-p.stop:
				return
			}
		}
	}()
}

// Stop stops p's main run loop.
func (p *Proposer) Stop() {
	p.stop <- struct{}{}
}

// DeliverPromise delivers promise prm to proposer p.
func (p *Proposer) DeliverPromise(prm Promise) {
	p.promiseIn <- prm
}

// DeliverClientValue delivers client value cval from to proposer p.
func (p *Proposer) DeliverClientValue(cval Value) {
	p.cvalIn <- cval
}

// IncrementAllDecidedUpTo increments the Proposer's adu variable by one.
func (p *Proposer) IncrementAllDecidedUpTo() {
	p.incDcd <- struct{}{}
}

// Internal: handlePromise processes promise prm according to the Multi-Paxos
// algorithm. If handling the promise results in proposer p emitting a
// corresponding accept slice, then output will be true and accs contain the
// accept messages. If handlePromise returns false as output, then accs will be
// a nil slice. See the Lab 5 text for a more complete specification.
func (p *Proposer) handlePromise(prm Promise) (accs []Accept, output bool) {
	// TODO(student)
	// type Promise struct{To int; From int; Rnd Round; Slots []PromiseSlot}
	//Ignore promises if prm.Rnd dosent match our current round (crnd)
	if prm.Rnd != p.crnd {
		return nil, false
	}

	//Ignore promises from same node on the current round
	for _, rPrm := range p.promises {
		if rPrm == nil {
			continue
		}
		if rPrm.Rnd == prm.Rnd && rPrm.From == prm.From {
			return nil, false
		}
	}

	//Append promise to slice of promises
	p.promises = append(p.promises, &prm)
	p.nonNILprm = nil
	for _, rPrm := range p.promises {
		if rPrm != nil {
			p.nonNILprm = append(p.nonNILprm, rPrm)
		}
	}
	//If slice of promises is less than quorom == not majority/quorom == ignore
	if len(p.nonNILprm) < p.quorum {
		return nil, false
	}
	//quorum/majority!!

	//type Accept struct{From int; Slot SlotID; Rnd Round; Val Value}
	accsUnsorted := p.purgePrmSlots()
	if len(accsUnsorted) == 0 {
		return []Accept{}, true
	}
	accsSorted := p.sortAndFillGap(accsUnsorted)

	return accsSorted, true
}

// Internal: increaseCrnd increases proposer p's crnd field by the total number
// of Paxos nodes.
func (p *Proposer) increaseCrnd() {
	p.crnd = p.crnd + Round(p.n)
}

// Internal: startPhaseOne resets all Phase One data, increases the Proposer's
// crnd and sends a new Prepare with Slot as the current adu.
func (p *Proposer) startPhaseOne() {
	p.phaseOneDone = false
	p.promises = make([]*Promise, p.n)
	p.increaseCrnd()
	p.prepareOut <- Prepare{From: p.id, Slot: p.adu, Crnd: p.crnd}
}

// Internal: sendAccept sends an accept from either the accept out queue
// (generated after Phase One) if not empty, or, it generates an accept using a
// value from the client request queue if not empty. It does not send an accept
// if the previous slot has not been decided yet.
func (p *Proposer) sendAccept() {
	const alpha = 1
	if !(p.nextSlot <= p.adu+alpha) {
		// We must wait for the next slot to be decided before we can
		// send an accept.
		//
		// For Lab 6: Alpha has a value of one here. If you later
		// implement pipelining then alpha should be extracted to a
		// proposer variable (alpha) and this function should have an
		// outer for loop.
		return
	}

	// Pri 1: If bounded by any accepts from Phase One -> send previously
	// generated accept and return.
	if p.acceptsOut.Len() > 0 {
		acc := p.acceptsOut.Front().Value.(Accept)
		p.acceptsOut.Remove(p.acceptsOut.Front())
		p.acceptOut <- acc
		p.nextSlot++
		return
	}

	// Pri 2: If any client request in queue -> generate and send
	// accept.
	if p.requestsIn.Len() > 0 {
		cval := p.requestsIn.Front().Value.(Value)
		p.requestsIn.Remove(p.requestsIn.Front())
		acc := Accept{
			From: p.id,
			Slot: p.nextSlot,
			Rnd:  p.crnd,
			Val:  cval,
		}
		p.nextSlot++
		p.acceptOut <- acc
	}
}

//purgePrmSlots iterates through all Promise messages and returns a slice with Accept messages []Accepts
//NOTE: It does not sort the Accept messages
func (p *Proposer) purgePrmSlots() []Accept {
	//type Promise struct{To int; From int; Rnd Round; Slots []PromiseSlot}

	//map to store PromiseSlot, with the SlotID as the key, and the Vrnd and Vval (in a tuple) as value
	ps := map[SlotID]votedTuple{}
	//iterate over all promise msg
	for _, rPrm := range p.promises {
		if rPrm == nil {
			continue
		}
		//iterate over all PromiseSlot in the promise msg
		for _, slot := range rPrm.Slots {
			if slot.ID <= p.adu {
				continue
			}
			//Checks if the slot ID is present in the ps map
			if vT, ok := ps[slot.ID]; ok {
				//fmt.Printf("Existing slot ID %d: Vrnd: %v, Vval: %v\n", slot.ID, vT.Vrnd, vT.Vval)
				//If slot ID is present, check if the existing PromiseSlot Vrnd is bigger or equal
				//fmt.Printf("Existing: (ID: %d - Vrnd: %d) New: (ID: %d - Vrnd: %d)", slot.ID, vT.Vrnd)
				if vT.Vrnd < slot.Vrnd {
					var newVT votedTuple
					newVT.Vrnd = slot.Vrnd
					newVT.Vval = slot.Vval
					ps[slot.ID] = newVT
				}
			} else {
				//fmt.Printf("Not existing slot ID %d: Vrnd: %v, Vval: %v\n", slot.ID, vT.Vrnd, vT.Vval)
				var newVT votedTuple
				newVT.Vrnd = slot.Vrnd
				newVT.Vval = slot.Vval
				ps[slot.ID] = newVT
			}
		}
	}
	//type Accept struct{From int; Slot SlotID; Rnd Round; Val Value}
	accSlice := []Accept{}
	for k, v := range ps {
		accSlice = append(accSlice, Accept{From: p.id, Slot: k, Rnd: p.crnd, Val: v.Vval})
	}
	return accSlice
}

//sortAndFIllGap takes a slice of Accept msg and sorts it and fills gaps with no-op values
func (p *Proposer) sortAndFillGap(accSlice []Accept) []Accept {
	sort.SliceStable(accSlice, func(i, j int) bool {
		return accSlice[i].Slot < accSlice[j].Slot
	})

	//minSlotID := accSlice[0].Slot
	//maxSlotID := accSlice[len(accSlice)].Slot

	for i, acc := range accSlice {
		if i+1 == len(accSlice) {
			break
		}
		if acc.Slot+1 != accSlice[i+1].Slot {
			var diff SlotID
			diff = accSlice[i+1].Slot - acc.Slot
			for j := 1; SlotID(j) < diff; j++ {
				noopSlot := acc.Slot + SlotID(j)
				noopAcc := Accept{
					From: p.id,
					Slot: noopSlot, //j+1
					Rnd:  p.crnd,
					Val: Value{
						Noop: true,
					},
				}
				accSlice = append(accSlice, noopAcc)
			}
		}
	}

	sort.SliceStable(accSlice, func(i, j int) bool {
		return accSlice[i].Slot < accSlice[j].Slot
	})

	return accSlice
}
