package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"grpc-todo/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetAllTasks(ctx context.Context) ([]*domain.Task, error)
	UpdateTaskStatus(ctx context.Context, id string, status string) error
	DeleteTask(ctx context.Context, id string) error
	DeleteDoneTasks(ctx context.Context) (int64, error)
}

type mongoTask struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Status      string             `bson:"status"`
	CreatedAt   int64              `bson:"created_at"`
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	collection := db.Collection("tasks")
	return &mongoRepository{
		collection: collection,
	}
}

func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *mongoRepository) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	doc := mongoTask{
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %v", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to get inserted ID")
	}

	task.Id = insertedID.Hex()
	return task, nil
}

func (r *mongoRepository) GetAllTasks(ctx context.Context) ([]*domain.Task, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	for cursor.Next(ctx) {
		var mt mongoTask
		if err := cursor.Decode(&mt); err != nil {
			return nil, fmt.Errorf("failed to decode task: %v", err)
		}

		tasks = append(tasks, &domain.Task{
			Id:          mt.ID.Hex(),
			Title:       mt.Title,
			Description: mt.Description,
			Status:      mt.Status,
			CreatedAt:   mt.CreatedAt,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return tasks, nil
}

func (r *mongoRepository) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"status": status}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return nil
}

func (r *mongoRepository) DeleteTask(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	filter := bson.M{"_id": objectID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return nil
}

func (r *mongoRepository) DeleteDoneTasks(ctx context.Context) (int64, error) {
	filter := bson.M{"status": "DONE"}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("error deleting DONE tasks: %v", err)
	}
	return result.DeletedCount, nil
}
