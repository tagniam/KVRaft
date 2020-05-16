package raft

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

	}
}

func NewCandidate(rf *Raft) State {
	c := Candidate{}

	// Send RequestVote RPCs to all peers
	rf.mu.Lock()

	rf.currentTerm++
	term := rf.currentTerm
	lastLogIndex := rf.log.GetLastLogIndex()
	lastLogTerm := rf.log.GetLastLogTerm()

	rf.mu.Unlock()

	args := RequestVoteArgs{
		Term:         term,
		CandidateId:  rf.me,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	for server, _ := range rf.peers {
		if server == rf.me {
			continue
		}

		go c.HandleRequestVote(rf, server, args)
	}

	DPrintf("%d (candidate): created new candidate", rf.me)
	return &c
}
