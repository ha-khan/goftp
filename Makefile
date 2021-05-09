build: cmd/goftp/main.go
	go build cmd/goftp/main.go
	cp ./main ./bin
	rm ./main

test:
	go test -v

clean:
	rm ./bin/* 

image:
	build