package litepool

import (
	"time"
)

// 动态管理协程池的函数。
// Dynamic goroutine pool management function.
// 目前开发停滞，与job链接时存在BUG。
// Development is currently stalled. There are bugs when interfacing with the job.
func (lp *ListPool) monitorAndScale() {
	ticker := time.NewTicker(AutoCheckScaleTime) // 用于自动扩展的计时器
	// Timer for auto-scaling.
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			workR := cap(lp.workRun) - len(lp.workRun) // 计算正在运行的协程数
			// Calculate the number of running goroutines.
			if workR == 0 {
				continue
			}
			idleR := workR - len(lp.idle) // 计算正在运行的任务数
			// Calculate the number of running tasks.
			taskR := workR*lp.jobQueuelen - len(lp.idleRun) // 计算正在运行的任务数
			// Calculate the number of tasks running.
			bl := float32(taskR+idleR) / float32(workR*lp.jobQueuelen+workR) // 计算负载平衡
			// Calculate the load balance.
			if bl >= MinAutoAdd && workR < cap(lp.workRun) {
				// 当一定比例的任务正在运行时，增加协程池的大小
				// When a certain ratio of tasks is running, increase the size of the goroutine pool.
				ad := int64(len(lp.workRun)) / (AutoAddScale * 100)
				if ad == 0 {
					ad = 1
				}
				lp.poolAction <- poolAction{add: ad}
				lp.lastScaleUpTime = time.Now()
			} else if bl < MaxAutoQuit && time.Since(lp.lastScaleUpTime) < TaskQuitCD {
				//fmt.Println("任务较少，但在新增协程后的冷却时间内，因此不减少协程")
				// "Tasks are few, but within the cooldown time after adding goroutines, so don't reduce the goroutines."
			} else if bl < MaxAutoQuit && time.Since(lp.lastScaleUpTime) >= TaskQuitCD {
				qu := int64(workR) / (AutoQuitScale * 100)
				if qu == 0 {
					qu = 1
				}
				lp.poolAction <- poolAction{quit: qu}
			}
		case <-lp.ctx.Done():
			return
		}
	}
}

func (g *ListPool) poolActionr() {
	//这是一个操作协程池新增和释放的任务
	for {
		select {
		case <-g.ctx.Done():
			return
		case w := <-g.poolAction:
			if w.add > 0 {
				for i := int64(0); i < w.add; i++ {
					if len(g.workRun) == 0 {
						break
					}
					n := <-g.workRun
					//g.task[n] = make(chan *TaskOptions)
					g.run(n, false)
				}
			} else {
				for i := int64(0); i < w.quit; i++ {
					if cap(g.workRun)-len(g.workRun) > 1 {
						//始终保持着有一个工作线程
						//fmt.Println("quit")
						g.quit <- struct{}{}
					}
				}
			}
		}
	}
}
