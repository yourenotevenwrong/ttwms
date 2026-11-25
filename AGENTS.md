This is a production server. Treat it as such. When significant changes are made, make sure to keep backups of critical data in order to restore service as fast as possible in the event of problems. We want to ship fast, but there is only so much room for breaking things.

Preferred stack is a fast, simple and lightweight mix of go, postgres, and htmx. Do not waste time on unit tests. You can use them when you are debugging issues but prefer end to end tests and integration tests. They catch failures across HTTP, database, and rendering boundaries that isolated tests can miss. Avoid using mock data. Code coverage is not important.

Avoid unintended runtime work. In hot paths, eliminate needless heap allocations, escapes, boxing, GC pressure, pointer chasing, and goroutines. Measure before claiming a speedup.

- Identify the hot path, realistic data sizes, and latency and allocation targets. Benchmark with `go test -bench=. -benchmem -count=5`; profile with `pprof` or trace as needed; inspect escapes with `go build -gcflags="-m -m"`; compare changes with `benchstat`.
- Write direct, typed Go in hot paths. Preallocate and reuse slices and maps; favor contiguous data and indices when they fit the access pattern. Avoid `any`, reflection, growth or allocation in loops, costly formatting and parsing, and unnecessary indirection. Use `strconv` and `strings.Cut` where useful. Pool frequently reused short-lived objects only, and clear references before `Put`. Bound concurrency; avoid goroutine-per-item and channel ping-pong. Use `unsafe` only when measured and justified.
- Keep cold paths simple and readable. Use precise names, explicit validation and errors, meaningful constants, and comments that explain why. Match local style. Keep related behavior, data, cleanup, and tests close; avoid needless abstraction, but centralize correctness-critical rules. Correctness comes first; measured hot-path performance takes priority over aesthetics.

Development workflow: use Go, postgres and htmx.
