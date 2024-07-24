package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading environment variable file")
	}

	//    PORT:= getting the port number from the environment variable
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000" // Default port if not set
	}
	//    MONGODB_URL:= getting the mongodb url from the environment variable
	MONGODB_URL := os.Getenv("MONGODB_URL")
	if MONGODB_URL == "" {
		log.Fatal("Unable to set MONGODB_URL environment")
	}

	clientOptions := options.Client().ApplyURI(MONGODB_URL)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MONGODB ATLAS")

	// once the main function done with running we will DisConnect the connection
	defer client.Disconnect(context.Background())

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	fmt.Printf("App is running on port %v\n", PORT)
	app.Listen(":" + PORT)
}

func getTodos(c *fiber.Ctx) error {

	todos := []Todo{}

	// todos = [
	// 	{
	// 		ID:'',
	// 		Completed:false,
	// 		Body:"ds"
	// 	},
	// 	{
	// 		ID:'',
	// 		Completed:false,
	// 		Body:"dgus"
	// 	}
	// ]

	filters := bson.M{}

	cursor, err := collection.Find(context.Background(), filters)

	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Println(cursor)

	// close the connection with the cursor as it frees the resources
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		todo := Todo{}

		fmt.Println(todo)
		err := cursor.Decode(&todo)
		if err != nil {
			log.Fatal(err)
			return err
		}

		fmt.Println(todo)

		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)

}

func createTodo(c *fiber.Ctx) error {
	todo := &Todo{}

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(404).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	// First Approach to insert the id before creating the todo
	// todo.ID = primitive.NewObjectID()

	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	// Second approach to insert the id after creating the todo
	todo.ID = result.InsertedID.(primitive.ObjectID)

	fmt.Println(result)

	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {

	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	filters := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filters, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}

func deleteTodo(c *fiber.Ctx) error {
	// collect the id from the params
	// id will be type string so convert it to primitive.ObjectID type and then add it to filter

	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	filters := bson.M{"_id": objectID}

	_, err = collection.DeleteOne(context.Background(), filters)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}
