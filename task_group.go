package litepool

import (
	"sync"
)

type TaskGroup struct {
	wg sync.WaitGroup
}

func (lp *ListPool) NewTaskGroup(taskNum int) *TaskGroup {
	wg := sync.WaitGroup{}
	wg.Add(taskNum)
	tg := &TaskGroup{wg: wg}
	lp.mutex.Lock()
	lp.TaskGroupList = append(lp.TaskGroupList, tg)
	lp.mutex.Unlock()
	return tg
}
func (tg *TaskGroup) Done() {
	tg.wg.Done()
}
func (tg *TaskGroup) Wait() {
	tg.wg.Wait()
}
