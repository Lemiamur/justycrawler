package crawler

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoStorage(ctx context.Context, uri, dbName, collectionName string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)

	return &MongoStorage{
		client:     client,
		collection: collection,
	}, nil
}

func (s *MongoStorage) Save(ctx context.Context, data CrawledData) error {
	_, err := s.collection.InsertOne(ctx, data)
	return err
}

func (s *MongoStorage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
