package source

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/types"
	"sync"
)

var (
	collectorMap = map[string]Collector{}
	mu           = sync.Mutex{}
)

type Collector interface {
	Initialize(setting interface{}) (err error)
	CollectionSource(ctx context.Context, recordQueue *chan types.Record)
}

func InitCollector(source config.Source) (err error) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := collectorMap[source.SourceType+"-"+source.Tag]; !ok {
		instance := CreateCollector(source)
		if instance != nil {
			err := instance.Initialize(source.Setting)
			if err != nil {
				return fmt.Errorf("Collector Initialize fail, source:%s, tag:%s, err:%v", source.SourceType, source.Tag, err)
			}
			log.Infof("collector initialize success, source:%s, tag:%s", source.SourceType, source.Tag)
			collectorMap[source.SourceType+"-"+source.Tag] = instance
		}
	}
	return
}

func RunCollector(recordQueue *chan types.Record) {
	ctx := context.Background()
	for _, collector := range collectorMap {
		go collector.CollectionSource(ctx, recordQueue)
	}
}

func CreateCollector(source config.Source) (instance Collector) {
	switch types.RepoType(source.SourceType) {
	case types.RepoKafka:
		switch types.TagType(source.Tag) {
		case types.TagBetlog, types.TagRecoveryBetlog, types.TagRecoveryTransaction, types.TagTransaction:
			instance = &Kafka{Tag: source.Tag, Worker: source.CollectWorker}
		}
	case types.RepoTahmkench:
		switch types.TagType(source.Tag) {
		case types.TagBrand:
			instance = &TahmkenchBrand{Worker: source.CollectWorker}
		case types.TagRecoveryBrand:
		}
	}
	return
}
