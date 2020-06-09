package raftkv

import (
	"labrpc"
	"sync"
)
import "crypto/rand"
import "math/big"


type Clerk struct {
	mu sync.Mutex
	servers []*labrpc.ClientEnd

	id ClientID
	seq Sequence

	leader int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.id = ClientID(nrand())
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("RaftKV.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {
	ck.mu.Lock()
	DPrintf("%d (client): called Get(%v)", ck.id, key)
	args := GetArgs{
		Key:    key,
		Client: ck.id,
		Seq:    ck.seq,
	}
	ck.seq++
	ck.mu.Unlock()


	for {
		var reply GetReply
		ok := ck.servers[ck.leader].Call("RaftKV.Get", &args, &reply)
		if ok && !reply.WrongLeader {
			DPrintf("Success")
			return reply.Value
		}
		ck.leader = (ck.leader + 1) % len(ck.servers)
	}
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("RaftKV.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	ck.mu.Lock()
	args := PutAppendArgs{
		Key:    key,
		Value:  value,
		Op:     op,
		Client: ck.id,
		Seq:    ck.seq,
	}
	ck.seq++
	ck.mu.Unlock()

	for {
		var reply PutAppendReply
		ok := ck.servers[ck.leader].Call("RaftKV.PutAppend", &args, &reply)
		if ok && !reply.WrongLeader {
			return
		}
		ck.leader = (ck.leader + 1) % len(ck.servers)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
