package scaler

import (
	pb "github.com/AliyunContainerService/scaler/proto"
)

type MyWindow struct {
	Threshold     uint64
	Time          []uint64
	ConcurrentNum []int
	ActiveRequest map[string]bool
}

func (w *MyWindow) Timeline() uint64 {
	if len(w.Time) == 0 {
		return 0
	}
	return w.Time[len(w.Time)-1] - w.Time[0]
}

func (w *MyWindow) Evict() {
	for len(w.Time) > 0 && w.Timeline() > w.Threshold {
		w.Time = w.Time[1:]
		w.ConcurrentNum = w.ConcurrentNum[1:]
	}
}

// 返回需要扩容的机器数量
func (w *MyWindow) Judge() int {
	if len(w.ConcurrentNum) > 5 {
		var ltime = w.Time[len(w.Time)-5]
		var rtime = w.Time[len(w.Time)-1]
		var lnum = w.ConcurrentNum[len(w.ConcurrentNum)-5]
		var rnum = w.ConcurrentNum[len(w.ConcurrentNum)-1]
		if rtime-ltime < 10 && rnum-lnum > 3 {
			return rnum - lnum
		}
	}
	return 0
}

func (w *MyWindow) Append(request *pb.AssignRequest) {
	w.Time = append(w.Time, request.GetTimestamp())
	w.ConcurrentNum = append(w.ConcurrentNum, len(w.ActiveRequest))
	w.ActiveRequest[request.GetRequestId()] = true
	w.Evict()
	return
}
