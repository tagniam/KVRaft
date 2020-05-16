package raft

import (
	"context"
	"time"
)

type Leader struct {
	context context.Context
	cancel context.CancelFunc

	nextIndex []int
	matchIndex []int
}

func (l *Leader) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (l *Leader) Kill(rf *Raft) {
	l.cancel()
	DPrintf("%d (candidate): killed", rf.me)
}

func (l *Leader) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	// panic("implement me")
}

func (l *Leader) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	//panic("implement me")
}

func (l *Leader) HandleHeartBeat (rf *Raft, server int) {
	args := AppendEntriesArgs{}
	args.LeaderId = rf.me

	for {
		// Send append entries
		var reply AppendEntriesReply
		DPrintf("%d (leader): sending AppendEntries to %d: %+v", rf.me, server, args)
		ok := rf.sendAppendEntries(server, args, &reply)

		select {
		case <-l.context.Done():
			return
		default:
			DPrintf("%v", ok)
		}
		<-time.After(rf.timeout)
	}
}

func NewLeader(rf *Raft) State {
	l := Leader{}

	l.nextIndex = make([]int, len(rf.peers))
	l.matchIndex = make([]int, len(rf.peers))
	for server := range rf.peers {
		l.nextIndex[server] = rf.log.GetLastLogIndex() + 1
		l.matchIndex[server] = 0
	}
	l.context, l.cancel = context.WithCancel(context.Background())

	for server := range rf.peers {
		go l.HandleHeartBeat(rf, server)
	}

	DPrintf("%d (leader): created new leader", rf.me)
	return &l
}
