darwin-build:
	mkdir -p example/darwin/build/golang-sdk-example && \
	cd example && \
	go build -o darwin/build/golang-sdk-example/golang-sdk-example main.go && \
	cd darwin && \
	cp entry.tp start.sh build/golang-sdk-example && \
	cd build && \
	zip -r -X golang-sdk-example.tpp golang-sdk-example
.PHONY: darwin-build

generate:
	go generate ./...
.PHONY: generate

clean:
	rm -Rvf example/**/build
.PHONY: clean