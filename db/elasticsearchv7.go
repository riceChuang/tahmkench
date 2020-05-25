package db

import (
	"github.com/cenk/backoff"
	"github.com/olivere/elastic/v7"
	"github.com/prometheus/common/log"
	"github.com/riceChuang/tahmkench/config"
	"time"
)

var elasticSearchClientV7 *elastic.Client

func SetUpElasticSearchClientV7(config *config.ElasticSearchSetting, isUpdate bool) (client *elastic.Client, err error) {

	//check config value
	if len(config.Address) <= 0 {
		return
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Duration(40) * time.Second

	if err = backoff.Retry(func() error {

		client, err = elastic.NewClient(
			elastic.SetSniff(config.Sniff),
			elastic.SetURL(config.Address...),
		)

		if err != nil {
			log.Infof("%s, try to connect..", err.Error())
			return err
		}

		return nil

	}, bo); err != nil {
		log.Errorf("elastic connect timeout: %s", err.Error())
	}

	log.Info("elastic ping success..")
	if isUpdate {
		elasticSearchClientV7 = client
	}
	return
}

func GetElasticSearchV7Client() *elastic.Client {
	return elasticSearchClientV7
}
