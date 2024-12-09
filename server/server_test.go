package server

import (
	"context"
	"testing"

	"grpc-todo/domain"
	"grpc-todo/proto"
	"grpc-todo/repository"
)

type mockRepository struct {
	tasks map[string]*domain.Task
}

func newMockRepository() repository.Repository {
	return &mockRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockRepository) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	task.Id = "mock_id_" + task.Title
	m.tasks[task.Id] = task
	return task, nil
}

func (m *mockRepository) GetAllTasks(ctx context.Context) ([]*domain.Task, error) {
	var res []*domain.Task
	for _, t := range m.tasks {
		res = append(res, t)
	}
	return res, nil
}

func (m *mockRepository) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	t, ok := m.tasks[id]
	if !ok {
		return repository.ErrNotFound
	}
	t.Status = status
	return nil
}

func (m *mockRepository) DeleteTask(ctx context.Context, id string) error {
	_, ok := m.tasks[id]
	if !ok {
		return repository.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockRepository) DeleteDoneTasks(ctx context.Context) (int64, error) {
	var count int64
	for id, t := range m.tasks {
		if t.Status == "DONE" {
			delete(m.tasks, id)
			count++
		}
	}
	return count, nil
}

func TestCreateTask(t *testing.T) {
	repo := newMockRepository()
	s := NewToDoServer(repo)

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
}

func TestGetAllTasks(t *testing.T) {
	repo := newMockRepository()
	s := NewToDoServer(repo)

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
	repo := newMockRepository()
	s := NewToDoServer(repo)

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

func TestDeleteTask(t *testing.T) {
	repo := newMockRepository()
	s := NewToDoServer(repo)

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

	tasks, err := repo.GetAllTasks(context.Background())
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}
