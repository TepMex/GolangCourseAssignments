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
	var resultChannels []chan string
	cnt := 0
	for data := range in {
		wg := &sync.WaitGroup{}

		dataString := fmt.Sprintf("%v", data)
		fmt.Printf("%d SingleHash data %v\n", cnt, dataString)

		crcMd5 := DataSignerMd5(dataString)
		fmt.Printf("%d SingleHash md5(data) %v\n", cnt, crcMd5)

		crcMd5Chan := make(chan string)
		crcChan := make(chan string)
		resultChan := make(chan string)

		startCrc32Worker(wg, crcMd5, crcMd5Chan)
		startCrc32Worker(wg, dataString, crcChan)

		combineSingleHashResult(cnt, crcMd5Chan, crcChan, resultChan)

		resultChannels = append(resultChannels, resultChan)
		cnt++
	}

	for _, ch := range resultChannels {
		res := <-ch
		out <- res
	}
}

func combineSingleHashResult(taskNum int, crcMd5chan, dataChan, resultChan chan string) {

	go func() {
		crcMd5 := <-crcMd5chan
		fmt.Printf("%d SingleHash crc(md5(data)) %v\n", taskNum, crcMd5)
		crcData := <-dataChan
		fmt.Printf("%d SingleHash crc(data) %v\n", taskNum, crcData)

		result := fmt.Sprintf("%s~%s", crcData, crcMd5)
		fmt.Printf("%d SingleHash result %v\n", taskNum, result)
		resultChan <- result
	}()

}

func MultiHash(in, out chan interface{}) {
	var resultChannels []chan string

	for data := range in {

		resultChan := make(chan string)

		go func(data interface{}, resultCh chan string) {

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
				res := <-ch
				result += res
				fmt.Printf("%s MultiHash: crc32(th+step1) %d %s\n", data, th, res)
			}

			fmt.Printf("%s MultiHash result: %s\n", dataString, result)
			resultCh <- result
		}(data, resultChan)

		resultChannels = append(resultChannels, resultChan)
	}

	for _, ch := range resultChannels {
		res := <-ch
		out <- res
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
