package coreoslib

import (
	"fmt"
	"github.com/coreos/locksmith/lock"
	"os"
)

func CoreOsLocksmith(machines []string, max int) error {

	// code from locksmithctl
	elc, err := lock.NewEtcdLockClient(machines)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing etcd client:", err)
		return err
	}
	l := lock.New("hi", elc)

	sem, old, err := l.SetMax(max)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error setting value:", err)
		return err
	}

	fmt.Println("Old-Max:", old)
	fmt.Println("Max:", sem.Max)

	return nil
}
