package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
)

type Person struct {
	Name  string `bson:"name"`
	Age   int    `bson:"age"`
	Email string `bson:"email"`
}

func TestMongoDB(t *testing.T) {
	config := DatabaseConfig{
		URI:        "mongodb://localhost:27017",
		DBName:     "mydb",
		Collection: "mycollection",
	}

	mongoDB := MongoDB{}
	err := mongoDB.ConnectToMongoDB(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	// 插入文档
	person := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}
	err = mongoDB.InsertDocument(context.Background(), person)
	if err != nil {
		t.Fatal(err)
	}

	filter := bson.M{"name": "Alice"}

	// 查询文档
	var result Person
	err = mongoDB.FindDocument(context.Background(), filter, &result)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Query Result:", result)

	// 更新文档
	update := Person{Name: "Alice", Age: 31, Email: "new_email@example.com"}
	err = mongoDB.UpdateDocument(context.Background(), filter, update)
	if err != nil {
		t.Fatal(err)
	}

	// 删除文档
	err = mongoDB.DeleteDocument(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
}
