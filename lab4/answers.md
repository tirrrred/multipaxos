## Answers to Paxos Questions 

You should write down your answers to the
[questions](https://dat520.github.io/r/?assignments/tree/master/lab4-singlepaxos#questions-10)
for Lab 4 in this file. 

1. Is it possible that Paxos enters an infinite loop? Explain.
There is a possibility that Paxos can enter an infinite loop if different proposers sends higher rounds, so they are "competing" about having the highest round. The acceptors will get different prepare messages from different nodes with different rounds than. This can be solved with a leader detection and a designated leader which is the only one to send prepare messages and thereby receive promise messages. I.e this assumes a working leader detection

2. Is the value to agree on included in the Prepare message?
No, prepare only have (From, Round). 

3. Does Paxos rely on an increasing proposal/round number in order to work? Explain.
Yes, because Paxos is a state machine algorithm it needs to know which "state" each node is in and transition the nodes in the correct order to the current state.

4. Look at this description for Phase 1B: If the proposal number N is larger than any previous proposal, then each Acceptor promises not to accept proposals less than N, and sends the value it last accepted for this instance to the Proposer. What is meant by “the value it last accepted”? And what is an “instance” in this case?
The "value last accepted" is the last value an acceptor received in a ACCEPT message from proposer. In single-decree paxos this value cannot be changed after consensus is reached.
An "instance" is a node/process running the same Paxos algorithm, meaning in the same cluster/network. 

5. Explain, with an example, what will happen if there are multiple proposers.
With multiple proposers, and no leader detection:
- Each proposer created unique round numbers
- When a proposer gets a new value from a client it will send a prepare with it's unique round. If it get's a promise from a majority of the acceptors it will send an accept with either
  1) The client value if not any last accepted value are received from the acceptors
  2) Received a last accepted value from acceptors and chooses this value. This is also communicated to the client. If there are different "last accepted values" the value with the higest round is choosen
- If Proposer 1 (P1) send a Prepare message with rnd 1 (PRP(1)) and the majority of acceptors reply with a Promise message (PRM(1,-)), P1 will then send a Accept message (ACC(1, "val")). If no other promises to any higher round is made in the meantime, acceptors will accept the proposed value and it will be consensus.
- However if P2 sends a PRP(2) after P1 sent his PRP(1), BUT P1 sent his ACC(1, "val) message, the acceptors would have sent a new PRM(2,-) to P2, since the round is higher. P1 will still send his ACC(1, "val") in good faith, but it will be ignored by the acceptors as the round is lower (1<2). 
- P2 can then send his ACC(2, "val2") to the acceptors, which will be accepted and val2 is the consensus value.

6. What happens if two proposers both believe themselves to be the leader and send Prepare messages simultaneously?
As described above, there might be several more prepare/promise/accept messages between proposers and acceptors, but it will eventually work out. As long as the proposer have unique rounds and rounds a increased it will work out.

7. What can we say about system synchrony if there are multiple proposers (or leaders)?
The system will work in the same matter, only with additonal message sent between the nodes in the network. The algorithm is async and don't need global clock (NTP) to correlate messages between nodes. It also don't care about latency or jitter between nodes (over the network). If the network is unrealiable (alot of latency jitter) it might create additional message and also decrese the performance of the algorithm to create consensus and majority. 

8. Can an acceptor accept one value in round 1 and another value in round 2? Explain.
No, this is not possible. An acceptor that has promised to a proposer and received an accept from a proposer (without any higher prepare message disturbing this process) will send this accepted value in all future promise messages, as the "last accepted value". In this way, any future proposer trying to propose a new value, will get the last accepted value in return from the acceptors and needs to use this value. To make the algorithm agree upon an new value, the paxos algorithm needs to be run again, which is multi-paxos

9. What must happen for a value to be “chosen”? What is the connection between chosen values and learned values?
1) Propser gets a client value
2) Proposer sends a prepare AND gets promise message from a MAJORITY of acceptors
3) Proposer send accept message to all acceptors and if acceptors "choose" this value, they broadcast this chosen value
4) Learners will listen for choosen value/message (LRN)
5) If learner "learns" a value from a majority of nodes/acceptors
