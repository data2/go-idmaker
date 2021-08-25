package main

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
)

func TestBenchmarkName(t *testing.T) {
	idMaker := &IdMaker{SeqId: SeqId{id: 0}}
	c := Client{}

	//resultMap := make(map[int32]int)
	//security map
	var resultMap sync.Map

	sw := sync.WaitGroup{}
	sw.Add(18000)
	for i := 0; i < 18000; i++ {
		go func() {
			defer sw.Done()
			resultMap.Store(multiThread(idMaker, c), i)
		}()
	}

	sw.Wait()

	//count
	count := 0
	resultMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	fmt.Println(count)

}

func multiThread(maker *IdMaker, client Client) int32 {
	return maker.GetNewSeqId(client).id
}

func TestStruct(t *testing.T) {
	seqid := SeqId{
		1,
		sync.RWMutex{},
	}
	seqid2 := &seqid
	println(&seqid == &(*seqid2)) // true
}

func TestWriteFile(t *testing.T) {
	ioutil.WriteFile("~/idMaker.txt", []byte("2"), 0664)
}

func TestPath(t *testing.T) {
	print(PathJoin())
}
