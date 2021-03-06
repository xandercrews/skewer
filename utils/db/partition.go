package db

import (
	"github.com/dgraph-io/badger"
	"github.com/stephane-martin/skewer/utils"
)

var trueBytes = []byte("true")

type partitionImpl struct {
	parent *badger.DB
	prefix string
}

func (p *partitionImpl) Prefix() string {
	return p.prefix
}

func (p *partitionImpl) Get(key utils.MyULID, dst []byte, txn *NTransaction) (ret []byte, err error) {
	return PTransactionFrom(txn, p.prefix).Get(string(key), dst)
}

func (p *partitionImpl) Set(key utils.MyULID, value string, txn *NTransaction) (err error) {
	err = PTransactionFrom(txn, p.prefix).Set(string(key), value)
	if err != nil {
		txn.Discard()
		return err
	}
	return nil
}

func (p *partitionImpl) AddManyTrueMap(m map[utils.MyULID]string, txn *NTransaction) (err error) {
	if len(m) == 0 {
		return nil
	}
	for uid := range m {
		err = PTransactionFrom(txn, p.prefix).Set(string(uid), "true")
		if err != nil {
			txn.Discard()
			return err
		}
	}
	return nil
}

func (p *partitionImpl) AddManySame(uids []utils.MyULID, v string, txn *NTransaction) (err error) {
	if len(uids) == 0 {
		return nil
	}
	ptxn := PTransactionFrom(txn, p.prefix)
	for _, uid := range uids {
		err = ptxn.Set(string(uid), v)

		if err != nil {
			txn.Discard()
			return err
		}
	}
	return nil
}

func (p *partitionImpl) AddMany(m map[utils.MyULID]string, txn *NTransaction) (err error) {
	if len(m) == 0 {
		return nil
	}
	ptxn := PTransactionFrom(txn, p.prefix)
	for key, v := range m {
		err = ptxn.Set(string(key), v)
		if err != nil {
			txn.Discard()
			return err
		}
	}
	return nil
}

func (p *partitionImpl) Exists(key utils.MyULID, txn *NTransaction) (bool, error) {
	_, err := PTransactionFrom(txn, p.prefix).Get(string(key), nil)
	if err == nil {
		return true, nil
	}
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	return false, err
}

func (p *partitionImpl) Delete(key utils.MyULID, txn *NTransaction) (err error) {
	err = PTransactionFrom(txn, p.prefix).Delete(string(key))
	if err != nil {
		txn.Discard()
		return err
	}
	return nil
}

func (p *partitionImpl) DeleteMany(keys []utils.MyULID, txn *NTransaction) (err error) {
	if len(keys) == 0 {
		return nil
	}

	ptxn := PTransactionFrom(txn, p.prefix)

	for _, key := range keys {
		err = ptxn.Delete(string(key))
		if err != nil {
			txn.Discard()
			return err
		}
	}
	return nil
}

func (p *partitionImpl) ListKeysTo(txn *NTransaction, dest []utils.MyULID) []utils.MyULID {
	dest = dest[:0]
	iter := p.KeyIterator(txn)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		dest = append(dest, iter.Key())
	}
	iter.Close()
	return dest
}

func (p *partitionImpl) ListKeys(txn *NTransaction) []utils.MyULID {
	return p.ListKeysTo(txn, nil)
}

func (p *partitionImpl) Count(txn *NTransaction) int {
	var l int
	iter := p.KeyIterator(txn)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		l++
	}
	iter.Close()
	return l
}

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

func (p *partitionImpl) KeyIterator(txn *NTransaction) *ULIDIterator {
	opt := badger.IteratorOptions{
		PrefetchValues: false,
		PrefetchSize:   100,
		Reverse:        false,
		AllVersions:    false,
	}
	return &ULIDIterator{iter: txn.NewIterator(opt), secret: nil, prefix: []byte(p.prefix)}
}

func (p *partitionImpl) KeyValueIterator(txn *NTransaction) *ULIDIterator {
	opt := badger.IteratorOptions{
		PrefetchValues: true,
		PrefetchSize:   100,
		Reverse:        false,
		AllVersions:    false,
	}
	return &ULIDIterator{iter: txn.NewIterator(opt), secret: nil, prefix: []byte(p.prefix)}
}

func NewPartition(parent *badger.DB, prefix string) Partition {
	return &partitionImpl{parent: parent, prefix: prefix}
}
