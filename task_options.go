package litepool

import (
	"time"
)

// TaskOptions 结构体定义了任务的选项和回调
// The TaskOptions structure defines options and callbacks for tasks.
type TaskOptions struct {
	task func() error // 需要执行的任务
	// Task to be executed.
	onSuccess func() // 任务成功执行后的回调
	// Callback after the task is successfully executed.
	onError func(*ErrHandle, error) // 任务执行错误的回调
	// Callback when the task encounters an error.
	onComplete func() // 任务执行完毕的回调
	// Callback after the task is completed.
	onTimeout func() // 任务超时的回调
	// Callback when the task times out.
	waitTimeOut time.Duration // 新增任务的等待时间
	// Waiting time to add a new task.
	reNum int // 任务的重试次数
	// Retry count for the task.
	autoDone bool // 是否自动完成任务
	// Whether to automatically finish the task.
}

// ErrHandle 结构体定义了错误处理的方式
// The ErrHandle structure defines error handling methods.
type ErrHandle struct {
	opt *TaskOptions // 对应的任务选项
	// Corresponding task options.
	reNum int // 任务的重试次数
	// Retry count for the task.
	lp *ListPool // 对应的协程池
	// Corresponding goroutine pool.
}

func (eh *ErrHandle) ErrReload(reNum int, afterFunc func(error)) {
	// 任务发生错误回调的重试，次数为-1的话会请求到error为nil为止
	// Retry the task when an error callback occurs. If the retry count is -1, it will retry until error is nil.
	// 由于任务失败回调的时候没有done 所以这里可以done
	// Since there's no "done" during the task failure callback, it can be done here.
	n := <-eh.lp.idleRun // 取出一个可以执行任务的协程
	// Fetch a goroutine that can execute the task.
	var err error
	defer func() {
		// 记得收回这个占用线程
		if afterFunc != nil {
			afterFunc(err)
		}
		if eh.opt.autoDone && err == nil {
			eh.lp.Done()
		}
		// Remember to reclaim the occupied thread.
		eh.lp.idleRun <- n // 完成后释放
		// Release after completion.
	}()
	if reNum < 1 { // 当重试次数<1则为无限
		// If the retry count is less than 1, it means infinite retries.
		for i := 1; i > -1; i++ {
			//log.Println(fmt.Sprintf("重试次数%v,重试第%v次", reNum, i))
			// Logging the retry count and the current attempt number.
			err = eh.opt.task()
			if err == nil {
				return
			}
		}
	}
	for i := 1; i <= reNum; i++ {
		//log.Println(fmt.Sprintf("重试次数%v,重试第%v次", reNum, i))
		// Logging the retry count and the current attempt number.
		err = eh.opt.task()
		if err == nil {
			return
		}
	}
	return
}

func NewTaskOptions() *TaskOptions {
	return &TaskOptions{}
}
func (t *TaskOptions) SetTask(f func() error) *TaskOptions {
	t.task = f
	return t
}
func (t *TaskOptions) SetOnSuccess(f func()) *TaskOptions {
	t.onSuccess = f
	return t
}

func (t *TaskOptions) SetAutoDone() *TaskOptions {
	//all为不管是否发生错误都直接done
	t.autoDone = true
	return t
}

func (t *TaskOptions) SetOnError(f func(*ErrHandle, error)) *TaskOptions {
	t.onError = f
	return t
}

func (t *TaskOptions) SetOnComplete(f func()) *TaskOptions {
	t.onComplete = f
	return t
}

func (t *TaskOptions) SetTimeout(duration time.Duration) *TaskOptions {
	t.waitTimeOut = duration
	return t
}

func (t *TaskOptions) SetonTimeout(f func()) *TaskOptions {
	t.onTimeout = f
	return t
}
