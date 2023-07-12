package stressTest

import (
	"fmt"
	"time"
)

const parallelCount = 100
const requestPerParallel = 100

func StressTestPost() {
	start := time.Now()
	errorCount := make(chan int)
	for i := 0; i < parallelCount; i++ {
		go func() {
			errors := 0
			for i := 0; i < requestPerParallel; i++ {
				if PostRandomPayload() != 200 {
					errors++
				}
			}
			errorCount <- errors
		}()
	}
	totalErrors := 0
	for i := 0; i < parallelCount; i++ {
		totalErrors += <-errorCount
	}
	fmt.Println("Total errors:", totalErrors)
	fmt.Println("Time for ", parallelCount*requestPerParallel, "requisitions in seconds:", time.Now().Sub(start).Seconds())
}
