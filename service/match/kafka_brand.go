package match

import (
	"context"
	"github.com/cenk/backoff"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/db"
	"github.com/riceChuang/tahmkench/types"
	"gitlab.paradise-soft.com.tw/glob/dispatcher"
	"sync"
	"time"
)

type KafkaBrand struct {
	Logger       *log.Entry
	MessageCount int
	WorkerCount  int
	Tag          string
	Setting      *config.KafkaSetting
	Queue        chan types.Record
}

func (kb *KafkaBrand) StoreQueue(record types.Record) {
	kb.Queue <- record
}

func (kb *KafkaBrand) Initialize(setting interface{}) (err error) {
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

	// init queue
	kb.Queue = make(chan types.Record, 500000)
	// set logger
	kb.Logger = log.WithFields(log.Fields{"type": types.RepoKafka, "tag": kb.Tag})
	// set setting config
	kb.Setting = &kafkaSetting
	return
}

func (kb *KafkaBrand) SendRecord(ctx context.Context) {
	tickerTime := time.NewTicker(10 * time.Second).C
	var mu sync.Mutex
	go func() {
		for range tickerTime {
			mu.Lock()
			if kb.MessageCount != 0 {
				kb.Logger.Infof("Message success push to kafka count: %v", kb.MessageCount)
				kb.MessageCount = 0
			}
			mu.Unlock()
		}
	}()

	for sendWorker := 1; sendWorker <= kb.WorkerCount; sendWorker++ {
		go kb.pushMessage()
		kb.Logger.Infof("send worker %v Start", sendWorker)
	}
}

func (kb *KafkaBrand) pushMessage() {
	defer func() {
		if r := recover(); r != nil {
			kb.Logger.Errorf("pushMessage error: %v", r.(error))
			kb.pushMessage()
			return
		}
	}()

	for target := range kb.Queue {
		//handle topic
		topics := []string{}
		if len(kb.Setting.Topic) > 0 {
			topics = kb.Setting.Topic
		} else {
			topics = []string{target.Note}
		}

		if len(topics) > 0 {
			for _, topic := range topics {
				bo := backoff.NewExponentialBackOff()
				bo.MaxElapsedTime = time.Duration(5) * time.Second
				kb.Logger.Debugf("startHandle message, topic: %v, info: %v", topic, string(target.Data))
				if err := backoff.Retry(func() error {
					return kb.sendMessage(topic, target.Data)
				}, bo); err != nil {
					kb.ackErrorHandle(target.Data, err)
					continue
				}
				//add log count
				kb.MessageCount++
				kb.Logger.Debugf("success push message to kafka,topic: %v, info: %v", topic, string(target.Data))
			}
		}
	}

}

func (kb *KafkaBrand) sendMessage(topic string, message []byte) (err error) {
	return dispatcher.Send(topic, message, dispatcher.ProducerAddErrHandler(kb.callBackErrorHandle))
}

func (kb *KafkaBrand) callBackErrorHandle(value []byte, err error) {
	kb.Logger.Errorf("Message Fail, value: %v, error: %v", string(value), err)
}

func (kb *KafkaBrand) ackErrorHandle(value []byte, err error) {
	kb.Logger.Errorf("Message Fail to kafka, value: %v, error: %v", string(value), err)
}
