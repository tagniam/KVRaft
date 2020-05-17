package raft

import "time"

type Follower struct {
	done chan struct{}
	heartbeat chan bool
}

func (f *Follower) Start(rf *Raft, command interface{}) (int, int, bool) {
	panic("implement me")
}

func (f *Follower) Kill(rf *Raft) {
	close(f.done)
	DPrintf("%d (follower): killed", rf.me)
}

func (f *Follower) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	reply.Term = rf.currentTerm
	insufficientTerm := args.Term < rf.currentTerm
	logMatches := rf.log.Contains(args.PrevLogIndex, args.PrevLogTerm)
	if insufficientTerm || !logMatches {
		reply.Success = false
		DPrintf("%d (follower): rejected AppendEntries request from %d\n", rf.me, args.LeaderId)
		return
	}

	reply.Success = true
	DPrintf("%d (follower): accepted AppendEntries request from %d\n", rf.me, args.LeaderId)
	f.heartbeat <- true
}

func (f *Follower) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	// Reject vote if candidate's term is less than current term
	// Accept vote if `votedFor` is null (-1 in this case) or args.candidateId and our log isn't more up to date than
	// candidate's log
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("%d (follower): received RequestVote call from %d: %+v\n", rf.me, args.CandidateId, args)
	reply.Term = rf.currentTerm

	insufficientTerm := args.Term < rf.currentTerm
	alreadyVoted := rf.votedFor != -1 && rf.votedFor != args.CandidateId
	moreUpToDate := rf.log.Compare(args.LastLogIndex, args.LastLogTerm) > 0
	if insufficientTerm || alreadyVoted || moreUpToDate {
		reply.VoteGranted = false
		DPrintf("%d (follower) rejected RequestVote from %d", rf.me, args.CandidateId)
		DPrintf("%d (follower): insufficientTerm = %v", rf.me, insufficientTerm)
		DPrintf("%d (follower) alreadyVoted = %v ", rf.me, alreadyVoted)
		DPrintf("%d (follower) moreUpToDate = %v", rf.me, moreUpToDate)
		return
	}

	DPrintf("%d (follower) accepted RequestVote from %d", rf.me, args.CandidateId)
	reply.VoteGranted = true

	rf.votedFor = args.CandidateId
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
			rf.SetState(NewCandidate(rf))
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

	DPrintf("%d (follower): created new follower", rf.me)
	return &f
}
