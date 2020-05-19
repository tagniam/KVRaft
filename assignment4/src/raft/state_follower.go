package raft

import "time"

type Follower struct {
	done      chan struct{}
	heartbeat chan bool
}

func (f *Follower) Start(rf *Raft, command interface{}) (int, int, bool) {
	DPrintf("%d (follower)  (term %d): Start(%v) called", rf.me, rf.currentTerm, command)
	return rf.log.GetLastLogIndex()+1, rf.currentTerm, false
}

func (f *Follower) Kill(rf *Raft) {
	close(f.done)
	DPrintf("%d (follower)  (term %d): killed", rf.me, rf.currentTerm)
	DPrintf("%d (follower)  (term %d): killed: log: %v", rf.me, rf.currentTerm, rf.log.Entries)
}

func (f *Follower) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	DPrintf("%d (follower)  (term %d): received AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
	}

	reply.Term = rf.currentTerm
	insufficientTerm := args.Term < rf.currentTerm
	logMatches := rf.log.Contains(args.PrevLogIndex, args.PrevLogTerm)

	// Reset heartbeat if not insufficient term (AppendEntries request comes from current leader)
	if !insufficientTerm {
		select {
		case <-f.done:
		default:
			f.heartbeat <- true
		}
	}

	if insufficientTerm || !logMatches {
		reply.Success = false
		DPrintf("%d (follower)  (term %d): rejected AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)
		return
	}

	reply.Success = true
	DPrintf("%d (follower)  (term %d): accepted AppendEntries request from %d\n", rf.me, rf.currentTerm, args.LeaderId)

	// if an existing entry conflicts with a new one(same index but different terms), delete the existing entry and all that follow it
	i := 0
	j := args.PrevLogIndex+1
	for i < len(args.Entries) && j < len(rf.log.Entries) && rf.log.Entries[j].Term == args.Entries[i].Term {
		i++
		j++
	}
	rf.log.Entries = rf.log.Entries[:j]
	// append any new entries not already in the log
	rf.log.Entries = append(rf.log.Entries, args.Entries[i:]...)
	// if leader commit > commit index, set commit index = min(leader commit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit < rf.log.GetLastLogIndex() {
			DPrintf("%d (follower)  (term %d): setting commit index from %d to %d", rf.me, rf.currentTerm, rf.commitIndex, args.LeaderCommit)
			rf.commitIndex = args.LeaderCommit
		} else {
			DPrintf("%d (follower)  (term %d): setting commit index from %d to %d", rf.me, rf.currentTerm, rf.commitIndex, rf.log.GetLastLogIndex())
			rf.commitIndex = rf.log.GetLastLogIndex()
		}
		rf.Commit()
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
