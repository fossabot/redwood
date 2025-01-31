package redwood

import (
	"redwood.dev/types"
)

type TxStore interface {
	Start() error
	Close()

	AddTx(tx *Tx) error
	RemoveTx(stateURI string, txID types.ID) error
	TxExists(stateURI string, txID types.ID) (bool, error)
	FetchTx(stateURI string, txID types.ID) (*Tx, error)
	AllTxsForStateURI(stateURI string, fromTxID types.ID) TxIterator
	KnownStateURIs() ([]string, error)
	MarkLeaf(stateURI string, txID types.ID) error
	UnmarkLeaf(stateURI string, txID types.ID) error
	Leaves(stateURI string) ([]types.ID, error)
}

type TxIterator interface {
	Next() *Tx
	Cancel()
	Error() error
}

type txIterator struct {
	ch       chan *Tx
	chCancel chan struct{}
	err      error
}

func (i *txIterator) Next() *Tx {
	select {
	case tx := <-i.ch:
		return tx
	case <-i.chCancel:
		return nil
	}
}

func (i *txIterator) Cancel() {
	close(i.chCancel)
}

func (i *txIterator) Error() error {
	return i.err
}
