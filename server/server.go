package server

import (
	"context"
	"fmt"
	"time"
	"log"

	"grpc-todo/domain"
	"grpc-todo/proto"
	"grpc-todo/repository"

	"github.com/robfig/cron/v3"
)

type ToDoServer struct {
	proto.UnimplementedToDoServiceServer
	repo repository.Repository
}

func NewToDoServer(repo repository.Repository) *ToDoServer {
	return &ToDoServer{repo: repo}
}

func protoStatusToString(status proto.Status) string {
	switch status {
	case proto.Status_TODO:
		return "TODO"
	case proto.Status_IN_PROGRESS:
		return "IN_PROGRESS"
	case proto.Status_DONE:
		return "DONE"
	default:
		return "TODO"
	}
}

func stringToProtoStatus(status string) proto.Status {
	switch status {
	case "TODO":
		return proto.Status_TODO
	case "IN_PROGRESS":
		return proto.Status_IN_PROGRESS
	case "DONE":
		return proto.Status_DONE
	default:
		return proto.Status_TODO
	}
}

func (s *ToDoServer) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	task := &domain.Task{
		Title: req.Title,
		Description: req.Description,
		Status: "TODO",
		CreatedAt: time.Now().Unix(),
	}

	createdTask, err := s.repo.CreateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("CreateTask failed: %v", err)
	}

	return &proto.CreateTaskResponse{
		Task: &proto.Task{
			Id: createdTask.Id,
			Title: createdTask.Title,
			Description: createdTask.Description,
			Status: stringToProtoStatus(createdTask.Status),
			CreatedAt: createdTask.CreatedAt,
		},
	}, nil
}

func (s *ToDoServer) GetAllTasks(ctx context.Context, _ *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	tasks, err := s.repo.GetAllTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllTasks failed: %v", err)
	}

	var protoTasks []*proto.Task
	for _, t := range tasks {
		protoTasks = append(protoTasks, &proto.Task{
			Id: t.Id,
			Title: t.Title,
			Description: t.Description,
			Status: stringToProtoStatus(t.Status),
			CreatedAt: t.CreatedAt,
		})
	}

	return &proto.GetAllTasksResponse{Tasks: protoTasks}, nil
}

func (s *ToDoServer) UpdateTaskStatus(ctx context.Context, req *proto.UpdateTaskStatusRequest) (*proto.UpdateTaskStatusResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	err := s.repo.UpdateTaskStatus(ctx, req.Id, protoStatusToString(req.Status))
	if err != nil {
		return nil, fmt.Errorf("UpdateTaskStatus failed: %v", err)
	}

	return &proto.UpdateTaskStatusResponse{}, nil
}

func (s *ToDoServer) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	err := s.repo.DeleteTask(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("DeleteTask failed: %v", err)
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

	deletedCount, err := s.repo.DeleteDoneTasks(ctx)
	if err != nil {
		log.Printf("Error deleting DONE tasks: %v", err)
		return
	}

	log.Printf("Cron job: Deleted %d DONE tasks", deletedCount)
}