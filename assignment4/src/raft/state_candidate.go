package raft

import (
	"time"
)

type Candidate struct {
	votes chan bool
	done chan struct{}
}

func (c *Candidate) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (c *Candidate) Kill(rf *Raft) {
	close(c.done)
	DPrintf("%d (candidate): killed", rf.me)
}

func (c *Candidate) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	if args.Term >= rf.currentTerm {
		DPrintf("%d (candidate): found AppendEntries with equal or higher term from %d: converting to follower", rf.me, args.LeaderId)
		rf.SetState(NewFollower(rf))
		rf.state.AppendEntries(rf, args, reply)
	}
}

func (c *Candidate) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	if args.Term > rf.currentTerm {
		DPrintf("%d (candidate): found RequestVote with higher term from %d: converting to follower", rf.me, args.CandidateId)
		rf.SetState(NewFollower(rf))
		rf.state.RequestVote(rf, args, reply)
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
}

func (c *Candidate) HandleRequestVote(rf *Raft, server int, args RequestVoteArgs) {
	var reply RequestVoteReply
	DPrintf("%d (candidate): sending RequestVote to %d: %+v", rf.me, server, args)
	ok := rf.sendRequestVote(server, args, &reply)

	if ok && reply.VoteGranted {
		DPrintf("%d (candidate): received yes RequestVote from %d: %+v", rf.me, server, reply)
		c.votes <- true
	} else if ok && !reply.VoteGranted {
		DPrintf("%d (candidate): received no RequestVote from %d: %+v", rf.me, server, reply)
	} else {
		DPrintf("%d (candidate): could not send RequestVote to %d", rf.me, server)
	}
}

//
// Wait for either the election to be won or the timeout to go off.
//
func (c *Candidate) Wait(rf *Raft) {
	needed := len(rf.peers) / 2 + 1

	timeout := make(chan struct{})
	go func() {
		<-time.After(rf.timeout)
		close(timeout)
	}()

	for {
		select {
		case <-timeout:
			DPrintf("%d (candidate): timed out", rf.me)
			rf.SetState(NewCandidate(rf))
			return
		case <-c.done:
			DPrintf("%d (candidate): stopped waiting", rf.me)
			return
		case <-c.votes:
			needed--
			if needed == 0 {
				DPrintf("%d (candidate): won election", rf.me)
				rf.SetState(NewLeader(rf))
				return
			}
		}
	}
}

func NewCandidate(rf *Raft) State {
	c := Candidate{}
	c.done = make(chan struct{})
	c.votes = make(chan bool, 1)
	c.votes <- true

	// Send RequestVote RPCs to all peers
	rf.mu.Lock()
	rf.votedFor = rf.me
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
