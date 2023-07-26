/*
Copyright 2023 The Alibaba Cloud Serverless Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"container/list"
	"fmt"
	"github.com/AliyunContainerService/scaler/go/pkg/config"
	"github.com/AliyunContainerService/scaler/go/pkg/model"
	scaler2 "github.com/AliyunContainerService/scaler/go/pkg/scaler"
	pb "github.com/AliyunContainerService/scaler/proto"
	"log"
	"sync"
)

type Manager struct {
	rw           sync.RWMutex
	schedulers   map[string]scaler2.Scaler
	RequestReply map[string]*pb.Assignment
	Window       map[string]MyWindow
	config       *config.Config
}

type MyWindow struct {
	Threshold  uint64
	Time       *list.List
	Concurrent *list.List
}

func (w *MyWindow) Timeline() uint64 {
	return w.Time.Back().Value.(uint64) - w.Time.Front().Value.(uint64)
}
func (w *MyWindow) Evict() {
	for w.Time.Len() > 0 && w.Timeline() > w.Threshold {
		w.Time.Remove(w.Time.Front())
		w.Concurrent.Remove(w.Concurrent.Front())
	}
}

func (w *MyWindow) Append(time uint64, concurrent_num int64) {
	w.Time.PushBack(time)
	w.Concurrent.PushBack(concurrent_num)
	w.Evict()
	return
}

func New(config *config.Config) *Manager {
	return &Manager{
		rw:           sync.RWMutex{},
		schedulers:   make(map[string]scaler2.Scaler),
		Window:       make(map[string]MyWindow),
		RequestReply: make(map[string]*pb.Assignment),
		config:       config,
	}
}

func (m *Manager) GetOrCreate(metaData *model.Meta) (scaler2.Scaler, MyWindow) {
	m.rw.RLock()
	if scheduler := m.schedulers[metaData.Key]; scheduler != nil {
		m.rw.RUnlock()
		return scheduler, m.Window[metaData.Key]
	}
	m.rw.RUnlock()

	m.rw.Lock()
	if scheduler := m.schedulers[metaData.Key]; scheduler != nil {
		m.rw.Unlock()
		return scheduler, m.Window[metaData.Key]
	}
	log.Printf("Create new scaler for app %s", metaData.Key)
	scheduler := scaler2.New(metaData, m.config)
	window := &MyWindow{ // 将 Window 字段的类型修改为指向 MyWindow 的指针
		Threshold:  1000,       // 初始化 Threshold 字段为 1000
		Time:       list.New(), // 初始化 Time 字段为一个新的双向链表
		Concurrent: list.New(), // 初始化 Concurrent 字段为一个新的双向链表
	}
	m.schedulers[metaData.Key] = scheduler
	m.Window[metaData.Key] = *window
	m.rw.Unlock()
	return scheduler, *window
}

func (m *Manager) Get(metaKey string) (scaler2.Scaler, error) {
	m.rw.RLock()
	defer m.rw.RUnlock()
	if scheduler := m.schedulers[metaKey]; scheduler != nil {
		return scheduler, nil
	}
	return nil, fmt.Errorf("scaler of app: %s not found", metaKey)
}
