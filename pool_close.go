package litepool

// 添加一个方法来优雅地关闭协程池
func (lp *ListPool) Close() {
	for _, tg := range lp.TaskGroupList {
		tg.Wait()
	}
	//printMemUsage()
	lp.close = true
	// Step 1: Cancel the associated context
	if lp.cancel != nil {
		lp.cancel()
	}

	// Step 2: Close all the channels
	//close(lp.quit)
	//return
	for _, ch := range lp.task {
		for i := 0; i < len(ch); i++ {
			_ = <-ch
		}
		//close(ch)
	}
	for _, ch := range lp.statusWorker {
		for i := 0; i < len(ch); i++ {
			_ = <-ch
		}
		//close(ch)
	}
	//close(lp.poolAction)
	for i := 0; i < len(lp.idleRun); i++ {
		<-lp.idleRun
	}
	//close(lp.idleRun)
	for i := 0; i < len(lp.workRun); i++ {
		<-lp.workRun
	}
	//close(lp.workRun)
	for i := 0; i < len(lp.idle); i++ {
		<-lp.idle
	}
	//close(lp.idle)

	// Step 3: Wait for all goroutines to complete

	// Step 4: Clean up resources
	//lp.task = nil
	//lp.numCount = nil
	//lp.timeCount = nil
	//lp.statusWorker = nil
	//lp.poolAction = nil
	//lp.quit = nil
	//lp.idleRun = nil
	//lp.workRun = nil
	//lp.idle = nil
	if lp.heap != nil {
		lp.heap.Close()
		//lp.heap = nil
	}
	//lp = nil
}
