package server

import (
	"context"
	"testing"
	"time"

	"grpc-todo/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestMongoDB(t *testing.T) (*mongo.Database, func()) {
	mongoURI := "mongodb://localhost:27017" // Используйте ваш URI, если он отличается
	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	db := client.Database("grpc_todo_test_db")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := db.Drop(ctx)
		if err != nil {
			t.Fatalf("Failed to drop test database: %v", err)
		}
		err = client.Disconnect(ctx)
		if err != nil {
			t.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	}

	return db, cleanup
}

func TestCreateTask(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	s := NewToDoServer(db)

	req := &proto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "This is a test task",
	}

	res, err := s.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if res.Task.Title != req.Title || res.Task.Description != req.Description {
		t.Errorf("Expected task with title %s and description %s, got %s and %s",
			req.Title, req.Description, res.Task.Title, res.Task.Description)
	}

	if res.Task.Id == "" {
		t.Errorf("Expected task to have an ID, got empty string")
	}

	collection := db.Collection("tasks")
	var result Task
	objectID, err := primitive.ObjectIDFromHex(res.Task.Id)
	if err != nil {
		t.Fatalf("Invalid task ID: %v", err)
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to find task in database: %v", err)
	}
}

func TestGetAllTasks(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	s := NewToDoServer(db)

	_, err := s.CreateTask(context.Background(), &proto.CreateTaskRequest{
		Title:       "Task 1",
		Description: "First task",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	res, err := s.GetAllTasks(context.Background(), &proto.GetAllTasksRequest{})
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(res.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(res.Tasks))
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	s := NewToDoServer(db)

	createRes, err := s.CreateTask(context.Background(), &proto.CreateTaskRequest{
		Title:       "Task 1",
		Description: "First task",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	req := &proto.UpdateTaskStatusRequest{
		Id:     createRes.Task.Id,
		Status: proto.Status_DONE,
	}

	_, err = s.UpdateTaskStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	collection := db.Collection("tasks")
	var task Task
	objectID, err := primitive.ObjectIDFromHex(createRes.Task.Id)
	if err != nil {
		t.Fatalf("Invalid task ID: %v", err)
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		t.Fatalf("Failed to find task: %v", err)
	}

	if task.Status != proto.Status_DONE {
		t.Errorf("Expected status DONE, got %v", task.Status)
	}
}

func TestDeleteTask(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	s := NewToDoServer(db)

	createRes, err := s.CreateTask(context.Background(), &proto.CreateTaskRequest{
		Title:       "Task 1",
		Description: "First task",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	_, err = s.DeleteTask(context.Background(), &proto.DeleteTaskRequest{
		Id: createRes.Task.Id,
	})
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	collection := db.Collection("tasks")
	objectID, err := primitive.ObjectIDFromHex(createRes.Task.Id)
	if err != nil {
		t.Fatalf("Invalid task ID: %v", err)
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&Task{})
	if err == nil {
		t.Fatalf("Expected task to be deleted, but it was found in the database")
	}
}
