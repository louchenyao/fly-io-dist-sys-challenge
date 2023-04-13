# fly-io-dist-sys-challenge

Repo for [fly.io dist-sys challenge](https://fly.io/dist-sys/)

## Notes

### Challenge 4

My first implementation is to use `seq-kv` with a single key `counter`. All add requests will increment the counter in the lock-free manner (using `cas` operation). And read requests will simply read the counter. It failed the test because [sequential consistency](https://jepsen.io/consistency/models/sequential) does not gurentee that all of nodes see the latest state. "A process in a sequentially consistent system may be far ahead, or behind, of other processes" quote from [sequential consistency](https://jepsen.io/consistency/models/sequential).

Using lineerizable kv store `lin-kv` is too overkill for this challenge. So I added the broadcast mechanism which will broadcast the latest counter value to all neighbors periodically. And the read request will simply read the counter from the local cache.

The challenge does not specify "how long" is the "eventual". In theory, the first implementation is technically correct if the "eventual" is infinite.  I found that the test actually waits 10 seconds before the final check. So setting the broadcast interval to 1 seconds is enough to pass the test.

### Challenge 5

For test c, the overall msgs-per-op is 14.768623 which is not the best compared to [other solutions](https://community.fly.io/t/challenge-5c-efficient-kafka-style-log/10971). Because in my solution, all writes are cooridinated by `lin-kv`, nodes does not need to send any message to each other. There are ways to improve the performance I did not implement:

- Assign a leader for each topic. The leader will be responsible for maintaining the offset, so no cas operation is needed.
- Compress consecutive messages into a single message. This will reduce the number of reads.
- Batch commit offsets. Return the `send_ok` message to the client only after the offset is committed. Trade off the latency for throughput.

### Challenge 6

You can see my implementation does not abort any transaction and pass the test. The write set is committed to the database only after all the reads are done. I think it's because **read_committed** is not a strong isolation level and does make too much sense in the real world.