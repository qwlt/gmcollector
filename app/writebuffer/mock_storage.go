package writebuffer

import "github.com/qwlt/gmcollector/app/models"

type MockStorage struct {
}

func (s *MockStorage) Write(data []models.Model) error {
	return nil
}
