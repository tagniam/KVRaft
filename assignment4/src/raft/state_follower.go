package raft

import "time"

type Follower struct {
	done chan bool
	heartbeat chan bool
}

func (f *Follower) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (f *Follower) Kill(rf *Raft) {
	f.done <- true
	DPrintf("%d (follower): killed", rf.me)
}

func (f *Follower) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	f.heartbeat <- true
}

func (f *Follower) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	panic("implement me")
}

func (f *Follower) Wait(rf *Raft) {
	for {
		select {
		case <-f.heartbeat:
			DPrintf("%d (follower): received heartbeat, resetting timeout", rf.me)
		case <-f.done:
			DPrintf("%d (follower): manually closing Wait", rf.me)
			return
		case <-time.After(rf.timeout):
			DPrintf("%d (follower): timed out", rf.me)
			// rf.state = NewCandidate(rf)
			return
		}
	}
}

func NewFollower(rf *Raft) State {
	f := Follower{}
	rf.votedFor = -1
	go f.Wait(rf)

	DPrintf("%d (follower): created new follower", rf.me)
	return &f
}
