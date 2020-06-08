package raftkv

//
// Data structure to store client sequence numbers
type Dedupe struct {
	items map[ClientID]Sequence
}

//
// Returns true if the given client and request id are new and have not
// been seen before.
//
func (d *Dedupe) Update(key ClientID, value Sequence) bool {
	if d.items == nil {
		d.items = make(map[ClientID]Sequence)
	}

	if _, ok := d.items[key]; !ok || value > d.items[key] {
		d.items[key] = value
		return true
	}

	return false
}

