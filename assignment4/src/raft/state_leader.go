package raft

import (
	"time"
)

type AppendEntriesMessage struct {
	Args AppendEntriesArgs
	Reply AppendEntriesReply
}

type Leader struct {
	nextIndex []int
	matchIndex []int
	done chan struct{}

	messages chan AppendEntriesMessage
}

func (l *Leader) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (l *Leader) Kill(rf *Raft) {
	close(l.done)
	DPrintf("%d (leader): killed", rf.me)
}

func (l *Leader) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	if args.Term > rf.currentTerm {
		DPrintf("%d (leader): found AppendEntries with higher term from %d: converting to follower", rf.me, args.LeaderId)
		rf.SetState(NewFollower(rf))
		rf.state.AppendEntries(rf, args, reply)
	}
}

func (l *Leader) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	if args.Term > rf.currentTerm {
		DPrintf("%d (leader): found RequestVote with higher term from %d: converting to follower", rf.me, args.CandidateId)
		rf.SetState(NewFollower(rf))
		rf.state.RequestVote(rf, args, reply)
	}
}

func (l *Leader) HandleAppendEntries(rf *Raft, server int) {
	rf.mu.Lock()
	args := AppendEntriesArgs{}
	args.Term = rf.currentTerm
	args.LeaderId = rf.me
	args.PrevLogIndex = 0
	args.PrevLogTerm = -1
	rf.mu.Unlock()

	for {
		// Send append entries
		DPrintf("%d (leader): sending AppendEntries to %d: %+v", rf.me, server, args)
		go func(server int, args AppendEntriesArgs) {
			var reply AppendEntriesReply
			ok := rf.sendAppendEntries(server, args, &reply)
			if ok && reply.Success {
				DPrintf("%d (leader): received successful AppendEntries reply from %d: %+v", rf.me, server, reply)
				l.messages <- AppendEntriesMessage{args, reply}
			} else if ok && !reply.Success {
				DPrintf("%d (leader): received unsuccessful AppendEntries reply from %d: %+v", rf.me, server, reply)
				l.messages <- AppendEntriesMessage{args, reply}
			} else {
				DPrintf("%d (leader): could not send AppendEntries to %d: %+v", rf.me, server, args)
			}
		}(server, args)
		<-time.After(rf.timeout)
	}
}

func (l *Leader) Wait(rf *Raft) {
	for {
		select {
		case <-l.done:
			return
		default:
		}

		select {
		case msg := <-l.messages:
			if msg.Args.Term != rf.currentTerm {
				// Outdated message, ignore
				break
			}

		}
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
	l.done = make(chan struct{})
	l.messages = make(chan AppendEntriesMessage)

	for server := range rf.peers {
		if server == rf.me {
			continue
		}

		go l.HandleAppendEntries(rf, server)
	}

	go l.Wait(rf)

	DPrintf("%d (leader): created new leader", rf.me)
	return &l
}
