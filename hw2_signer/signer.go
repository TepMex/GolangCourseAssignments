package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

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

		dataString := fmt.Sprintf("%v", data)

		for th := 0; th < 6; th++ {
			thStr := strconv.Itoa(th)
			crc := DataSignerCrc32(thStr + dataString)
			fmt.Printf("%s MultiHash: crc32(th+step1) %d %s\n", thStr+dataString, th, crc)
			result += crc
		}
		fmt.Printf("%s MultiHash result: %s\n", dataString, result)
		out <- result
	}
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
