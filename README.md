# fly-io-dist-sys-challenge

Repo for [fly.io dist-sys challenge](https://fly.io/dist-sys/)

## Notes

### Challenge 4

My first implementation is to use `seq-kv` with a single key `counter`. All add requests will increment the counter in the lock-free manner (using `cas` operation). And read requests will simply read the counter. It failed the test because [sequential consistency](https://jepsen.io/consistency/models/sequential) does not gurentee that all of nodes see the latest state. "A process in a sequentially consistent system may be far ahead, or behind, of other processes" quote from [sequential consistency](https://jepsen.io/consistency/models/sequential).
