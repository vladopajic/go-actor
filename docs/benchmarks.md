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
BenchmarkActorConsumerToProducer-16              6230671               188.4 ns/op             8 B/op          0 allocs/op
BenchmarkActorConsumerToProducer-16              6039529               199.3 ns/op             8 B/op          0 allocs/op
BenchmarkActorConsumerToProducer-16              6087484               188.7 ns/op             8 B/op          0 allocs/op
BenchmarkActorConsumerToProducer-16              5946004               198.0 ns/op             8 B/op          0 allocs/op
BenchmarkActorConsumerToProducer-16              5867820               189.4 ns/op             8 B/op          0 allocs/op
BenchmarkMailbox-16                              3355429               457.4 ns/op            18 B/op          0 allocs/op
BenchmarkMailbox-16                              3396717               369.5 ns/op            16 B/op          0 allocs/op
BenchmarkMailbox-16                              3343170               645.9 ns/op            14 B/op          0 allocs/op
BenchmarkMailbox-16                              3224894               368.7 ns/op            20 B/op          0 allocs/op
BenchmarkMailbox-16                              2717319               401.0 ns/op            20 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3969968               587.6 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3325489               321.4 ns/op             4 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  4037344               443.1 ns/op             3 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3961963               341.4 ns/op             4 B/op          0 allocs/op
BenchmarkMailboxWithLargeCap-16                  3359266               383.6 ns/op             4 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        4139724               412.8 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        2125810               653.8 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        2221368               719.0 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        2568808               660.0 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChan-16                        2935008               894.7 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9582193               115.1 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9603475               117.6 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9391846               125.6 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9724878               116.2 ns/op             0 B/op          0 allocs/op
BenchmarkMailboxAsChanWithLargeCap-16            9827776               122.4 ns/op             0 B/op          0 allocs/op
```