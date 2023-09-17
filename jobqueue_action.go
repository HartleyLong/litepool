package litepool

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerStatus 表示工作者的状态，包括其ID和其完成的工作数量
type WorkerStatus struct {
	Id       int64
	JobCount int
}

// IntHeap 是一个自定义堆结构，主要用于跟踪工作者并确保工作均匀分配
type IntHeap struct {
	Count         []*int64           // 使用索引存储工作次数
	Keys          []int64            // 储存工作者ID
	statusAdd     chan int64         //协程启动后的通知
	statusDone    chan int64         //协程关闭后的通知
	offlineWorker map[int64]struct{} //离线管理，当离线发现有job任务没执行，转发 给其他worker
	KeyMap        map[int64]int      // 哈希表，用于快速查找工作者ID对应的索引
	mutex         sync.RWMutex       // 读写锁，确保并发安全
	jobAdd        chan int64         // 接收新工作的通道
	jobDone       chan int64         // 接收完成工作的通道
	jobQueuelen   int64              // 工作队列长度，当count>=这个数表示队列满了
}

// NewIntHeap 初始化一个新的IntHeap
func NewIntHeap(maxProcess, jobQueuelen int64) *IntHeap {
	return &IntHeap{
		Count:         make([]*int64, 0),
		Keys:          make([]int64, 0),
		KeyMap:        map[int64]int{},
		offlineWorker: map[int64]struct{}{},
		statusAdd:     make(chan int64, maxProcess),
		statusDone:    make(chan int64, maxProcess),
		mutex:         sync.RWMutex{},
		jobAdd:        make(chan int64, maxProcess*jobQueuelen),
		jobDone:       make(chan int64, maxProcess*jobQueuelen),
		jobQueuelen:   jobQueuelen,
	}
}

// Len 返回堆中的元素数量
func (h *IntHeap) Len() int { return len(h.Keys) }

// Less 是一个比较方法，决定堆中的元素顺序
func (h *IntHeap) Less(i, j int) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	ii := h.Keys[i]
	jj := h.Keys[j]
	return *h.Count[ii] < *h.Count[jj]
}

// Swap 交换堆中的两个元素
func (h *IntHeap) Swap(i, j int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Keys[i], h.Keys[j] = h.Keys[j], h.Keys[i]
}

// Delete 从堆中删除一个工作者
func (h *IntHeap) Delete(id0 int64) {
	h.mutex.Lock()
	index, exists := h.KeyMap[id0]
	if !exists {
		return
	}
	//a := int64(0)
	//h.Count[id0] = &a
	//h.Count = append(h.Count[:id0], h.Count[id0+1:]...)
	h.Keys = append(h.Keys[:index], h.Keys[index+1:]...)
	delete(h.KeyMap, id0)
	h.mutex.Unlock()
	if len(h.Keys) > 1 {
		//heap.Fix(h, index)
	}
}

// Push 将一个新的工作者添加到堆中
func (h *IntHeap) Push(x interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	o := x.(WorkerStatus)
	i64 := int64(o.JobCount)
	if len(h.Count) <= int(o.Id) {
		h.Count = append(h.Count, &i64)
	}
	h.Count[o.Id] = &i64
	h.Keys = append(h.Keys, o.Id)
	h.KeyMap[o.Id] = o.JobCount
}

// Pop 从堆顶部移除并返回一个工作者ID，同时更新其工作计数
func (h *IntHeap) Pop() interface{} {
	if len(h.Keys) == 0 {
		return nil
	}
	n := len(h.Keys)
	id := h.Keys[n-1] // 注意，我们正在获取顶部的元素，而不是最后一个
	h.mutex.RLock()
	_, exists := h.KeyMap[id]
	h.mutex.RUnlock()
	if !exists {
		return nil
	}
	if *h.Count[id] >= h.jobQueuelen {
		return nil
	}
	//每次pop这个值就加1，然后重新排序
	atomic.AddInt64(h.Count[id], 1)
	heap.Fix(h, n-1) // 重新排序堆，确保其属性
	return id
}

// Add 增加工作者的工作计数，并重新排序堆
func (h *IntHeap) Add(n int64) {
	atomic.AddInt64(h.Count[n], 1)
	index, exists := h.KeyMap[n]
	if !exists {
		return
	}
	heap.Fix(h, index) // 重新排序堆，确保其属性
}

// Done 减少工作者的工作计数，并重新排序堆
func (h *IntHeap) Done(n int64) {
	atomic.AddInt64(h.Count[n], -1)
	index, exists := h.KeyMap[n]
	if !exists {
		return
	}
	heap.Fix(h, index) // 重新排序堆，确保其属性
	return
}

// jobQueueaction 是一个循环，监听新工作和完成的工作，然后更新堆状态
func (h *IntHeap) jobQueueaction(ctx context.Context) {
	//未使用，直接使用Add和Done同步更好。
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-h.jobAdd:
			h.Add(n)
		case n := <-h.jobDone:
			h.Done(n)
		}
	}
}

// Close gracefully shuts down the IntHeap by closing channels and stopping goroutines
func (h *IntHeap) Close() {
	// Close all channels
	close(h.statusAdd)
	close(h.statusDone)
	close(h.jobAdd)
	close(h.jobDone)

	// Clear data structures
	for k := range h.KeyMap {
		delete(h.KeyMap, k)
	}
	for k := range h.offlineWorker {
		delete(h.offlineWorker, k)
	}
	h.Count = nil
	h.Keys = nil
	h.offlineWorker = nil
}

// job动态管理，开发停滞中。。。
func (h *IntHeap) jobQueueStatus(ctx context.Context) {
	//job状态管理
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for n, count := range h.Count {
				if *count > 0 {
					fmt.Println(n, "离线，job任务还有：", *h.Count[n])
				}
			}
			for n, _ := range h.offlineWorker {
				fmt.Println(n, "离线，job任务还有：", *h.Count[n])
			}
		case n := <-h.statusAdd:
			delete(h.offlineWorker, n)
		case n := <-h.statusDone:
			h.offlineWorker[n] = struct{}{}
		}
	}
}
