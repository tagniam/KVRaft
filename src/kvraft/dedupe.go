package raftkv

/*import "sync"

//
// Data structure to store client sequence numbers
type Dedupe struct {
	mu sync.Mutex
	items map[ClientID]Sequence
}

//
// Returns true if the given client and request id are new and have not
// been seen before.
//
func (d *Dedupe) Update(key ClientID, value Sequence) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.items == nil {
		d.items = make(map[ClientID]Sequence)
	}

	if _, ok := d.items[key]; !ok || value > d.items[key] {
		d.items[key] = value
		return true
	}

	return false
}
*/

type ClientSequence struct {
	lastSeq map[int64]int
}

func (cs *ClientSequence) Update(client int64, seq int) bool {
	last, ok := cs.lastSeq[client]
	if !ok || last < seq {
		cs.lastSeq[client] = seq
		return true
	}
	return false
}

