# litepool

LitePool 是一个用 Golang 开发的协程池。它具有以下特点：

低内存使用：优化了资源管理，确保高效的内存使用。

任务状态回调：可以实时追踪任务的状态并进行相应的响应。

错误的重试机制：当任务出错时，LitePool 提供了重新执行任务的机制，增加任务完成的成功率。

灵活的任务定义：LitePool 的设计允许用户使用闭包或接口来定义任务。这意味着你可以选择最适合的方式来描述任务逻辑，不论是使用简单的函数闭包还是更结构化的接口形式，LitePool 都能够完美支持。

支持n个协程作为一个任务组：func (lp *ListPool) AddTaskGroup(opts ...*TaskOptions) error


```
go get -u github.com/HartleyLong/litepool
```

使用闭包来执行任务

```
package main

import (
	"errors"
	"github.com/HartleyLong/litepool"
	"math/rand"
	"time"
)

func main() {
	// 创建一个协程池，其中有5个协程和一个大小为10的任务队列
	lp := litepool.NewPool(5, 10)
	// 确保在主函数结束时关闭协程池
	defer lp.Close()

	// 定义一个任务的数量为1000
	n := 1000
	// 为这些任务创建一个任务组
	tg := lp.NewTaskGroup(n)

	// 循环添加任务到协程池
	for i := 0; i < n; i++ {
		// 定义任务选项
		opt := tg.NewTaskOptions().
			// 设置任务内容
			SetTask(func() error {
				// 任务的执行会随机暂停0到2毫秒
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(3)))

				// 有1/3的机率任务会返回错误
				if rand.Intn(3) == 1 {
					return errors.New("error")
				}
				return nil
			}).
			// 如果任务执行出错，则执行以下内容
			SetOnError(func(handle *litepool.ErrHandle, tg1 *litepool.TaskGroup, err error) {
				// 尝试重新加载任务最多3次
				handle.ErrReload(3, func(err error) {
					// 如果任务连续3次执行失败，标记任务为完成状态
					if err != nil {
						tg1.Done()
					}
				})
			}).
			// 任务成功执行后调用此函数
			SetOnSuccess(func() {

			}).
			// 任务完成后（无论成功或失败）调用此函数
			SetOnComplete(func() {

			}).
			// 设置任务完成后自动标记为完成状态
			SetAutoDone()

		// 向协程池添加任务
		lp.AddTask(opt)
	}

	// 等待所有任务完成
	tg.Wait()

	// 输出协程池的使用情况
	lp.Usage()
}

```

使用接口来执行任务

```
package main

import (
	"errors"
	"fmt"
	"github.com/HartleyLong/litepool"
	"math/rand"
	"sync"
	"time"
)

// myTask 结构体定义了一个任务和它的状态
type myTask struct {
	count int
	n     int
	mu    sync.Mutex // 用于保护 count 和 n 的并发修改
}

// Exec 是任务的主要执行函数
func (mt *myTask) Exec() error {
	// 模拟一个耗时的任务
	time.Sleep(time.Millisecond)
	// 更新任务状态
	mt.mu.Lock()
	mt.count++
	mt.mu.Unlock()
	// 随机产生一个错误
	if rand.Intn(3) == 1 {
		return nil
	}
	return errors.New("故意错误")
}

// OnSuccess 任务成功后的回调函数
func (mt *myTask) OnSuccess() {
	mt.mu.Lock()
	mt.n++
	mt.mu.Unlock()
}

// OnError 任务出错时的回调函数
func (mt *myTask) OnError(handle *litepool.ErrHandle, tg *litepool.TaskGroup, err error) {
	if err != nil {
		// 任务失败后，尝试重新执行任务
		handle.ErrReload(-1, func(err error) {
			fmt.Println(mt.count)
			mt.mu.Lock()
			mt.n++
			mt.mu.Unlock()
			//假设重试次数不是-1
			if err != nil {
				tg.Done()
			}
		})
		// 注意: 这里的 -1 意味着无限重试，应当在实际应用中慎重考虑
	}
}

// OnComplete 任务完成后（无论成功还是失败）的回调函数
func (mt *myTask) OnComplete() {

}

func main() {
	// 设置随机数种子，确保每次的错误随机生成
	rand.Seed(time.Now().UnixNano())

	// 创建一个协程池，包含5个协程和一个任务队列大小为10
	lp := litepool.NewPool(5, 10)
	// 使用 defer 确保 main 函数结束时关闭协程池
	defer lp.Close()

	// 定义任务的数量
	n := 1000
	// 创建一个任务组
	tg := lp.NewTaskGroup(n)
	mt := &myTask{count: 1}
	// 循环添加任务到协程池
	for i := 0; i < n; i++ {
		opt := tg.NewTaskOptions().
			// 使用接口设置任务内容
			SetTaskWithInterface(mt).
			// 使用接口设置任务出错时的回调
			SetOnErrorWithInterface(mt).
			// 使用接口设置任务成功后的回调
			SetOnSuccessWithInterface(mt).
			// 使用接口设置任务完成后的回调
			SetOnCompleteWithInterface(mt).
			// 确保任务完成后自动标记为完成状态
			SetAutoDone()
		// 将任务添加到协程池中
		lp.AddTask(opt)
	}
	// 等待所有任务完成
	tg.Wait()
	// 输出结果以验证所有任务是否都执行了
	fmt.Println(mt.n == n)
	// 打印协程池的使用情况
	lp.Usage()
}

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
	lp := litepool.NewPool(5, 5)
	defer lp.Close()
	// The first parameter is the number of coroutine pools, the second parameter is the maximum task queue for each coroutine pool.
	// 第一个参数是协程池的数量，第二个参数是每个协程池的最大任务队列。
	n := 1000
	tg := lp.NewTaskGroup(n) // If you need to use wait to wait for all tasks to be completed, please set the task number in advance.
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
		opt := tg.NewTaskOptions().
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
				handle.ErrReload(3, func(err error) {
					fmt.Println(err)
					if err != nil {
						//发生错误不会自动done
						tg.Done()
					}
				}) // Retry once.
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
			SetAddTimeout(time.Second).
			// 设置任务超时回调
			// Set the callback for when the task times out.
			SetOnAddTimeout(func() {
				// The callback when not added to the task queue within the custom time.
			})

		// 添加任务到协程池
		// Add the task to the coroutine pool.
		lp.AddTask(opt)
	}
	// 等待所有任务完成
	// Wait for all tasks to complete.
	tg.Wait()
	lp.Usage()
	// Usage 用于输出每个协程的运行信息
	// Usage is used to print out the runtime information of each goroutine
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

