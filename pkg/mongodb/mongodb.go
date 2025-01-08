package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseConfig struct {
	URI        string
	DBName     string
	Collection string
}

type MongoDB struct {
	collection *mongo.Collection
}

func (m *MongoDB) ConnectToMongoDB(ctx context.Context, config DatabaseConfig) error {
	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	m.collection = client.Database(config.DBName).Collection(config.Collection)
	return nil
}

func (m *MongoDB) InsertDocument(ctx context.Context, data interface{}) error {
	_, err := m.collection.InsertOne(ctx, data)
	return err
}

func (m *MongoDB) FindDocument(ctx context.Context, filter bson.M, result interface{}) error {
	err := m.collection.FindOne(ctx, filter).Decode(result)
	return err
}

func (m *MongoDB) UpdateDocument(ctx context.Context, filter bson.M, update interface{}) error {
	_, err := m.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	return err
}

func (m *MongoDB) DeleteDocument(ctx context.Context, filter bson.M) error {
	_, err := m.collection.DeleteOne(ctx, filter)
	return err
}
