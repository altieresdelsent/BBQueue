package stressTest

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const parallelCountProcessing = 20000
const requestPerParallelProcessing = ((parallelCount * requestPerParallel) / parallelCountProcessing) + 1

func StressTestGetAndDelete() {
	start := time.Now()
	totalErrors := 0
	errorCount := make(chan int)
	go func() {
		for errorCountItem := range errorCount {
			totalErrors += errorCountItem
		}
	}()

	totalTasksProcessed := 0
	taskCount := make(chan int)

	go func() {
		for taskCountItem := range taskCount {
			totalTasksProcessed += taskCountItem
		}
	}()
	endPartialResultsLoop := false
	go func() {
		for i := 0; i < 100000; i++ {
			if endPartialResultsLoop == true {
				break
			}
			fmt.Println("Report number:", i)
			fmt.Println("Partial errors:", totalErrors)
			fmt.Println("Partial tasksProcessed:", totalTasksProcessed)
			fmt.Println("Time for ", totalTasksProcessed, "requisitions processed in seconds:", time.Now().Sub(start).Seconds())
			fmt.Println()
			time.Sleep(time.Second * 10)
		}
	}()

	markAsFinished := make(chan struct{})
	for i := 0; i < parallelCountProcessing; i++ {
		go func() {
			for i := 0; i < requestPerParallelProcessing; i++ {
				statusCodeGET, key := GetPayload()
				if statusCodeGET == http.StatusNoContent {
					continue
				}
				if statusCodeGET > 300 {
					errorCount <- 1
				}
				time.Sleep(time.Second * time.Duration(1+rand.Intn(timeoutQueueSeconds-4)))
				statusCodeDELETE := DeletePayload(key)
				if statusCodeDELETE > 300 {
					errorCount <- 1
				}

				if statusCodeGET == 200 && statusCodeDELETE == 200 {
					taskCount <- 1
				}
			}
			markAsFinished <- struct{}{}
		}()
	}

	for i := 0; i < parallelCountProcessing; i++ {
		<-markAsFinished
	}

	close(markAsFinished)
	close(errorCount)
	close(taskCount)
	endPartialResultsLoop = true

	fmt.Println("Total errors:", totalErrors)
	fmt.Println("Total tasksProcessed:", totalTasksProcessed)
	fmt.Println("Time for ", parallelCount*requestPerParallel, "requisitions in seconds:", time.Now().Sub(start).Seconds())
}
