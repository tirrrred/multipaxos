// +build !solution

package detector

// A MonLeaderDetector represents a Monarchical Eventual Leader Detector as
// described at page 53 in:
// Christian Cachin, Rachid Guerraoui, and Luís Rodrigues: "Introduction to
// Reliable and Secure Distributed Programming" Springer, 2nd edition, 2011.
type MonLeaderDetector struct {
	// TODO(student): Add needed fields
	nodes          []int //The Node ID of the leader
	suspectedNodes []int //slice with suspected nodes
	leaderNode     int   //ID of the current leader node
}

// NewMonLeaderDetector returns a new Monarchical Eventual Leader Detector
// given a list of node ids.
func NewMonLeaderDetector(nodeIDs []int) *MonLeaderDetector {
	m := &MonLeaderDetector{
		nodes:      nodeIDs,   // What if the there is no values (nil) in the nodeIDs slice?
		leaderNode: UnknownID, // Sets the leaderNode entry to "UnkownID" which is a constant int = -1 from defs.go
	}

	//m.LeaderChange()

	return m
}

// Leader returns the current leader. Leader will return UnknownID if all nodes
// are suspected.
func (m *MonLeaderDetector) Leader() int {
	// TODO(student): Implement
	for _, nodeID := range m.nodes {
		suspected := false
		for _, sNode := range m.suspectedNodes {
			if nodeID == sNode {
				suspected = true
				break //Assumes it is only uniqe node IDs in cluster
			}
		}
		if suspected == false && nodeID > m.leaderNode {
			m.leaderNode = nodeID
		}
	}
	return m.leaderNode
}

// Suspect instructs m to consider the node with matching id as suspected. If
// the suspect indication result in a leader change the leader detector should
// this publish this change its subscribers.
func (m *MonLeaderDetector) Suspect(id int) {
	// TODO(student): Implement
	m.suspectedNodes = append(m.suspectedNodes, id)
}

// Restore instructs m to consider the node with matching id as restored. If
// the restore indication result in a leader change the leader detector should
// this publish this change its subscribers.
func (m *MonLeaderDetector) Restore(id int) {
	// TODO(student): Implement
	for i, sID := range m.suspectedNodes { //Iterate through all suspected Node IDs in m.suspectedNodes
		if sID == id { //If the suspected node id is equal to the given node id to restore do...
			m.suspectedNodes = append(m.suspectedNodes[:i], m.suspectedNodes[i+1:]...) //Modify m.suspectedNodes slice. Removes the entry with the resotred node ID at index i
		}
	}
}

// Subscribe returns a buffered channel that m on leader change will use to
// publish the id of the highest ranking node. The leader detector will publish
// UnknownID if all nodes become suspected. Subscribe will drop publications to
// slow subscribers. Note: Subscribe returns a unique channel to every
// subscriber; it is not meant to be shared.
func (m *MonLeaderDetector) Subscribe() <-chan int {
	// TODO(student): Implement
	//c := make(chan int)
	//c <- m.leaderNode
	//return c
	return nil
}

// LeaderChange is used to change leaderNode in the struct MonLeaderDetector
func (m *MonLeaderDetector) LeaderChange() bool {
	for _, nodeID := range m.nodes {
		suspected := false
		for _, sNode := range m.suspectedNodes {
			if nodeID == sNode {
				suspected = true
				//break
			}
		}
		if suspected == false && nodeID > m.leaderNode {
			m.leaderNode = nodeID
		}
	}
	for _, nodeID := range m.nodes {

	}
	if newLeader != m.leaderNode {
		return true
	} else {
		return false
	}
}

// TODO(student): Add other unexported functions or methods if needed.
