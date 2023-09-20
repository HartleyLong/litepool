package litepool

func (t *TaskOptions) SetTaskWithInterface(f TaskExec) *TaskOptions {
	t.task = f.Exec
	return t
}

func (t *TaskOptions) SetOnSuccessWithInterface(f TaskOnSuccess) *TaskOptions {
	t.onSuccess = f.OnSuccess
	return t
}

func (t *TaskOptions) SetOnErrorWithInterface(f TaskOnError) *TaskOptions {
	t.onError = f.OnError
	return t
}

func (t *TaskOptions) SetOnCompleteWithInterface(f TaskOnComplete) *TaskOptions {
	t.onComplete = f.OnComplete
	return t
}
