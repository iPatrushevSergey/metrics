package filestorage

import (
	"encoding/json"
	"os"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

type FileStorage struct {
	FilePath string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		FilePath: path,
	}
}

func (fs *FileStorage) Save(metrics map[string]model.Metric) error {
	file, err := os.OpenFile(fs.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var metricList []model.Metric
	for _, m := range metrics {
		metricList = append(metricList, m)
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(metricList)
}

func (fs *FileStorage) Load() ([]model.Metric, error) {
	file, err := os.OpenFile(fs.FilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metrics []model.Metric
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
