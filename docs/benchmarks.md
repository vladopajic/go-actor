# `go-actor` Benchmarks

Here you can find benchmarks for the `go-actor` library.
To run benchmarks locally, use the `make benchmark` command.

```
go version
go version go1.22.12 linux/amd64
go test -bench=. github.com/vladopajic/go-actor/actor -run=^# -count 5  -benchmem
goos: linux
goarch: amd64
pkg: github.com/vladopajic/go-actor/actor
cpu: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz
BenchmarkActorProducerToConsumerMailbox-16       7818805               157.7 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerMailbox-16       7624102               150.6 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerMailbox-16       7796350               149.6 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerMailbox-16       7522546               152.1 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerMailbox-16       7701626               148.7 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerChannel-16      16145071               112.6 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerChannel-16      15803265                92.71 ns/op            8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerChannel-16      14075094               125.7 ns/op             8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerChannel-16      14202432                87.32 ns/op            8 B/op          0 allocs/op
BenchmarkActorProducerToConsumerChannel-16      14893779                99.48 ns/op            8 B/op          0 allocs/op
BenchmarkNativeProducerToConsumer-16            15541105                91.34 ns/op            8 B/op          0 allocs/op
BenchmarkNativeProducerToConsumer-16            17050603               101.3 ns/op             8 B/op          0 allocs/op
BenchmarkNativeProducerToConsumer-16            14693918                69.90 ns/op            8 B/op          0 allocs/op
BenchmarkNativeProducerToConsumer-16            14200236                84.82 ns/op            8 B/op          0 allocs/op
BenchmarkNativeProducerToConsumer-16            15024007                76.14 ns/op            8 B/op          0 allocs/op
BenchmarkMailbox-16                              3209701               603.9 ns/op            12 B/op          0 allocs/op
BenchmarkMailbox-16                              2559504               534.9 ns/op             7 B/op          0 allocs/op
BenchmarkMailbox-16                              2535686               718.5 ns/op             7 B/op          0 allocs/op
BenchmarkMailbox-16                              3135724               438.5 ns/op            12 B/op          0 allocs/op
BenchmarkMailbox-16                              2795941               745.3 ns/op             9 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3514538               326.3 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  2684362               570.2 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3059289               499.1 ns/op             1 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3051763               403.7 ns/op             1 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  2872962               622.2 ns/op             2 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4372791               272.0 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4828120               361.1 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        3621937               442.6 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        2096506               534.8 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4394918               592.5 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16           10698741               113.5 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            8980978               111.9 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9154912               111.1 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9057835               112.5 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9076858               113.3 ns/op             0 B/op          0 allocs/op
```