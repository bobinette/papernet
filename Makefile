install:
	go get -u github.com/blevesearch/bleve
	go get -u github.com/boltdb/bolt
	go get -u github.com/gin-gonic/gin

test:
	go test ./...
