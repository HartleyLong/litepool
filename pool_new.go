package litepool

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

func NewPool(maxProcess int64, jobQueuelen int) *ListPool {
	// 使用给定的背景创建一个新的带取消功能的上下文。
	// Create a new context with cancellation using the provided background.
	ctx, cancel := context.WithCancel(context.Background())

	// 创建一个长度为maxProcess的TaskOptions通道数组。
	// Create a slice of TaskOptions channels with a length of maxProcess.
	taskChans := make([]chan *TaskOptions, maxProcess)

	g := &ListPool{
		task: taskChans, // 任务通道
		// Task channels
		numCount:     make([]int64, maxProcess),
		timeCount:    make([]time.Duration, maxProcess),
		statusWorker: make([]chan struct{}, maxProcess),
		quit:         make(chan struct{}), // 退出通道
		// Exit channel
		poolAction: make(chan poolAction), // 工作通道
		// Work channel
		ctx:     ctx,
		cancel:  cancel,
		idleRun: make(chan int64, maxProcess*int64(jobQueuelen)), // 可以处理任务的协程
		// Goroutines that can handle tasks
		workRun: make(chan int64, maxProcess), // 可以工作的协程
		// Goroutines that can work
		jobQueuelen: jobQueuelen,
		maxProcess:  int(maxProcess),
		idle:        make(chan int64, maxProcess*2),
		mutex:       sync.Mutex{},
		heap:        NewIntHeap(maxProcess, int64(jobQueuelen)), // 创建一个新的整数堆
		// Create a new integer heap
	}

	// 初始化整数堆
	// Initialize the integer heap
	heap.Init(g.heap)

	for i := int64(0); i < maxProcess; i++ {
		g.run(i, true)
	}

	// 以下是一些被注释掉的goroutine，它们是为了协程池的其他功能，如监视和自动调整大小。
	// The following are some commented out goroutines that are for other features of the goroutine pool, like monitoring and auto-scaling.
	//go g.heap.jobQueueStatus(ctx)
	//go g.monitorAndScale() // 运行一个自动调整协程池大小的任务
	// Run a task to auto-scale the goroutine pool size
	//go g.poolActionr()     // 协程管理任务
	// Goroutine management task

	return g
}
