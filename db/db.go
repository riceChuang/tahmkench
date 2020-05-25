package db

import (
	"fmt"
	"github.com/riceChuang/tahmkench/config"
)

func SetStorage(config *config.Config) (err error) {
	_, err = SetUpElasticSearchClient(config.ElasticSearch, true)
	if err != nil {
		err = fmt.Errorf("SetUpElasticSearchClient error: %v", err)
		return
	}
	err = SetUpKafkaClient(config.Kakfa)
	if err != nil {
		err = fmt.Errorf("SetUpKafkaClient error: %v", err)
		return
	}
	_, err = SetUpMysqlClient(config.Mysql, true)
	if err != nil {
		err = fmt.Errorf("SetUpMysqlClient error: %v", err)
		return
	}
	_, err = SetUpElasticSearchClientV7(config.ElasticSearchV7, true)
	if err != nil {
		err = fmt.Errorf("SetUpElasticSearchClientV7 error: %v", err)
		return
	}
	return
}
