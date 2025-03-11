package ledger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"github.com/fxamacker/cbor/v2"
)

type Ledger struct {
	Changes []ChangeSet // the first entry is the genesis change set
	Head ChangeSetHash
}

// TODO: should validation be moved from ReadLedger() to DecodeLedger()?
func DecodeLedger(bytes []byte) (*Ledger, error) {
	changes, err := decodeLedgerChanges(bytes)
	if err != nil {
		return nil, err
	}

	var head ChangeSetHash

	l := &Ledger{changes, head}

	l.syncHead()

	return l, nil
}

func (l *Ledger) Encode() []byte {
	return encodeLedgerChanges(l.Changes)
}

func (l *Ledger) syncHead() {
	l.Head = l.Changes[len(l.Changes)-1].Hash()
}

func makeLedgerPath(projectPath string) string {
	return projectPath + "/ledger"
}

func getLedgerPath(genesis *ChangeSet) string {
	h := StringifyProjectHash(genesis.Hash())	

	projectPath := HomeDir + "/" + h

	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	return makeLedgerPath(projectPath)
}

func ReadLedger(validateAssets bool) (*Ledger, error) {
	g, err := LookupGenesisChangeSet()
	if err != nil {
		return nil, err
	}

	gHash := g.Hash()

	ledgerPath := getLedgerPath(g)

	l, ok := readLedger(ledgerPath)
	if !ok {
		l = &Ledger{[]ChangeSet{*g}, gHash}
		if err := l.ValidateAll(validateAssets); err != nil {
			return nil, err
		}

		l.Write()
	} else {
		if err := l.ValidateAll(validateAssets); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	l.syncHead()

	return l, nil
}

func readLedger(ledgerPath string) (*Ledger, bool) {
	if _, err := os.Stat(ledgerPath); err == nil {
		dat, err := os.ReadFile(ledgerPath)

		if err == nil {
			l, err := DecodeLedger(dat)	
			if err == nil {
				return l, true
			} else {
				fmt.Println("Failed to decode ledger")
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	return nil, false
}

func (l *Ledger) ApplyAll(m ResourceManager) {
	for _, c := range l.Changes {
		c.Apply(m)
	}
}

func (l *Ledger) KeepChangeSets(until int) {
	l.Changes = l.Changes[0:until+1]

	l.syncHead()
}

func (l *Ledger) AppendChangeSet(cs *ChangeSet, validateAssets bool) error {
	backup := l.Changes

	l.Changes = append(backup, *cs)

	if err := l.ValidateAll(validateAssets); err != nil {
		l.Changes = backup
		return err
	}

	l.syncHead()

	return nil
}

func (l *Ledger) GetChangeSet(h ChangeSetHash) (*ChangeSet, bool) {
	// start from end
	for i := len(l.Changes); i >= 0; i-- {
		if i > 0 && i == len(l.Changes) {
			if IsSameChangeSetHash(h, l.Head) {
				return &(l.Changes[i-1]), true
			}
		} else if i < len(l.Changes) {
			if i > 1 && IsSameChangeSetHash(h, l.Changes[i].Parent) {
				return &(l.Changes[i-1]), true
			}
		} else if i > 0 {
			if (IsSameChangeSetHash(h, l.Changes[i].Hash())) {
				return &(l.Changes[i]), true
			}
		}
	}

	return nil, false
}

func (l *Ledger) GetChangeSetHashes() *ChangeSetHashes {
	hashes := make([]ChangeSetHash, len(l.Changes))

	for i := 0; i < len(l.Changes); i++ {
		if i+1 == len(l.Changes) {
			hashes[i] = l.Head
		} else if i+1 < len(l.Changes) {
			hashes[i] = l.Changes[i+1].Parent
		} else {
			hashes[i] = l.Changes[i].Hash()
		}
	}

	return &ChangeSetHashes{
		hashes,
	}
}

func (l *Ledger) Write() {
	ledgerPath := getLedgerPath(&(l.Changes[0]))

	bytes := l.Encode()

	err := os.WriteFile(ledgerPath, bytes, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func decodeLedgerChanges(bytes []byte) ([]ChangeSet, error) {
	v := []ChangeSetCbor{}

	err := cbor.Unmarshal(bytes, &v)

	if err != nil {
		return nil, err
	}

	changes := make([]ChangeSet, len(v))

	for i, c := range v {
		cc, err := c.convertToChangeSet()

		if err != nil {
			return nil, err
		}

		changes[i] = *cc
	}

	return changes, nil
}

func encodeLedgerChanges(changes []ChangeSet) []byte {
	v := make([]ChangeSetCbor, len(changes))

	for i, c := range changes {
		v[i] = c.convertToChangeSetCbor(false)
	}

	bs, err := cbor.Marshal(v)

	if err != nil {
		log.Fatal(err)
	}

	return bs
}