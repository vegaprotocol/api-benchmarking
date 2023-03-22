package orders

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/datanode/ratelimit"
	v2 "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"github.com/spf13/cobra"
	"github.com/vegaprotocol/datanode-api-benchmarking/benchmark"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ListOrdersCmd = &cobra.Command{
		Use:   "ListOrders",
		Short: "Test orders API",
		Long:  "Run tests against the orders APIs",
		Run:   run,
	}

	urls       []string
	marketID   string
	partyID    string
	reference  string
	timeout    time.Duration
	iterations int
	workers    int
	queryCount int
	startDate  string
	endDate    string
)

func run(cmd *cobra.Command, args []string) {
	log.Printf("Benchmarking ListOrders for: %v", urls)

	for _, url := range urls {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()), ratelimit.WithSecret())
		if err != nil {
			log.Printf("could not corrent to %s, %v", url, err)
			continue
		}
		client := v2.NewTradingDataServiceClient(conn)

		for i := 0; i < iterations; i++ {
			reqCh := make(chan struct{})
			resultCh := make(chan time.Duration)
			doneCh := make(chan struct{})

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			for i := 0; i < workers; i++ {
				go benchmark.Worker(ctx, client, ListOrders, reqCh, resultCh, doneCh)
			}

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				for i := 0; i < queryCount; i++ {
					reqCh <- struct{}{}
					// slight pause to avoid hitting the rate limiter
					time.Sleep(time.Millisecond * 5)
				}
				close(reqCh)
			}()

			metrics := make([]time.Duration, 0)

			go func() {
				count := 0
				for {
					select {
					case m := <-resultCh:
						count++
						if m > 0 {
							metrics = append(metrics, m)
						}
					case <-doneCh:
						close(resultCh)
						wg.Done()
						return
					}
				}
			}()
			wg.Wait()
			mean := mean(metrics)
			median := median(metrics)

			log.Printf(", %d, %s, %v, %v", i+1, url, mean.Milliseconds(), median.Milliseconds())

			cancel()
		}
	}
}

func init() {
	ListOrdersCmd.Flags().StringSliceVarP(&urls, "url", "u", []string{"localhost:3007"}, "URL of the data node API endpoint to use")
	ListOrdersCmd.Flags().StringVarP(&marketID, "market", "m", "", "UUID of market to query orders for")
	ListOrdersCmd.Flags().StringVarP(&partyID, "party", "p", "", "UUID of Party to query orders for")
	ListOrdersCmd.Flags().StringVarP(&reference, "reference", "r", "", "Status to query orders for")
	ListOrdersCmd.Flags().DurationVarP(&timeout, "timeout", "t", time.Minute, "Timeout for each benchmark test")
	ListOrdersCmd.Flags().IntVarP(&iterations, "iterations", "i", 1, "Number of iterations to run")
	ListOrdersCmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of concurrent workers to use")
	ListOrdersCmd.Flags().IntVarP(&queryCount, "query-count", "q", 100, "Number of queries to run per iteration")
	ListOrdersCmd.Flags().StringVarP(&startDate, "start-date", "s", "", "Start date for the date range to use in the query")
	ListOrdersCmd.Flags().StringVarP(&endDate, "end-date", "e", "", "Start date for the date range to use in the query")

	if startDate != "" {
		_, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			log.Fatalf("could not parse start date, please use RFC3339 format, %v", err)
		}
	}

	if endDate != "" {
		_, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			log.Fatalf("could not parse end date, please use RFC3339 format, %v", err)
		}
	}
}

func mean(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	if len(durations) == 1 {
		return durations[0]
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func median(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	if len(durations) == 1 {
		return durations[0]
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	middle := len(durations) / 2
	if len(durations)%2 == 0 {
		return (durations[middle-1] + durations[middle]) / 2
	}

	middle = len(durations)/2 + 1
	return durations[middle]
}

func ListOrders(client v2.TradingDataServiceClient) time.Duration {
	start := time.Now()
	var market *string
	var party *string
	var ref *string

	if marketID != "" {
		market = &marketID
	}

	if partyID != "" {
		party = &partyID
	}

	if reference != "" {
		ref = &reference
	}

	var dateRangeStart, dateRangeEnd *int64

	if startDate != "" {
		s, _ := time.Parse(time.RFC3339, startDate)
		snanos := s.UnixNano()
		dateRangeStart = &snanos
	}

	if endDate != "" {
		e, _ := time.Parse(time.RFC3339, endDate)
		enanos := e.UnixNano()
		dateRangeEnd = &enanos
	}

	var dateRange *v2.DateRange
	if dateRangeStart != nil || dateRangeEnd != nil {
		dateRange = &v2.DateRange{
			StartTimestamp: dateRangeStart,
			EndTimestamp:   dateRangeEnd,
		}
	}

	_, err := client.ListOrders(context.Background(), &v2.ListOrdersRequest{
		PartyId:   party,
		MarketId:  market,
		Reference: ref,
		DateRange: dateRange,
	})
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("could not list orders: %v", err)
	}
	return elapsed
}
