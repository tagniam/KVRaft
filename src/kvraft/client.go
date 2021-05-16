package raftkv

import (
	"crypto/rand"
	"labrpc"
	"math/big"
	"sync"
)

type Clerk struct {
	servers []labrpc.Client
	id      int64
	seq     int
	mu      sync.Mutex
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []labrpc.Client) *Clerk {
	ck := new(Clerk)
	ck.servers = servers

	ck.id = nrand()
	ck.seq = 0
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
	args := GetArgs{
		Key:    key,
		Client: ck.id,
	}

	ck.mu.Lock()
	args.Seq = ck.seq
	ck.seq++
	ck.mu.Unlock()

	i := 0
	for {
		server := ck.servers[i]
		var reply GetReply

		ok := server.Call("RaftKV.Get", &args, &reply)
		if ok && !reply.WrongLeader {
			return reply.Value
		}
		i = (i + 1) % len(ck.servers)
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
	args := PutAppendArgs{
		Key:    key,
		Value:  value,
		Op:     op,
		Client: ck.id,
	}

	ck.mu.Lock()
	args.Seq = ck.seq
	ck.seq++
	ck.mu.Unlock()

	i := 0
	for {
		server := ck.servers[i]

		var reply PutAppendReply
		ok := server.Call("RaftKV.PutAppend", &args, &reply)
		if ok && !reply.WrongLeader {
			return
		}
		i = (i + 1) % len(ck.servers)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
