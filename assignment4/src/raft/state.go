package raft

type State interface {
	Start(*Raft, interface{})
	Kill(*Raft)

	AppendEntries(*Raft, AppendEntriesArgs, *AppendEntriesReply)
	RequestVote(*Raft, RequestVoteArgs, *RequestVoteReply)
}

type Follower struct {
}

func (f *Follower) Start(rf *Raft, command interface{}) {
	panic("implement me")
}

func (f *Follower) Kill(rf *Raft) {
	panic("implement me")
}

func (f *Follower) AppendEntries(rf *Raft, args AppendEntriesArgs, reply *AppendEntriesReply) {
	panic("implement me")
}

func (f *Follower) RequestVote(rf *Raft, args RequestVoteArgs, reply *RequestVoteReply) {
	panic("implement me")
}

func NewFollower(rf *Raft) State {
	f := Follower{}
	rf.votedFor = -1
	return &f
}

type Candidate struct {
}

type Leader struct {

}
