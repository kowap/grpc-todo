package server

import (
	"context"
	"testing"

	"grpc-todo/proto"
)

func TestCreateTask(t *testing.T) {
	s := NewToDoServer()

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
}

func TestGetAllTasks(t *testing.T) {
	s := NewToDoServer()

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
	s := NewToDoServer()

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

	res, err := s.UpdateTaskStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	if res.Task.Status != proto.Status_DONE {
		t.Errorf("Expected status %v, got %v", proto.Status_DONE, res.Task.Status)
	}
}

func TestDeleteTask(t *testing.T) {
	s := NewToDoServer()

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

	res, err := s.GetAllTasks(context.Background(), &proto.GetAllTasksRequest{})
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(res.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(res.Tasks))
	}
}
