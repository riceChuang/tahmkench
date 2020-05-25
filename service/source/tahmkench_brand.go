package source

import (
	"context"
	"github.com/riceChuang/tahmkench/types"
)

type TahmkenchBrand struct {
	Worker int
}

func (b *TahmkenchBrand) Initialize(setting interface{}) (err error) {
	return
}
func (b *TahmkenchBrand) CollectionSource(ctx context.Context, recordQueue *chan types.Record) {
	for i := 0; i <= b.Worker; i++ {
		go func() {
			for record := range types.BrandQueue {
				*recordQueue <- record
			}
		}()
	}
	return
}
