# Event Study Implementation

Event Study implementation in Go with [Tinkoff Open API sdk](https://github.com/TinkoffCreditSystems/invest-openapi-go-sdk)

## Installation

```shell script
go install github.com/gaarutyunov/eventstudy/cmd/eventstudy
```

## Usage

```shell script
Usage of eventstudy:

  -apiKey string
        api key
  -benchmark string
        benchmark market index (default "MOEX")
  -delimiter string
        delimiter for output file (default ",")
  -event string
        event date (default "2020-03-06")
  -eventWindow duration
        event window (default 720h0m0s)
  -from string
        starting date (default "2019-01-01")
  -h    help
  -out string
        output file (default "out/returns.csv")
  -period int
        period of estimation (default 30)
  -rf float
        risk free rate (default 0.045)
  -ticker string
        instrument ticker (default "TCS")
  -to string
        ending date (default "2020-10-04")
  -window int
        rolling window (default 30)
```


## References

- MacKinlay, A. Craig. “Event Studies in Economics and Finance.” Journal of Economic Literature, vol. 35, no. 1, 1997, pp. 13–39. JSTOR, www.jstor.org/stable/2729691. Accessed 4 Oct. 2020.