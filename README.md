# litepool
LitePool协程池特点是低内存，高稳定，易使用，任务状态回调，协程池运行信息的监测等

example

```package main

import (
	"fmt"
	"github.com/HartleyLong/litepool"
	"math/rand"
	"sync/atomic"
	"time"
)

// randomDelay 是一个工具函数，模拟随机延迟
// randomDelay is a utility function that simulates a random delay.
func randomDelay(min, max int) {
	randDuration := time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond
	time.Sleep(randDuration)
}

func main() {
	// 初始化一个新的协程池
	// Initialize a new coroutine pool.
	lite := litepool.NewPool(5, 5)
	// The first parameter is the number of coroutine pools, the second parameter is the maximum task queue for each coroutine pool.
	// 第一个参数是协程池的数量，第二个参数是每个协程池的最大任务队列。
	defer lite.Close()
	n := 120
	lite.Add(n) // If you need to use wait to wait for all tasks to be completed, please set the task number in advance.
	// 如果要使用wait等待所有任务完成，请预先设置任务数量。
	// Special note: If you need to complete manually, please set lite.Done() when returning nil in SetTask.
	// 特别说明，如果需要手动完成，请在SetTask里 return nil的时候设置lite.Done()
	// Otherwise, you must set SetAutoDone in options.
	// 否则必须在options里设置SetAutoDone。
	tmp := int32(0)
	for i := 0; i < n; i++ {
		a := i

		// 定义任务选项
		// Define the task options.
		opt := litepool.NewTaskOptions().
			// 当任务成功执行时的回调
			// Callback when the task is executed successfully.
			SetOnSuccess(func() {
				atomic.AddInt32(&tmp, 1)
				if int(tmp) == n {
					fmt.Println(fmt.Sprintf("任务总数量：%v,完成了%v次", n, atomic.LoadInt32(&tmp)))
				}
			}).
			// 当任务执行发生错误时的回调
			// Callback when there's an error in task execution.
			SetOnError(func(handle *litepool.ErrHandle, err error) {
				handle.ErrReload(1) // Retry once.
				fmt.Println(err)
			}).
			// 当任务完成（无论成功或失败）时的回调
			// Callback when the task completes (whether successful or not).
			SetOnComplete(func() {
				fmt.Println("complate")
			}).
			// 定义具体的任务内容
			// Define the actual task content.
			SetTask(func() error {
				randomDelay(5, 100)
				fmt.Println("这是业务，我输出了", a)
				return nil
			}).
			// 设置自动完成任务
			// Set the task to auto-complete.
			SetAutoDone().
			// 设置任务超时时长
			// Set the task timeout duration.
			SetTimeout(time.Second).
			// 设置任务超时回调
			// Set the callback for when the task times out.
			SetonTimeout(func() {
				// The callback when not added to the task queue within the custom time.
			})

		// 添加任务到协程池
		// Add the task to the coroutine pool.
		lite.AddTask(opt)
	}
	// 等待所有任务完成
	// Wait for all tasks to complete.
	lite.Wait()
	lite.Usage()
	// Usage 用于输出每个协程的运行信息
	// Usage is used to print out the runtime information of each goroutine
}
```
