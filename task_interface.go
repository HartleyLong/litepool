package litepool

// 定义一些task接口
type TaskExec interface {
	Exec() error
}

type TaskOnSuccess interface {
	OnSuccess()
}

type TaskOnError interface {
	OnError(*ErrHandle, *TaskGroup, error)
}

type TaskOnComplete interface {
	OnComplete()
}
