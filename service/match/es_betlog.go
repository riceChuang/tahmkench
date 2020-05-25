package match

import (
	"context"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/db"
	"github.com/riceChuang/tahmkench/types"
	"time"
)

type ElasticSearchBetlog struct {
	Logger        *log.Entry
	Tag           string
	ElasticClient *elastic.Client
	BetlogQueue   chan types.Record
}

func (esb *ElasticSearchBetlog) StoreQueue(record types.Record) {
	esb.BetlogQueue <- record
	return
}

func (esb *ElasticSearchBetlog) Initialize(setting interface{}) (err error) {
	elasticSearchSetting := config.ElasticSearchSetting{}
	err = mapstructure.Decode(setting, &elasticSearchSetting)
	if err != nil {
		return
	}

	if len(elasticSearchSetting.Address) > 0 {
		esb.ElasticClient, err = db.SetUpElasticSearchClient(&elasticSearchSetting, false)
		if err != nil {
			return
		}
	} else {
		esb.ElasticClient = db.GetElasticSearchClient()
	}

	// initial Queue and config
	esb.BetlogQueue = make(chan types.Record, 500000)
	esb.Logger = log.WithFields(log.Fields{"type": types.RepoElasticSearch, "tag": esb.Tag})
	return
}

func (esb *ElasticSearchBetlog) SendRecord(ctx context.Context) {

	transactions := []*TransactionWithTarget{}
	betlogQueue := esb.BetlogQueue
	timer := time.NewTicker(10 * time.Second).C
	timeoutUTC := time.Now().UTC().Add(60 * time.Second)

	go func() {
		for {
			select {
			case b := <-betlogQueue:
				trans := &TransactionWithTarget{}
				err := json.Unmarshal(b.Data, trans)
				if err != nil {
					esb.Logger.Errorf("json Unmarshal err:%v, data:%v", err, b)
				}
				transactions = append(transactions, trans)
				if len(transactions) > 5000 {
					esb.BulkInsert(transactions)
					transactions = []*TransactionWithTarget{}
					timeoutUTC = time.Now().UTC().Add(60 * time.Second)
				}
			case now := <-timer:
				if len(transactions) > 0 && now.After(timeoutUTC) {
					esb.BulkInsert(transactions)
					transactions = []*TransactionWithTarget{}
					timeoutUTC = time.Now().UTC().Add(60 * time.Second)
				}
			}
		}
	}()
	return
}

func (esb *ElasticSearchBetlog) BulkInsert(transactionsWithTargets []*TransactionWithTarget) {
	defer func() {
		if r := recover(); r != nil {
			esb.Logger.Errorf(" error: %v", r.(error))
			return
		}
	}()
	transactionMap := map[string]*TransactionWithTarget{}
	bulkRequest := esb.ElasticClient.Bulk()
	for _, item := range transactionsWithTargets {
		//transaction
		if _, ok := transactionMap[item.ID]; !ok {
			transactionMap[item.ID] = item
		}
		// valid betlog log
		nowUTC := time.Now().UTC()

		if item.ProviderUpdatedAt == nil {
			item.ProviderUpdatedAt = &nowUTC
		}

		item.UpdatedAt = &nowUTC
		item.CreatedAt = &nowUTC

		index := esb.parseTimeToPartialIndex(tranIndexPrefix, *item.PartitionAt)
		request := elastic.NewBulkIndexRequest().
			OpType("create").
			Index(index).
			Type("logs").
			Id(item.ID).
			Doc(item.Transaction)
		bulkRequest = bulkRequest.Add(request)
	}

	bulkStartTime := time.Now()
	bulkResponse, err := bulkRequest.Do(context.TODO())
	if err != nil {
		esb.Logger.Errorf("bulk insert fail: %s", err.Error())
		return
	}
	elapsed := time.Since(bulkStartTime)
	esb.Logger.Infof("Bulk insert transactions took %s", elapsed)
	mSearch := esb.ElasticClient.MultiSearch()
	failedItems := bulkResponse.Failed()
	if len(failedItems) <= 0 {
		esb.Logger.Debug("no es failed resp")
		return
	} else {
		esb.Logger.Debugf("cm: failedCount: %d", len(failedItems))
		for _, val := range failedItems {
			termQuery := elastic.NewMatchQuery("_id", val.Id)
			req1 := elastic.NewSearchRequest().
				Index(val.Index).
				Type("logs").
				Source(elastic.NewSearchSource().Query(termQuery))
			mSearch = mSearch.Add(req1)
		}
	}
	searchResult, err := mSearch.Do(context.TODO())
	if err != nil {
		esb.Logger.Errorf("transaction mSearch fail: %s", err.Error())
		return
	}
	if searchResult.Responses == nil {
		esb.Logger.Debug("no es failed resp")
		return
	}

	hitTransactions := []*TransactionWithTarget{}
	for _, val1 := range searchResult.Responses {
		for _, val2 := range val1.Hits.Hits {
			oldTransaction := Transaction{}
			err := json.Unmarshal(*val2.Source, &oldTransaction)
			if err != nil {
				esb.Logger.Error(err)
			}
			if v, ok := transactionMap[oldTransaction.ID]; ok {
				hitTransactions = append(hitTransactions, v)
			}

		}
	}

	if len(hitTransactions) > 0 {
		err := esb.BulkUpdateTransactionsV2(hitTransactions)
		if err != nil {
			esb.Logger.Errorf("transaction bulk update execute fail: %s", err.Error())
		}
	}
	return
}

func (esb *ElasticSearchBetlog) BulkUpdateTransactionsV2(items []*TransactionWithTarget) error {
	bulkRequest := esb.ElasticClient.Bulk()
	nowUTC := time.Now().UTC()
	for _, item := range items {
		// valid betlog log
		item.UpdatedAt = &nowUTC
		index := esb.parseTimeToPartialIndex(tranIndexPrefix, *item.PartitionAt)
		request := elastic.NewBulkUpdateRequest().
			Index(index).
			Type("logs").
			Id(item.ID).
			Doc(item.Transaction)
		bulkRequest = bulkRequest.Add(request)
	}
	bulkStartTime := time.Now()
	bulkResponse, err := bulkRequest.Do(context.TODO())
	if err != nil {
		return err
	}
	elapsed := time.Since(bulkStartTime)
	esb.Logger.Infof("Bulk update transactions took %s", elapsed)

	failed := bulkResponse.Failed()
	for _, val := range failed {
		esb.Logger.Errorf("cm: transaction bulk update fail: Id = %s, %s", val.Id, val.Result)
	}
	return nil
}

func (esb *ElasticSearchBetlog) parseTimeToPartialIndex(indexPrefix string, dateTime time.Time) string {
	dateStr := dateTime.Format(indexDateLayout)
	index := indexPrefix + dateStr
	return index
}
