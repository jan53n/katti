.PHONY: test bench

test:
	go test . -v

bench:
	go test -bench=. | tee BENCHMARK

doc:
	go doc -http