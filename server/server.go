package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"grpc-todo/proto"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ToDoServer struct {
	proto.UnimplementedToDoServiceServer
	mongoDatabase   *mongo.Database
	mongoCollection *mongo.Collection
}

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Status      proto.Status       `bson:"status"`
	CreatedAt   int64              `bson:"created_at"`
}

func NewToDoServer(db *mongo.Database) *ToDoServer {
	collection := db.Collection("tasks")
	return &ToDoServer{
		mongoDatabase:   db,
		mongoCollection: collection,
	}
}

func (s *ToDoServer) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	task := &Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      proto.Status_TODO,
		CreatedAt:   time.Now().Unix(),
	}

	result, err := s.mongoCollection.InsertOne(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %v", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to get inserted ID")
	}

	protoTask := &proto.Task{
		Id:          insertedID.Hex(),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
	}

	return &proto.CreateTaskResponse{Task: protoTask}, nil
}

func (s *ToDoServer) GetAllTasks(ctx context.Context, _ *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cursor, err := s.mongoCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}
	defer cursor.Close(ctx)

	var tasks []*proto.Task
	for cursor.Next(ctx) {
		var taskData Task
		if err := cursor.Decode(&taskData); err != nil {
			return nil, fmt.Errorf("failed to decode task: %v", err)
		}

		protoTask := &proto.Task{
			Id:          taskData.ID.Hex(),
			Title:       taskData.Title,
			Description: taskData.Description,
			Status:      taskData.Status,
			CreatedAt:   taskData.CreatedAt,
		}

		tasks = append(tasks, protoTask)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return &proto.GetAllTasksResponse{Tasks: tasks}, nil
}

func (s *ToDoServer) UpdateTaskStatus(ctx context.Context, req *proto.UpdateTaskStatusRequest) (*proto.UpdateTaskStatusResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %v", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"status": req.Status}}

	result, err := s.mongoCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("task with ID %s not found", req.Id)
	}

	return &proto.UpdateTaskStatusResponse{}, nil
}

func (s *ToDoServer) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %v", err)
	}

	filter := bson.M{"_id": objectID}

	result, err := s.mongoCollection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete task: %v", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("task with ID %s not found", req.Id)
	}

	return &proto.DeleteTaskResponse{}, nil
}

func (s *ToDoServer) StartCronJob() *cron.Cron {
	c := cron.New()
	c.AddFunc("@every 1m", func() {
		log.Println("Cron job started: Deleting DONE tasks")
		s.deleteDoneTasks()
	})
	c.Start()
	return c
}

func (s *ToDoServer) deleteDoneTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"status": proto.Status_DONE}
	result, err := s.mongoCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("Error deleting DONE tasks: %v", err)
		return
	}

	log.Printf("Cron job: Deleted %d DONE tasks", result.DeletedCount)
}
