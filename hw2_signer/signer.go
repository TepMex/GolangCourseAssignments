package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const threadNum = 6

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

func SingleHash(in, out chan interface{}) {
	cnt := 0
	for data := range in {
		dataString := fmt.Sprintf("%v", data)
		fmt.Printf("%d SingleHash data %v\n", cnt, dataString)

		crcMd5 := DataSignerMd5(dataString)
		fmt.Printf("%d SingleHash md5(data) %v\n", cnt, crcMd5)

		crcMd5 = DataSignerCrc32(crcMd5)
		fmt.Printf("%d SingleHash crc(md5(data)) %v\n", cnt, crcMd5)

		crc := DataSignerCrc32(dataString)
		fmt.Printf("%d SingleHash crc(data) %v\n", cnt, crc)

		result := fmt.Sprintf("%s~%s", crc, crcMd5)
		fmt.Printf("%d SingleHash result %v\n", cnt, result)

		cnt++
		out <- result
	}
}

func MultiHash(in, out chan interface{}) {
	for data := range in {
		var result string
		var workerChannels []chan string

		wg := &sync.WaitGroup{}

		dataString := fmt.Sprintf("%v", data)

		for th := 0; th < threadNum; th++ {
			workerChannels = append(workerChannels, make(chan string, 1))
			thStr := strconv.Itoa(th)
			startCrc32Worker(wg, thStr+dataString, workerChannels[th])
		}

		wg.Wait()

		for th, ch := range workerChannels {
			res, ok := <-ch

			if ok {
				result += res
				fmt.Printf("%s MultiHash: crc32(th+step1) %d %s\n", data, th, res)
			}
		}

		fmt.Printf("%s MultiHash result: %s\n", dataString, result)
		out <- result
	}
}

func startCrc32Worker(wg *sync.WaitGroup, data string, out chan<- string) {
	wg.Add(1)
	go func() {
		crc := DataSignerCrc32(data)
		out <- crc
		wg.Done()
		close(out)
	}()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for data := range in {
		resultStr := fmt.Sprintf("%v", data)
		results = append(results, resultStr)
	}

	sort.Strings(results)
	result := strings.Join(results, "_")
	fmt.Printf("CombineResults %s\n", result)
	out <- result
}
