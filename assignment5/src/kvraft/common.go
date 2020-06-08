package raftkv

const (
	OK       = "OK"
	ErrNoKey = "ErrNoKey"
)

type Err string
type ClientID int64
type Sequence int

// Put or Append
type PutAppendArgs struct {
	// You'll have to add definitions here.
	Key   string
	Value string
	Op    string // "Put" or "Append"
	Client ClientID
	Seq Sequence
}

type PutAppendReply struct {
	WrongLeader bool
	Err         Err
}

type GetArgs struct {
	Key string
	Client ClientID
	Seq Sequence
}

type GetReply struct {
	WrongLeader bool
	Err         Err
	Value       string
}
