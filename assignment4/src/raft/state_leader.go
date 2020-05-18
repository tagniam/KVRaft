package raft

import (
	"sync"
	"time"
)

type AppendEntriesMessage struct {
	Args  AppendEntriesArgs
	Reply AppendEntriesReply
}

type Leader struct {
	mu         sync.Mutex
	nextIndex  []int
	matchIndex []int
	done       chan struct{}

	messages chan AppendEntriesMessage
}

func (l *Leader) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (l *Leader) Kill(rf *Raft) {
	close(l.done)
	DPrintf("%d (leader)    (term %d): killed", rf.me, rf.currentTerm)
}

func (l *Leader) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	if args.Term > rf.currentTerm {
		DPrintf("%d (leader)    (term %d): found AppendEntries with higher term from %d: converting to follower", rf.me, rf.currentTerm, args.LeaderId)
		rf.SetState(NewFollower(rf))
		rf.state.AppendEntries(rf, args, reply)
	}
}

func (l *Leader) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	if args.Term > rf.currentTerm {
		DPrintf("%d (leader)    (term %d): found RequestVote with higher term from %d: converting to follower", rf.me, rf.currentTerm, args.CandidateId)
		rf.SetState(NewFollower(rf))
		rf.state.RequestVote(rf, args, reply)
	}
}

func (l *Leader) HandleAppendEntries(rf *Raft, server int) {

	for {
		// Send append entries
		rf.mu.Lock()
		l.mu.Lock()

		args := AppendEntriesArgs{}
		args.Term = rf.currentTerm
		args.LeaderId = rf.me
		args.PrevLogIndex = 0
		args.PrevLogTerm = -1

		l.mu.Unlock()
		rf.mu.Unlock()

		DPrintf("%d (leader)    (term %d): sending AppendEntries to %d: %+v", rf.me, rf.currentTerm, server, args)
		go func(server int, args AppendEntriesArgs) {
			var reply AppendEntriesReply
			ok := rf.sendAppendEntries(server, args, &reply)
			if ok {
				DPrintf("%d (leader)    (term %d): received %v AppendEntries reply from %d: %+v", rf.me, rf.currentTerm, reply.Success, server, reply)
				select {
				case <-l.done:
				default:
					l.messages <- AppendEntriesMessage{args, reply}
				}
			} else {
				DPrintf("%d (leader)    (term %d): could not send AppendEntries to %d: %+v", rf.me, rf.currentTerm, server, args)
			}
		}(server, args)
		select {
		case <-l.done:
			return
		case <-time.After(rf.timeout):
		}
	}
}

func (l *Leader) Wait(rf *Raft) {
	for {
		select {
		case <-l.done:
			DPrintf("%d (leader)    (term %d): manually closing Wait", rf.me, rf.currentTerm)
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

	DPrintf("%d (leader)    (term %d): created new leader", rf.me, rf.currentTerm)
	return &l
}
