package benchmark

import (
	"context"
	"time"

	v2 "code.vegaprotocol.io/vega/protos/data-node/api/v2"
)

type APITest func(client v2.TradingDataServiceClient) time.Duration

func Worker(ctx context.Context, client v2.TradingDataServiceClient, apiTest APITest, reqCh <-chan struct{}, resultCh chan<- time.Duration, doneCh chan<- struct{}) {
	defer func() {
		doneCh <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-reqCh:
			if !ok {
				return
			}
			elapsed := apiTest(client)
			resultCh <- elapsed
		}
	}
}
