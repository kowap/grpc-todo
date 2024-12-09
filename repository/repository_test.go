package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"grpc-todo/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database("grpc_todo_test_db")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Drop(ctx); err != nil {
			t.Fatalf("Failed to drop test database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Fatalf("Failed to disconnect MongoDB: %v", err)
		}
	}

	return db, cleanup
}

func TestRepository_CreateTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	task := &domain.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "TODO",
		CreatedAt:   time.Now().Unix(),
	}

	createdTask, err := repo.CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if createdTask.Id == "" {
		t.Errorf("Expected task to have an ID, got empty string")
	}

	collection := db.Collection("tasks")
	objID, err := primitive.ObjectIDFromHex(createdTask.Id)
	if err != nil {
		t.Fatalf("Invalid ID: %v", err)
	}

	var doc bson.M
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		t.Fatalf("Failed to find inserted task: %v", err)
	}
}

func TestRepository_GetAllTasks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	_, err := repo.CreateTask(context.Background(), &domain.Task{
		Title:       "Task 1",
		Description: "Description 1",
		Status:      "TODO",
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	tasks, err := repo.GetAllTasks(context.Background())
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
}

func TestRepository_UpdateTaskStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	createdTask, err := repo.CreateTask(context.Background(), &domain.Task{
		Title:       "Task Update",
		Description: "To be updated",
		Status:      "TODO",
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = repo.UpdateTaskStatus(context.Background(), createdTask.Id, "DONE")
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	tasks, err := repo.GetAllTasks(context.Background())
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Status != "DONE" {
		t.Errorf("Expected status DONE, got %s", tasks[0].Status)
	}
}

func TestRepository_DeleteTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	createdTask, err := repo.CreateTask(context.Background(), &domain.Task{
		Title:       "To be deleted",
		Description: "Will be removed",
		Status:      "TODO",
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = repo.DeleteTask(context.Background(), createdTask.Id)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	tasks, err := repo.GetAllTasks(context.Background())
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

func TestRepository_DeleteDoneTasks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	_, err := repo.CreateTask(context.Background(), &domain.Task{
		Title:       "Done Task",
		Description: "This is done",
		Status:      "DONE",
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	_, err = repo.CreateTask(context.Background(), &domain.Task{
		Title:       "Todo Task",
		Description: "This is todo",
		Status:      "TODO",
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	deletedCount, err := repo.DeleteDoneTasks(context.Background())
	if err != nil {
		t.Fatalf("DeleteDoneTasks failed: %v", err)
	}

	if deletedCount != 1 {
		t.Errorf("Expected to delete 1 DONE task, got %d", deletedCount)
	}

	tasks, err := repo.GetAllTasks(context.Background())
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 remaining task, got %d", len(tasks))
	}

	if tasks[0].Status == "DONE" {
		t.Errorf("Expected no DONE tasks remaining")
	}
}
