package litepool

import (
	"context"
	"sync"
	"time"
)

// ListPool 结构体用于管理和操作协程池
// The ListPool structure is used to manage and operate a goroutine pool.
type ListPool struct {
	task []chan *TaskOptions // 每个协程都有自己的专属chan，用于大并发情况下避免竞争
	// Each goroutine has its own dedicated channel, used to avoid contention in high concurrency scenarios.
	numCount []int64 // 记录每个协程的任务计数
	// Record task count for each goroutine.
	timeCount []time.Duration // 记录每个协程的执行时间
	// Record execution time for each goroutine.
	statusWorker []chan struct{} // 记录每个协程的状态
	// Record the status of each goroutine.
	wg sync.WaitGroup // 用于等待所有协程完成
	// Used to wait for all goroutines to complete.
	poolAction chan poolAction // 用于管理协程池的通道
	// Channel used for managing the goroutine pool.
	quit chan struct{} // 关闭通道，用于停止协程
	// Closing channel, used to stop the goroutines.
	lastScaleUpTime time.Time // 记录协程池最后一次增加协程的时间
	// Records the last time the goroutine pool scaled up.
	ctx context.Context // 上下文，常用于协程的生命周期管理
	// Context, commonly used for goroutine lifecycle management.
	cancel context.CancelFunc // 与上下文配合使用的取消函数
	// Cancel function to be used in conjunction with the context.
	close bool // 指示协程池是否关闭
	// Indicates whether the goroutine pool is closed.
	idleRun chan int64 // 空闲协程的通道，表示可以执行任务的协程
	// Channel of idle goroutines, indicating goroutines that can execute tasks.
	workRun chan int64 // 通道，代表这个协程可以启动
	// Channel, indicating that this goroutine can start.
	maxProcess int // 最大处理数量
	// Maximum processing count.
	jobQueuelen int // 工作队列的长度
	// Length of the job queue.
	idle chan int64 // 空闲协程的通道
	// Channel for idle goroutines.
	mutex sync.Mutex // 互斥锁，用于同步
	// Mutex for synchronization.
	heap *IntHeap // job的优先级算法
	// Priority algorithm for jobs.
	chanCount *chanCount
}

// poolAction 结构体用于描述协程池的操作，如新增和退出协程
// The poolAction structure is used to describe operations on the goroutine pool, such as adding and exiting goroutines.
type poolAction struct {
	add int64 // 需要新增的协程数量
	// Number of goroutines to be added.
	quit int64 // 需要退出的协程数量
	// Number of goroutines to be quit.
}
