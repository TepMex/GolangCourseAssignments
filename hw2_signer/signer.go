package main

import "sync"

func ExecutePipeline(jobs ...job) {
	var chans []chan interface{}

	chans = append(chans, make(chan interface{}))
	wg := &sync.WaitGroup{}
	for i, jb := range jobs {
		out := make(chan interface{})
		startJob(wg, jb, chans[i], out)
		chans = append(chans, out)
	}
	wg.Wait()
}

func startJob(wg *sync.WaitGroup, jb job, in, out chan interface{}) {
	wg.Add(1)
	go func() {
		jb(in, out)
		close(out)
		wg.Done()
	}()
}

func SingleHash(in, out chan interface{})     {}
func MultiHash(in, out chan interface{})      {}
func CombineResults(in, out chan interface{}) {}
