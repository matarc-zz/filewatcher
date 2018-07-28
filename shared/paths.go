package shared

import (
	"github.com/boltdb/bolt"
	"github.com/matarc/filewatcher/log"
)

type Paths struct {
	Db *bolt.DB
}

// Update is an RPC that take a list of operations as an argument (`transaction`) and
// returns a list of all successful operations in `reply`.
// It returns an error if any operation can't be completed.
func (p *Paths) Update(transaction *Transaction, reply *Transaction) error {
	log.Info("Update")
	return p.Db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(transaction.Id))
		if err != nil {
			return err
		}
		log.Infof("Attempt to update paths for nodewatcher '%s'", transaction.Id)
		for _, op := range transaction.Operations {
			if op.Event&Create == Create {
				log.Infof("Adding '%s' in the database", op.Path)
				err = b.Put([]byte(op.Path), []byte{})
				if err != nil {
					return err
				}
				reply.Operations = append(reply.Operations, op)
			} else if op.Event&Remove == Remove {
				log.Infof("Removing '%s' in the database", op.Path)
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

func (p *Paths) ListFiles(_ *struct{}, list *[]Node) error {
	log.Info("ListFiles")
	return p.Db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			node := Node{Id: string(name)}
			err := b.ForEach(func(k, v []byte) error {
				log.Infof("'%s' - '%s'", node.Id, string(k))
				node.Files = append(node.Files, string(k))
				return nil
			})
			if err != nil {
				return err
			}
			*list = append(*list, node)
			log.Infof("Summary--------\n%v", *list)
			return nil
		})
	})
}

func (p *Paths) DeleteList(id string, _ *struct{}) error {
	return p.Db.Batch(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(id))
	})
}
