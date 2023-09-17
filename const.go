package litepool

import "time"

const (
	// 扩充协程池后一定时间内不释放协程池
	// Duration after which a goroutine pool is not released once expanded
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	TaskQuitCD = time.Second * 10

	// 运行任务的数量和总协程数量的比例大于多少开始自动增加协程
	// When the ratio of running tasks to total goroutines exceeds this threshold, goroutines are automatically added
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	MinAutoAdd = 0.7

	// 运行任务的数量和总协程数量比例小于多少开始释放协程
	// When the ratio of running tasks to total goroutines is below this threshold, goroutines are automatically released
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	MaxAutoQuit = 0.3

	// 每次自动新增协程池对于总数的比例
	// The ratio of the number of goroutines added automatically to the total number of goroutines each time
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	AutoAddScale = 0.1

	// 每次自动退出协程池的比例
	// The ratio of goroutines that are automatically exited from the pool each time
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	AutoQuitScale = 0.1

	// 自动检测缩放的时间
	// Time interval for automatic scale checks
	// Note: This is for future auto-scaling of goroutine pool, not implemented in the current version
	AutoCheckScaleTime = time.Second * 1
)
