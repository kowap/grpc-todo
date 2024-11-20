package server

import (
	"context"
	"fmt"
	"sync"

	"grpc-todo/proto"
)

type ToDoServer struct {
	proto.UnimplementedToDoServiceServer
	mu     sync.Mutex
	tasks  map[int32]*proto.Task
	nextID int32
}

func NewToDoServer() *ToDoServer {
	return &ToDoServer{
		tasks:  make(map[int32]*proto.Task),
		nextID: 1,
	}
}

func (s *ToDoServer) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task := &proto.Task{
		Id:          s.nextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      proto.Status_TODO,
	}
	s.tasks[s.nextID] = task
	s.nextID++

	return &proto.CreateTaskResponse{Task: task}, nil
}

func (s *ToDoServer) GetAllTasks(ctx context.Context, _ *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]*proto.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return &proto.GetAllTasksResponse{Tasks: tasks}, nil
}

func (s *ToDoServer) UpdateTaskStatus(ctx context.Context, req *proto.UpdateTaskStatusRequest) (*proto.UpdateTaskStatusResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[req.Id]
	if !exists {
		return nil, fmt.Errorf("task with ID %d not found", req.Id)
	}
	task.Status = req.Status
	return &proto.UpdateTaskStatusResponse{Task: task}, nil
}

func (s *ToDoServer) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[req.Id]
	if !exists {
		return nil, fmt.Errorf("task with ID %d not found", req.Id)
	}
	delete(s.tasks, req.Id)
	return &proto.DeleteTaskResponse{}, nil
}
