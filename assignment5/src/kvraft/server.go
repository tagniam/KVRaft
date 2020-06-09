package raftkv

import (
	"encoding/gob"
	"labrpc"
	"log"
	"raft"
	"sync"
)

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}


type Op struct {
	Client ClientID
	Seq Sequence

	Type string // "Put" or "Append" or "Get"
	Key string
	Value string
}

type RaftKV struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	maxraftstate int // snapshot if log grows this big

	store  map[string]string // stores key/value pairs
	commit *PubSub // communication channel with goroutines handling client requests
	seen   Dedupe // maps client ID -> last seen request id
	done   chan struct{}
}

func (kv *RaftKV) Start(op Op) bool {
	index, _, isLeader := kv.rf.Start(op)
	if !isLeader {
		return false
	}

	kv.mu.Lock()
	// Subscribe to topic `index`
	ch := kv.commit.Subscribe(index)
	kv.mu.Unlock()

	applied := <-ch
	if applied != op {
		return false
	}

	kv.mu.Lock()
	switch op.Type {
	case "Put":
		kv.store[op.Key] = op.Value
	case "Append":
		kv.store[op.Key] += op.Value
	}
	kv.mu.Unlock()

	return true
}

func (kv *RaftKV) Get(args *GetArgs, reply *GetReply) {
	op := Op{
		Client: args.Client,
		Seq:    args.Seq,
		Type:   "Get",
		Key:    args.Key,
	}

	ok := kv.Start(op)

	if !ok {
		reply.WrongLeader = true
	} else {
		reply.Err = OK

		kv.mu.Lock()
		if v, ok := kv.store[args.Key]; ok {
			reply.Value = v
		}
		// TODO update duplicates
		kv.mu.Unlock()
	}
}

func (kv *RaftKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	op := Op{
		Client: args.Client,
		Seq:    args.Seq,
		Type:   args.Op,
		Key:    args.Key,
		Value:  args.Value,
	}

	ok := kv.Start(op)
	if !ok {
		reply.WrongLeader = true
	} else {
		reply.Err = OK
	}
}

//
// the tester calls Kill() when a RaftKV instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *RaftKV) Kill() {
	close(kv.done)
	kv.rf.Kill()
}

func (kv *RaftKV) Wait() {
	for {
		select {
		case msg := <-kv.applyCh:
			DPrintf("%d (server): received committed op: %+v", kv.me, msg)
			kv.mu.Lock()
			// TODO check duplicate
			select {
			case <-kv.done:
				return
			default:
				kv.commit.Publish(msg.Index, msg.Command.(Op))
			}
			kv.mu.Unlock()
		case <-kv.done:
			return
		}
	}
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots with persister.SaveSnapshot(),
// and Raft should save its state (including log) with persister.SaveRaftState().
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *RaftKV {
	// call gob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	gob.Register(Op{})

	kv := new(RaftKV)
	kv.me = me
	kv.maxraftstate = maxraftstate

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.store = make(map[string]string)
	kv.done = make(chan struct{})
	kv.commit = NewPubSub()

	go kv.Wait()

	return kv
}
