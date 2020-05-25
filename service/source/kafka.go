package source

import (
	"context"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/db"
	"github.com/riceChuang/tahmkench/types"
	"gitlab.paradise-soft.com.tw/glob/dispatcher"
	"time"
)

type Kafka struct {
	Tag     string
	Worker  int
	Setting *config.KafkaSetting
	Queue   *chan types.Record
}

func (k *Kafka) Initialize(setting interface{}) (err error) {
	kafkaSetting := config.KafkaSetting{}
	err = mapstructure.Decode(setting, &kafkaSetting)
	if err != nil {
		return
	}

	if len(kafkaSetting.Brokers) > 0 {
		err = db.SetUpKafkaClient(&kafkaSetting)
		if err != nil {
			return
		}
	}

	// set setting config
	k.Setting = &kafkaSetting
	return
}

func (k *Kafka) CollectionSource(ctx context.Context, recordQueue *chan types.Record) {
	// set input Queue
	k.Queue = recordQueue
	// subscribe all topic from config setting
	// than handle record input queue
	if len(k.Setting.Topic) > 0 {
		for _, v := range k.Setting.Topic {
			ctrl := dispatcher.SubscribeWithRetry(v, k.HandleRecord, 5, k.WaitingTime, dispatcher.ConsumerSetAsyncNum(k.Worker), dispatcher.ConsumerSetGroupID(k.Setting.Group))
			go k.HandleConnectionError(ctrl.Errors())
		}
	}
}

func (k *Kafka) HandleRecord(value []byte) error {
	*k.Queue <- types.Record{
		Tag:            k.Tag,
		CollectionTime: time.Now(),
		Data:           value,
	}
	return nil
}

func (k *Kafka) WaitingTime(failCount int) time.Duration {
	return time.Duration(5)
}

func (k *Kafka) HandleConnectionError(errorMessage <-chan error) {
	filters := log.Fields{
		"SourceType": types.RepoKafka,
		"Tag":        k.Tag,
	}
	logger := log.WithFields(filters)
	err := <-errorMessage
	logger.Errorf("connection Error : %v", err)
	return
}
