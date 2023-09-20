package litepool

import (
	"container/heap"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

func (lp *ListPool) run(n int64, index bool) error {
	if len(lp.workRun) == 0 && !index {
		// 进程池已满
		// The process pool is full
		return errors.New("协程池满了")
	}
	// 发送开启协程给job接口监听。当前版本未启用。
	// Send the coroutine startup to the job interface listener. The current version is not enabled.
	// lp.heap.statusAdd <- n
	if lp.task[n] == nil {
		// 初始化通道
		// Initialize channel
		lp.task[n] = make(chan *TaskOptions, lp.jobQueuelen+1)
	}
	if lp.statusWorker[n] == nil {
		// 初始化通道
		// Initialize channel
		lp.statusWorker[n] = make(chan struct{}, 1)
	}
	lp.statusWorker[n] <- struct{}{} // 发送工作状态
	// Send Work Status
	// 堆优先算法压入
	// Heap priority algorithm push
	heap.Push(lp.heap, WorkerStatus{
		Id:       n,
		JobCount: 0,
	})
	lp.idle <- n // 发送空闲协程
	// Send Idle Protocol
	go func() {
		defer func() {
			// 收回工作状态，此时len lp.statusWorker[n]==0
			// Retract the working state, at which point len lp.statusWorker[n]==0
			if !lp.close {
				<-lp.statusWorker[n]
				// 收回可接收的任务
				// Withdraw acceptable tasks
				for i := 0; i < lp.jobQueuelen; i++ {
					_ = <-lp.idleRun
				}
				<-lp.idle // 收回可用协程
				// Retrieving available processes
				lp.workRun <- n // 告诉通道我可以工作了
				// Tell the channel that I can work now
				lp.heap.statusDone <- n // 发送关闭协程给job接口监听
				// Send the shutdown protocol to the job interface for listening
			}
		}()
		// 发送job队列（可用任务队列）给通道
		// Send job queue (available task queue) to channel
		for i := 0; i < lp.jobQueuelen; i++ {
			// 发送可用
			// Send available
			lp.idleRun <- n
		}
		for {
			select {
			case <-lp.ctx.Done():
				//close(lp.task[n])
				//close(lp.statusWorker[n])
				return
			case <-lp.quit:
				lp.heap.Delete(n) // 从job队列中删除
				// Remove from job queue
				// 收到了减少协程池的信号
				// Received the signal to reduce the coroutine pool
				// 判断是否有job
				// Check if there's a job
				if len(lp.task[n]) == 0 {
					return
				}
			case f, ok := <-lp.task[n]:
				if !ok {
					// 当lp.task 关闭，这里将为false
					// When lp.task is closed, this will be false
					return
				}
				atomic.AddInt64(&lp.numCount[n], 1)
				// 协程处理的任务计数
				// Count of tasks processed by the coroutine
				// 错误处理：防止panic导致工作协程终止
				// Error handling: prevent panic causing worker coroutine to terminate
				var err error
				func() {
					defer func() {
						r := recover()
						// 无论任务是否成功，都执行onComplete回调
						// Execute the onComplete callback whether the task is successful or not
						if f.onComplete != nil {
							f.onComplete()
						}
						if r != nil && f.onError != nil {
							// 执行错误的回调
							// Execute the error callback
							f.onError(&ErrHandle{
								lp:  lp,
								opt: f,
							}, f.tg, fmt.Errorf("task panicked: %v", r))
						}
						if !lp.close {
							if len(lp.task[n]) == 0 {
								lp.idle <- n
							} else {
								lp.idleRun <- n
							}
						}
						lp.heap.Done(n)
					}()
					start := time.Now()
					err = f.task() // 执行任务
					// Execute the task
					lp.timeCount[n] += time.Since(start)
					if f.onSuccess != nil && err == nil {
						// 如果没有panic，执行成功的回调
						// If there is no panic, execute the successful callback
						f.onSuccess()
					}
					if err == nil && f.autoDone {
						// 执行完毕success函数后才执行done
						// Only execute done after the success function is completed
						f.tg.wg.Done()
					}
					if err != nil && f.onError != nil {
						// 执行错误的回调
						// Execute the error callback
						f.onError(&ErrHandle{
							lp:  lp,
							opt: f,
						}, f.tg, err)
					}
				}()
			}
		}
	}()
	return nil
}

// Usage 用于输出每个协程的运行信息
// Usage is used to print out the runtime information of each goroutine
func (lp *ListPool) Usage() {
	for n, t := range lp.task {
		// 输出指定协程的正在运行的任务数
		// Print the number of tasks that are still running for a specified goroutine
		fmt.Println(fmt.Sprintf("Goroutine %v, with a total of %v jobs", n, len(t)))
		// 输出协程的ID，执行次数，以及总执行时间
		// Print the ID of the goroutine, the number of executions, and the total execution time
		fmt.Println("My goroutine ID is", n, "I have executed tasks", lp.numCount[n], "times", "My total execution time for tasks is:", lp.timeCount[n].Milliseconds(), "milliseconds")
	}
}
