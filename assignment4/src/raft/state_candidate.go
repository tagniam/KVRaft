package raft

import (
	"sync"
	"time"
)

type Candidate struct {
	wg sync.WaitGroup
	done []chan struct{}
}

func (c *Candidate) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (c *Candidate) Kill(rf *Raft) {
	// Kill all running goroutines
	for server := range rf.peers {
		close(c.done[server])
	}
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
	ok := make(chan bool)
	go func() {
		ok <- rf.sendRequestVote(server, args, &reply)
	}()

	select {
	case sent := <-ok:
		if sent && reply.VoteGranted {
			DPrintf("%d (candidate): received yes RequestVote from %d: %+v", rf.me, server, reply)
			c.wg.Done()
		} else if sent && !reply.VoteGranted {
			DPrintf("%d (candidate): received no RequestVote from %d: %+v", rf.me, server, reply)
		} else {
			DPrintf("%d (candidate): could not send RequestVote to %d", rf.me, server)
		}
	case <-c.done[server]:
	}
}

//
// Wait for either the election to be won or the timeout to go off.
//
func (c *Candidate) Wait(rf *Raft) {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
		case <-done:
			DPrintf("%d (candidate): new leader bro", rf.me)
			c.Kill(rf)
		case <-time.After(rf.timeout):
			DPrintf("%d (candidate): timed out", rf.me)
			c.Kill(rf)
			rf.SetState(NewCandidate(rf))
	}
}

func NewCandidate(rf *Raft) State {
	c := Candidate{}
	c.wg.Add(len(rf.peers) / 2 + 1)
	c.done = make([]chan struct{}, len(rf.peers))
	for i := range c.done {
		c.done[i] = make(chan struct{})
	}

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

	for server := range rf.peers {
		if server == rf.me {
			continue
		}

		go c.HandleRequestVote(rf, server, args)
	}

	go c.Wait(rf)

	DPrintf("%d (candidate): created new candidate", rf.me)
	return &c
}
