package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// database implements the database connection for the graphql experiment. It
// is implemented quite sloppily because all references are hard-coded, the
// mongo context is kept at 'TODO()' and all properties are public.

// Context is the collection of references to the database
type Context struct {
	Client     *mongo.Client
	Context    context.Context
	MaxResults int64
}

// NewContext connects to the mongo database and ensures all documents
// and indices are there.
func NewContext() (*Context, error) {

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	// Compose result
	context := Context{
		Client:     client,
		Context:    context.TODO(),
		MaxResults: 250,
	}

	return &context, nil
}
