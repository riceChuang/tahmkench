package db

import (
	"github.com/riceChuang/tahmkench/config"
	"gitlab.paradise-soft.com.tw/glob/dispatcher"
)

func SetUpKafkaClient(config *config.KafkaSetting) (err error) {
	if len(config.Brokers) > 0 {
		err = dispatcher.Init(config.Brokers)
		if err != nil {
			return
		}
	}
	return
}

