package main

import (
	"errors"
	"fmt"
	"sync"
)

func halo(name string) (*sync.Mutex, error) {
	if name != "halo" {
		return nil, errors.New("jsjjsjsj")
	}
	return &sync.Mutex{}, nil
}

func main() {
	mtx, err := halo("hh")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer fmt.Println(mtx.TryLock())
}
