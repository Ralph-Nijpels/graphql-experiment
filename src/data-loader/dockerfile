FROM golang:latest

ENV GOPATH /projects/graphql-experiment/src/data-loader

WORKDIR /projects/graphql-experiment/src/data-loader

COPY . .

# Get the dependencies
RUN go get github.com/gorilla/mux
RUN go get go.mongodb.org/mongo-driver/mongo
RUN go get github.com/minio/minio-go
RUN go get github.com/graphql-go/graphql

# Create the dataloader
RUN go build -o data-loader main.go

# And set it to go
CMD ["./data-loader/data-loader"]