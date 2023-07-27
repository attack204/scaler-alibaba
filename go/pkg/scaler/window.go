package scaler

import (
	pb "github.com/AliyunContainerService/scaler/proto"
	"sync"
	"sync/atomic"
)

type MyWindow struct {
	window_mu      sync.Mutex
	Threshold      uint64
	TimeList       []uint64
	ConcurrentList []int32
	ConNum         atomic.Int32
}

func (w *MyWindow) Timeline() uint64 {
	if len(w.TimeList) == 0 {
		return 0
	}
	return w.TimeList[len(w.TimeList)-1] - w.TimeList[0]
}

func (w *MyWindow) Evict() {
	for len(w.TimeList) > 0 && w.Timeline() > w.Threshold {
		w.TimeList = w.TimeList[1:]
		w.ConcurrentList = w.ConcurrentList[1:]
	}
}

// 返回需要扩容的机器数量
func (w *MyWindow) Judge() int32 {
	if len(w.ConcurrentList) > 5 {
		var ltime = w.TimeList[len(w.TimeList)-5]
		var rtime = w.TimeList[len(w.TimeList)-1]
		var lnum = w.ConcurrentList[len(w.ConcurrentList)-5]
		var rnum = w.ConcurrentList[len(w.ConcurrentList)-1]
		if rtime-ltime < 10 && rnum-lnum > 3 {
			return rnum - lnum
		}
	}
	return 0
}

func (w *MyWindow) Append(request *pb.AssignRequest) {
	w.window_mu.Lock()
	defer w.window_mu.Unlock()
	w.TimeList = append(w.TimeList, request.GetTimestamp())
	w.ConcurrentList = append(w.ConcurrentList, w.ConNum.Load())
	w.ConNum.Add(1)
	w.Evict()
	return
}
