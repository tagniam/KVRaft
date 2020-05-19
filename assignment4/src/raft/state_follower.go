package raft

import "time"

type Follower struct {
	done      chan struct{}
	heartbeat chan bool
}

func (f *Follower) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (f *Follower) Kill(rf *Raft) {
	close(f.done)
	DPrintf("%d (follower)  (term %d): killed", rf.me, rf.currentTerm)
}

func (f *Follower) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	DPrintf("%d (follower)  (term %d): received AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
	}

	reply.Term = rf.currentTerm
	insufficientTerm := args.Term < rf.currentTerm
	logMatches := rf.log.Contains(args.PrevLogIndex, args.PrevLogTerm)
	if insufficientTerm || !logMatches {
		reply.Success = false
		DPrintf("%d (follower)  (term %d): rejected AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)
		return
	}

	reply.Success = true
	DPrintf("%d (follower)  (term %d): accepted AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)
	select {
	case <-f.done:
	default:
		f.heartbeat <- true
	}
}

func (f *Follower) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	// Reject vote if candidate's term is less than current term
	// Accept vote if `votedFor` is null (-1 in this case) or args.candidateId and our log isn't more up to date than
	// candidate's log
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
	}

	DPrintf("%d (follower)  (term %d): received RequestVote call from %d: %+v\n", rf.me, rf.currentTerm, args.CandidateId, args)
	reply.Term = rf.currentTerm

	insufficientTerm := args.Term < rf.currentTerm
	alreadyVoted := rf.votedFor != -1 && rf.votedFor != args.CandidateId
	moreUpToDate := rf.log.Compare(args.LastLogIndex, args.LastLogTerm) > 0
	if insufficientTerm || alreadyVoted || moreUpToDate {
		reply.VoteGranted = false
		DPrintf("%d (follower)  (term %d) rejected RequestVote from %d", rf.me, rf.currentTerm, args.CandidateId)
		DPrintf("%d (follower)  (term %d): insufficientTerm = %v", rf.me, rf.currentTerm, insufficientTerm)
		DPrintf("%d (follower)  (term %d) alreadyVoted = %v ", rf.me, rf.currentTerm, alreadyVoted)
		DPrintf("%d (follower)  (term %d) moreUpToDate = %v", rf.me, rf.currentTerm, moreUpToDate)
		return
	}

	DPrintf("%d (follower)  (term %d) accepted RequestVote from %d", rf.me, rf.currentTerm, args.CandidateId)
	reply.VoteGranted = true

	rf.votedFor = args.CandidateId
	select {
	case <-f.done:
	default:
		f.heartbeat <- true
	}
}

func (f *Follower) Wait(rf *Raft) {
	for {
		select {
		case <-f.heartbeat:
			DPrintf("%d (follower)  (term %d): received heartbeat, resetting timeout", rf.me, rf.currentTerm)
		case <-f.done:
			DPrintf("%d (follower)  (term %d): manually closing Wait", rf.me, rf.currentTerm)
			return
		case <-time.After(rf.timeout):
			DPrintf("%d (follower)  (term %d): timed out", rf.me, rf.currentTerm)
			rf.mu.Lock()
			rf.SetState(NewCandidate(rf))
			rf.mu.Unlock()
			return
		}
	}
}

func NewFollower(rf *Raft) State {
	f := Follower{}
	f.done = make(chan struct{})
	f.heartbeat = make(chan bool, 1)
	rf.votedFor = -1
	go f.Wait(rf)

	DPrintf("%d (follower)  (term %d): created new follower", rf.me, rf.currentTerm)
	return &f
}
