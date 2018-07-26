package shared

import (
	"github.com/boltdb/bolt"
)

type Paths struct {
	Db *bolt.DB
}

// Update is an RPC that take a list of operations as an argument (`transaction`) and
// returns a list of all successful operations in `reply`.
// It returns an error if any operation can't be completed.
func (p *Paths) Update(transaction *Transaction, reply *Transaction) error {
	return p.Db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(transaction.Id))
		if err != nil {
			return err
		}
		for _, op := range transaction.Operations {
			if op.Event&Create == Create {
				err = b.Put([]byte(op.Path), []byte{})
				if err != nil {
					return err
				}
				reply.Operations = append(reply.Operations, op)
			} else if op.Event&Remove == Remove {
				err = b.Delete([]byte(op.Path))
				if err != nil {
					return err
				}
				reply.Operations = append(reply.Operations, op)
			}
		}
		return nil
	})
}
