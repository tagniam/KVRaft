package raft

import "time"

type Candidate struct {
}

func (c *Candidate) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (c *Candidate) Kill(rf *Raft) {
	panic("implement me")
}

func (c *Candidate) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	panic("implement me")
}

func (c *Candidate) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	panic("implement me")
}

func (c *Candidate) HandleRequestVote(rf *Raft, server int, args RequestVoteArgs) {
	var reply RequestVoteReply
	DPrintf("%d (candidate): sending RequestVote to %d: %+v", rf.me, server, args)
	if ok := rf.sendRequestVote(server, args, &reply); ok {
		if reply.VoteGranted {
			DPrintf("%d (candidate): received yes RequestVote from %d: %+v", rf.me, server, reply)
		} else {
			DPrintf("%d (candidate): received no RequestVote from %d: %+v", rf.me, server, reply)
		}
	}
}

//
// Wait for either the election to be won or the timeout to go off.
//
func (c *Candidate) Wait(rf *Raft) {
	select {
		// case <-c.election.won:
		case <-time.After(rf.timeout):
			DPrintf("%d (candidate): timed out", rf.me)
			c.Kill(rf)
			rf.SetState(NewCandidate(rf))
	}
}

func NewCandidate(rf *Raft) State {
	c := Candidate{}

	// Send RequestVote RPCs to all peers
	rf.mu.Lock()
	rf.currentTerm++
	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.log.GetLastLogIndex(),
		LastLogTerm:  rf.log.GetLastLogTerm(),
	}
	rf.mu.Unlock()

	for server, _ := range rf.peers {
		if server == rf.me {
			continue
		}

		go c.HandleRequestVote(rf, server, args)
	}

	go c.Wait(rf)

	DPrintf("%d (candidate): created new candidate", rf.me)
	return &c
}
