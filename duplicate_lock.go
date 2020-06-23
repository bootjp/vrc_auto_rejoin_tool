package vrcarjt

import (
	"github.com/gofrs/flock"
	"log"
)

type DupRunLock struct {
	Path string
	lock *flock.Flock
}

type DupRunLocker interface {
	Try() (bool, error)
	Lock() error
	UnLock()
}

func NewDupRunLock(p string) *DupRunLock {
	return &DupRunLock{
		Path: p,
		lock: flock.New(p),
	}
}

func (d *DupRunLock) Try() (bool, error) {
	return d.lock.TryLock()
}

func (d *DupRunLock) Lock() error {
	return d.lock.Lock()
}

func (d *DupRunLock) UnLock() {
	err := d.lock.Unlock()
	if err != nil {
		log.Println(err)
	}

}
