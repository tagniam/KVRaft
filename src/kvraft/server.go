package raftkv

import (
	"encoding/gob"
	"labrpc"
	"log"
	"net/rpc"
	"raft"
	"sync"
	"time"
)

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.

	Type   string
	Client int64
	Seq    int

	Key    string
	Value  string
}

type RaftKV struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	maxraftstate int // snapshot if log grows this big

	// Your definitions here.
	store map[string]string
	lastSeq map[int64]int
	commitCh map[int]chan Op
}

func (kv *RaftKV) Get(args *GetArgs, reply *GetReply) error {
	op := Op{
		Type: "Get",
		Client: args.Client,
		Seq: args.Seq,
		Key: args.Key,
	}

	ok := kv.AppendEntryToLog(op)
	if ok {
		reply.WrongLeader = false

		kv.mu.Lock()
		defer kv.mu.Unlock()
		val, ok := kv.store[args.Key]
		if !ok {
			reply.Err = ErrNoKey
		} else {
			reply.Err = OK
			reply.Value = val
			kv.lastSeq[args.Client] = args.Seq
		}
	} else {
		reply.WrongLeader = true
	}

	return nil
}

func (kv *RaftKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) error {
	op := Op{
		Type: args.Op,
		Client: args.Client,
		Seq: args.Seq,
		Key: args.Key,
		Value: args.Value,
	}

	ok := kv.AppendEntryToLog(op)
	if !ok {
		reply.WrongLeader = true
	} else {
		reply.WrongLeader = false
		reply.Err = OK
	}
	return nil
}

func (kv *RaftKV) AppendEntryToLog(entry Op) bool {
	index, _, isLeader := kv.rf.Start(entry)
	if !isLeader {
		return false
	}

	kv.mu.Lock()
	ch, ok := kv.commitCh[index]
	if !ok {
		ch = make(chan Op, 1)
		kv.commitCh[index] = ch
	}
	kv.mu.Unlock()

	select {
	case op := <-ch:
		return op == entry
	case <-time.After(1000 * time.Millisecond):
		// timeout
		return false
	}
}

//
// the tester calls Kill() when a RaftKV instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *RaftKV) Kill() {
	kv.rf.Kill()
}

func (kv *RaftKV) Wait() {
	for {
		msg := <-kv.applyCh
		cmd := msg.Command.(Op)

		kv.mu.Lock()
		v, ok := kv.lastSeq[cmd.Client]
		if !ok || v < cmd.Seq {
			switch cmd.Type {
			case "Put":
				kv.store[cmd.Key] = cmd.Value
			case "Append":
				kv.store[cmd.Key] += cmd.Value
			}
			kv.lastSeq[cmd.Client] = cmd.Seq
		}

		ch, ok := kv.commitCh[msg.Index]
		if ok {
			ch <- cmd
		} else {
			kv.commitCh[msg.Index] = make(chan Op, 1)
		}
		kv.mu.Unlock()
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
func StartKVServer(servers []labrpc.Client, me int, persister *raft.Persister, maxraftstate int) *RaftKV {
	// call gob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	gob.Register(Op{})

	kv := new(RaftKV)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// Your initialization code here.

	kv.store = make(map[string]string)
	kv.lastSeq = make(map[int64]int)
	kv.commitCh = make(map[int]chan Op)

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	go kv.Wait()

	err := rpc.Register(kv)
	if err != nil {
		log.Printf("could not register %d: %v", me, err)
	}

	err = rpc.Register(kv.rf)
	if err != nil {
		log.Printf("could not register %d: %v", me, err)
	}
	return kv
}
