package main

import (
	"fmt"
	"sync"
	"time"
)
// 功能：限制单位时间内routine的数量，确保所有routine能执行完，有一个超时检测
// 更准确应该叫：goroutine并发数限制
// 因为所谓的池，池子里面的资源可复用，如sync.Pool，但这里并没有复用起的routine
func main() {
	goroutinePoolAndTimeout()
}

// 阻塞型channel+waitgroup
type gorountineConcurrencyControl struct {
	ch chan struct{}
	wg *sync.WaitGroup
}

func NewGorountineConcurrencyControl(cap int) *gorountineConcurrencyControl {
	if cap < 1 {
		panic("channel capacity must >= 1")
	}
	return &gorountineConcurrencyControl{
		wg: new(sync.WaitGroup),
		ch: make(chan struct{}, cap),
	}
}

func (p *gorountineConcurrencyControl) Add(num int) {
	if num < 1 {
		panic("add routine num must >= 1")
	}
	p.wg.Add(num)
	for i := 0; i < num; i++ {
		p.ch <- struct{}{}
	}
}

func (p *gorountineConcurrencyControl) Done() {
	<-p.ch
	p.wg.Done()
}

func (p *gorountineConcurrencyControl) Wait() {
	p.wg.Wait()
}

func goroutinePoolAndTimeout() {
	done := make(chan struct{})
	// 阻塞channel，接收routine函数的错误
	errChan := make(chan error)

	gcc := NewGorountineConcurrencyControl(3)
	rnum := 10
	for i := 1; i <= rnum; i++ {
		// 如果超过pool容量，会阻塞在此
		gcc.Add(1)
		go func(id int){
			defer gcc.Done()
			time.Sleep(2*time.Second)
			fmt.Printf("exec %dth routine...\n", id)
			err := someFunc(id)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	go func() {
		gcc.Wait()
		close(done)
	}()

	// 只能监控剩余的最后剩余的协程是否超时
	select {
	case err := <-errChan:
		fmt.Printf("get error:%v\n", err)
	case <-done:
		fmt.Println("all routine done.")
	case <-time.After(3*time.Second):
		fmt.Println("timeout!")
	}
}
// 缺点：超时并不是对gorountineConncurent中所有的routine时间总和的控制
// 57行：gcc.Add(1)，如果数量达到gcc上限(gcc.ch已满)，将会阻塞在这里，等待已加入的routine先运行
// 假设gcc容量为3，有10个待运行的routine，每次加入3个就阻塞，等到有运行完的才会加入第4个，等到10个全部加入后
// 执行到60行的go fun(){gcc.Wait()...}以及下面的select，也就是说，这里的timeout监控实际值监控了
// 最后<=3个routine的超时（或者说最多只监控了gcc容量个数的routine的时间）

func someFunc(id int) error {
	//if rand.Intn(10) % 2 == 1 {
	//	return fmt.Errorf("exec %dth routine error", id)
	//}
	return nil
}