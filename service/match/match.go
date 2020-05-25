package match

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/types"
	"sync"
)

var (
	senderMap = map[string]Sender{}
	mu        = sync.Mutex{}
)

type Sender interface {
	Initialize(setting interface{}) (err error)
	StoreQueue(record types.Record)
	SendRecord(ctx context.Context)
}

func GetSender(match config.Match) (sender Sender) {
	mu.Lock()
	defer mu.Unlock()
	var ok bool
	if sender, ok = senderMap[match.MatchType+"-"+match.Tag]; !ok {
		instance := CreateSender(match)
		if instance != nil {
			err := instance.Initialize(match.Setting)
			if err != nil {
				log.Errorf("Sender Initialize fail, source:%v, tag:%v, err:%v", match.MatchType, match.Tag, err)
				return
			}
			log.Infof("Sender initialize success, source:%s, tag:%s", match.MatchType, match.Tag)
			senderMap[match.MatchType+"-"+match.Tag] = instance
			sender = instance
		}
	} else {
		return
	}
	return
}

func CreateSender(match config.Match) (instance Sender) {
	switch types.RepoType(match.MatchType) {
	case types.RepoElasticSearch:
		switch types.TagType(match.Tag) {
		case types.TagBetlog, types.TagRecoveryBetlog:
		case types.TagTransaction, types.TagRecoveryTransaction:
			instance = &ElasticSearchBetlog{Tag: match.Tag}
		}
	case types.RepoKafka:
		switch types.TagType(match.Tag) {
		case types.TagBrand, types.TagRecoveryBrand:
			instance = &KafkaBrand{WorkerCount: match.SendWorker, Tag: match.Tag}
		}
	}

	return
}
