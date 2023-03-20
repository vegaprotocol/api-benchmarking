# Vega API benchmarking tool

Small tool to benchmark the Vega Data Node API's performance. Tests are performed against the gRPC API back-end as this back-end
also provides the data returned from the REST and GraphQL APIs.

## Building the application

A Makefile is provided to build the application. To build the application, run:

```bash
make
```

This will create a binary file called `api-benchmark` in the project root directory.

## Usage

To use the benchmarking tool, run:

```bash
./api-benchmark [API function to test] [args]
```

The following API functions are supported:

- ListOrders
- TBD

To get help on the usage of the tool, run:

```bash
./api-benchmark --help
```

To get help regarding the parameters you can pass each API function, run:

```bash
./api-benchmark [API function to test] --help
```

### ListOrders

The ListOrders API function allows you to retrieve a list of orders from the Vega network. This API function supports the following parameters:

- `-e`, `--end-date` string.     Start date for the date range to use in the query
- `-i`, `--iterations` int.      Number of iterations to run (default 1)
- `-m`, `--market` string.       UUID of market to query orders for
- `-p`, `--party` string.        UUID of Party to query orders for
- `-q`, `--query-count` int.     Number of queries to run per iteration (default 100)
- `-r`, `--reference` string.    Status to query orders for
- `-s`, `--start-date` string.   Start date for the date range to use in the query
- `-t`, `--timeout` duration.    Timeout for each benchmark test (default 1m0s)
- `-u`, `--url` strings.         URL of the data node API endpoint to use (default [localhost:3007])
- `-w`, `--workers` int.         Number of concurrent workers to use (default 1)

To benchmark the ListOrders API function, run:

```bash
./api-benchmark ListOrders
```

By default this will attempt to benchmark the ListOrders API on the machine it is currently running on.

To benchmark a list of servers, run:

```bash
./api-benchmark ListOrder -u host1:port1,host2:port2,host3:port3
```

Each iteration will create `query-count` number of calls to the ListOrders API for the listed servers.
The timing for each request will be collected and an average time will be calculated for each iteration.
The average time for each iteration will be displayed at the end of each iteration.

> Note: In the event of an error, the benchmarking tool will exit immediately.

By default the number of workers is set to 1. This means that the benchmarking tool will only make one request at a time.
To test the performance when multiple requests are made at the same time, you can increase the number of workers.

Each iteration will execute for a maximum of 1 minute. If the number of queries per iteration is not reached, the average time is calculated using the results
that were obtained in the iteration. This prevents the tests for running overly long should the endpoint be slow. To extend the timeout for each iteration,
use the `-t` or `--timeout` flag.

## TODO:

Add support for other API functions as required
