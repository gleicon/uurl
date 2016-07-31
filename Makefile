include Makefile.defs

all: deps server

test:
	go test -v


deps:
	go get -v

server:
	go build -v -o $(NAME) -ldflags "-X main.VERSION=$(VERSION)"

deploy:
	$(FLAGS_LINUX_AMD64) go build -v -o $(NAME) -ldflags "-X main.VERSION=$(VERSION)"
	file $(NAME)

clean:
	rm -f $(NAME)

.PHONY: server
