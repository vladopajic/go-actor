# `go-actor` Benchmarks

Here you can find benchmarks of `go-actor` library.


```
go version
go version go1.22.12 linux/amd64
go test -bench=. github.com/vladopajic/go-actor/actor -run=^# -count 5  -benchmem
goos: linux
goarch: amd64
pkg: github.com/vladopajic/go-actor/actor
cpu: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz
BenchmarkActorProducerToConsumer-16              6257289               197.0 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumer-16              6255667               191.5 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumer-16              6292658               187.9 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumer-16              6224390               205.5 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumer-16              6328387               183.9 ns/op             8 B/op          0 allocs/op
BenchmarkMailbox-16                              4006864               437.2 ns/op            14 B/op          0 allocs/op
BenchmarkMailbox-16                              1298016               871.7 ns/op            12 B/op          0 allocs/op
BenchmarkMailbox-16                              2595217               751.7 ns/op            13 B/op          0 allocs/op
BenchmarkMailbox-16                              1747360               909.2 ns/op             9 B/op          0 allocs/op
BenchmarkMailbox-16                              2028043              1126 ns/op              10 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  1477164              1173 ns/op               0 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  1625610               957.4 ns/op             2 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  1413207               826.2 ns/op             3 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  2883579               368.5 ns/op             5 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  2676229               390.7 ns/op             4 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        3844426               309.7 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4264402               917.8 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4005238               566.2 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        1954975               793.2 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        3280045               618.2 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9353541               121.3 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9558877               115.3 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9771594               113.1 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9749954               112.8 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9487732               116.3 ns/op             0 B/op          0 allocs/op
```