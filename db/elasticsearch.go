package db

import (
	"github.com/cenk/backoff"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"time"
)

var elasticSearchClient *elastic.Client

func SetUpElasticSearchClient(config *config.ElasticSearchSetting, isUpdate bool) (client *elastic.Client, err error) {

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
		elasticSearchClient = client
	}
	return
}

func GetElasticSearchClient() *elastic.Client {
	return elasticSearchClient
}
