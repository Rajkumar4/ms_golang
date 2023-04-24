package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Model struct {
	Logger *Logger
}

type Logger struct {
	ID        string    `bson:"_id",omitempty json:"id",omitempty`
	Name      string    `bson:"name",omitempty`
	Data      string    `bson:"data",omitempty`
	CreatedAt time.Time `bson:"created_at",omitempty`
	UpdatedAt time.Time `bson:"updated_at",omitempty`
}

func New(mongo *mongo.Client) *Model {
	client = mongo
	return &Model{
		Logger: &Logger{},
	}
}

func (l *Logger) Insert() error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), l)
	if err != nil {
		log.Printf("Failed to insert data: %s", err.Error())
		return err
	}
	return nil
}

func (l *Logger) All() ([]Logger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})
	cur, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Printf("Failed to find record: %s", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	var itemList []Logger

	for cur.Next(ctx) {
		var item Logger
		err := cur.Decode(item)
		if err != nil {
			log.Printf("Failed to get item: %s", err.Error())
			return nil, err
		}
		itemList = append(itemList, item)
	}
	return itemList, nil
}

func (l *Logger) GetOne(id string) (*Logger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed to generagte Doc ID: %s", err.Error())
		return nil, err
	}
	var out *Logger
	collection.FindOne(ctx, bson.M{"_id": docId}).Decode(out)
	return out, nil
}

func (l *Logger) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	if err := collection.Drop(ctx); err != nil {
		log.Printf("Failed to drop collection: %s", err.Error())
		return err
	}
	return nil
}

func (l *Logger) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docId, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		log.Printf("Failed to generagte Doc ID: %s", err.Error())
		return nil, err
	}
	result, err := collection.UpdateByID(ctx,
		bson.M{"_id": docId},
		bson.D{{
			"$set", bson.D{
				{"name", l.Name},
				{"data", l.Data},
				{"updated_at", time.Now()},
			}},
		},
	)
	if err != nil {
		log.Printf("Failed to update document: %s", err.Error())
		return nil, err
	}

	return result, nil
}
