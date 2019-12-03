package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	testConcurrencyAndTimeout()
}

func testConcurrencyAndTimeout() {
	var wg sync.WaitGroup
	// 完成通知
	done := make(chan struct{}, 0)
	num := 5
	for i := 1; i <= num; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			time.Sleep(2*time.Second)
			fmt.Printf("process req %d.\n", idx)
		}(i)
	}
	// wg.Wait()必须在goroutine，否则会卡住，后面的超时检测也就没有意义
	go func() {
		wg.Wait()
		// 完成通知
		done <- struct{}{}
	}()

	select {
	case <-done:
		// 正常结束
		fmt.Println("all concurrent req accomplished.")
	case <-time.After(2*time.Second):
		fmt.Println("timeout! we need to break off")
	}
	return
}
// 缺点：只有超时检测，协程数量不可控，可以一下子add非常多routine（需要一个routine pool），否则可能导致系统资源耗尽