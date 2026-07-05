package parse

import (
	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

func UUIDPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return &id
	}
	return nil
}

func TaskStatus(s string) *domain.TaskStatus {
	if s == "" {
		return nil
	}
	st := domain.TaskStatus(s)
	if st.IsValid() {
		return &st
	}
	return nil
}
