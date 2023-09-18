package litepool

import (
	"container/heap"
	"errors"
	"time"
)

func (lp *ListPool) AddTask(opt *TaskOptions) error {
	// 检查任务是否存在
	// Check if the task is present
	if opt.task == nil {
		return errors.New("请添加任务")
	}

	var n int64
	add := true

	// 检查是否有空闲的协程可用
	// Check if there are any idle goroutines available
	if len(lp.idle) > 0 {
		n = <-lp.idle
	} else {
		// 从堆中尝试获取协程
		// Try to get a goroutine from the heap
		w := heap.Pop(lp.heap)
		// TODO: Limit the number of jobs here
		if w != nil {
			add = false
			_ = <-lp.idleRun // 获取一个空闲协程
			// Fetch an idle goroutine
			n = w.(int64)
		} else {
			n = -1
		}
	}

	// 如果没有等待超时并且没有可用的协程
	// If there's no wait timeout and no available goroutines
	if opt.waitTimeOut == 0 && n == -1 {
		n = <-lp.idleRun
	}

	// 如果没有可用的协程则使用计时器
	// Use a timer if no goroutines are available
	if n == -1 {
		timer := time.NewTimer(opt.waitTimeOut)
		defer timer.Stop()

		select {
		// 如果没有协程可用则任务超时
		// Task times out if no goroutine becomes available
		case <-timer.C:
			if opt.onTimeout != nil {
				opt.onTimeout() // 处理超时场景
				// Handle timeout scenario
			}
			if opt.autoDone {
				opt.tg.wg.Done() // 自动标记任务完成
				// Automatically mark the task as done
			}
			return errors.New("任务超时")
		// 等待协程变为可用
		// Wait until a goroutine becomes available
		case w := <-lp.idleRun:
			n = w
		}
	}

	// 将任务发送给选定的协程
	// Send the task to the selected goroutine
	lp.task[n] <- opt

	// 如果堆中添加了新的协程，更新堆
	// If a new goroutine was added to the heap, update the heap
	if add {
		lp.heap.Add(n)
	}

	return nil
}
