package storage

import (
	"context"
	"time"

	"justycrawler/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoTimeout = 10 * time.Second
)

// MongoStorage — реализация интерфейса Storage для MongoDB.
type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoStorage создает и возвращает новый экземпляр MongoStorage.
func NewMongoStorage(ctx context.Context, uri, dbName, collectionName string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(ctx, mongoTimeout)
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

// Save реализует метод сохранения данных в MongoDB.
func (s *MongoStorage) Save(ctx context.Context, data domain.CrawledData) error {
	filter := bson.M{"url": data.URL}
	update := bson.M{"$set": data}
	opts := options.Update().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Close закрывает соединение с MongoDB.
func (s *MongoStorage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
