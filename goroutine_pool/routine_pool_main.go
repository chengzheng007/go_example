package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	goroutinePoolAndTimeout()
}

// 阻塞型channel+waitgroup
type gorountinePool struct {
	ch chan struct{}
	wg *sync.WaitGroup
}

func NewGorountinePool(cap int) *gorountinePool {
	if cap <= 0 {
		panic("channel capacity must > 0")
	}
	return &gorountinePool{
		wg: new(sync.WaitGroup),
		ch: make(chan struct{}, cap),
	}
}

func (p *gorountinePool) Add(num int) {
	if num <= 0 {
		panic("add routine num must > 0")
	}
	p.wg.Add(num)
	for i := 0; i < num; i++ {
		p.ch <- struct{}{}
	}
}

func (p *gorountinePool) Done() {
	<-p.ch
	p.wg.Done()
}

func (p *gorountinePool) Wait() {
	p.wg.Wait()
}

func goroutinePoolAndTimeout() {
	done := make(chan struct{})
	// 阻塞channel，接收routine函数的错误
	errChan := make(chan error)

	pool := NewGorountinePool(3)
	rnum := 10
	for i := 1; i <= rnum; i++ {
		// 如果超过pool容量，会阻塞在此
		pool.Add(1)
		go func(id int){
			defer pool.Done()
			time.Sleep(2*time.Second)
			fmt.Printf("exec %dth routine...\n", id)
			err := someFunc(id)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	go func() {
		pool.Wait()
		close(done)
	}()

	// 只能监控剩余的最后剩余的协程是否超时
	select {
	case err := <-errChan:
		fmt.Printf("get error:%v\n", err)
	case <-done:
		fmt.Println("all routine done.")
	case <-time.After(time.Second):
		fmt.Println("timeout!")
	}
}
// 缺点：超时并不是对pool中所有的routine时间总和的控制
// 57行：pool.Add(1)，如果数量达到pool上限(pool.ch已满)，将会阻塞在这里，等待已加入的routine先运行
// 假设pool容量为3，有10个待运行的routine，每次加入3个就阻塞，等到有运行完的才会加入第4个，等到10个全部加入后
// 执行到60行的go fun(){pool.Wait()...}以及下面的select，也就是说，这里的timeout监控实际值监控了
// 最后<=3个routine的超时（或者说最多只监控了pool容量个数的routine的时间）

func someFunc(id int) error {
	//if rand.Intn(10) % 2 == 1 {
	//	return fmt.Errorf("exec %dth routine error", id)
	//}
	return nil
}