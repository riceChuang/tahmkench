package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/service/match"
	"github.com/riceChuang/tahmkench/service/source"
	"github.com/riceChuang/tahmkench/types"
	"time"
)

var manager *WarehouseManager

type WarehouseManager struct {
	WarehouseQueue chan types.Record
	MatchMap       map[string][]match.Sender
}

func NewWarehouseManager() *WarehouseManager {
	manager = &WarehouseManager{
		WarehouseQueue: make(chan types.Record, 500000),
		MatchMap:       make(map[string][]match.Sender),
	}
	return manager
}

func (wm *WarehouseManager) InitialAndMonitorMatch(config []config.Match) {
	ctx := context.Background()
	for _, m := range config {
		sender := match.GetSender(m)
		if sender == nil {
			log.Errorf("Sender not Found, source:%v, tag:%v", m.MatchType, m.Tag)
			continue
		}
		go sender.SendRecord(ctx)
		log.Infof("start match sender, match:%v, tag:%v, startTime:%v", m.MatchType, m.Tag, time.Now().String())
		wm.MatchMap[m.Tag] = append(wm.MatchMap[m.Tag], sender)
	}
}

func (wm *WarehouseManager) InitialSource(config []config.Source) (err error) {
	for _, s := range config {
		err := source.InitCollector(s)
		if err != nil {
			return err
		}
	}
	return
}

func (wm *WarehouseManager) RunSource() {
	source.RunCollector(&wm.WarehouseQueue)
}

func (wm *WarehouseManager) PopSourceRecord(recordQueue *chan types.Record, managerWorker int) {
	if managerWorker <= 0 {
		managerWorker = types.DefaultWorkerCount
	}
	for i := 1; i <= managerWorker; i++ {
		go func() {
			for record := range *recordQueue {
				if senders, ok := wm.MatchMap[record.Tag]; ok {
					for _, s := range senders {
						s.StoreQueue(record)
					}
				}
			}
		}()
		log.Infof("manager worker start, worker:%d, startTime:%v", i, time.Now().String())
	}
}
