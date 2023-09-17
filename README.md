# litepool
litePool是用Golang开发的协程池。它的特点是低内存使用、任务状态回调以及协程池状态的监测。

litePool is a goroutine pool developed in Golang. It features low memory usage, task status callbacks, and the ability to monitor the state of the coroutine pool.

```
go get -u github.com/HartleyLong/litepool
```

example

以下代码进行了1000次的随机5-500毫秒耗时任务:
The following code carried out 1,000 tasks that took a random time between 5 to 500 milliseconds.


```
package main

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
	n := 1000
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
				randomDelay(5, 500)
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
	lite.Close()
	// 调用 runtime.GC() 以立即释放内存
	// Call runtime.GC() to release memory immediately.
	//runtime.GC()
}
```


Goroutine 0 is still running, with a total of 0 jobs:
Goroutine 0 仍在运行，总计有0个任务。

My goroutine ID is 0 I have executed tasks 189 times My total execution time for tasks is: 51862 milliseconds:
我的 goroutine ID是0，我已经执行了189次任务，我的任务总执行时间为：51862毫秒。

Goroutine 1 is still running, with a total of 0 jobs:
Goroutine 1 仍在运行，总计有0个任务。

My goroutine ID is 1 I have executed tasks 205 times My total execution time for tasks is: 52365 milliseconds:
我的 goroutine ID是1，我已经执行了205次任务，我的任务总执行时间为：52365毫秒。

Goroutine 2 is still running, with a total of 0 jobs:
Goroutine 2 仍在运行，总计有0个任务。

My goroutine ID is 2 I have executed tasks 201 times My total execution time for tasks is: 51899 milliseconds:
我的 goroutine ID是2，我已经执行了201次任务，我的任务总执行时间为：51899毫秒。

Goroutine 3 is still running, with a total of 0 jobs:
Goroutine 3 仍在运行，总计有0个任务。

My goroutine ID is 3 I have executed tasks 194 times My total execution time for tasks is: 52113 milliseconds:
我的 goroutine ID是3，我已经执行了194次任务，我的任务总执行时间为：52113毫秒。

Goroutine 4 is still running, with a total of 0 jobs:
Goroutine 4 仍在运行，总计有0个任务。

My goroutine ID is 4 I have executed tasks 211 times My total execution time for tasks is: 52779 milliseconds:
我的 goroutine ID是4，我已经执行了211次任务，我的任务总执行时间为：52779毫秒。


该协程池在任务分配上实现了较好的负载均衡。
The goroutine pool has achieved good load balancing in task allocation.

