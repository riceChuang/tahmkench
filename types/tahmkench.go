package types

import (
	"encoding/json"
	"time"
)

var (
	BrandQueue         = make(chan Record, 500000)
	BrnadRevoceryQueue = make(chan Record, 500000)
)

type Record struct {
	Tag            string
	Note           string
	CollectionTime time.Time
	Data           json.RawMessage
}

type RepoType string

const (
	RepoTahmkench        RepoType = "tahmkench" // 用於cm es寫入後再將資料當成輸入點收進來,再依照match發送給品牌
	RepoKafka            RepoType = "kafka"
	RepoMysql            RepoType = "mysql"
	RepoElasticSearch    RepoType = "es"
	RepoElasticSearchNew RepoType = "es7"
)

type TagType string

const (
	TagBetlog              TagType = "betlog"
	TagRecoveryBetlog      TagType = "recovery_betlog"
	TagTransaction         TagType = "transaction"
	TagRecoveryTransaction TagType = "recovery_transaction"
	TagBrand               TagType = "brand"
	TagRecoveryBrand       TagType = "recovery_brand"
)

const (
	DefaultWorkerCount = 3
)
