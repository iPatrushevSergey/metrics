package filestorage

import (
	"encoding/json"
	"os"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

type metricFileDTO struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func metricToDTO(m model.Metric) metricFileDTO {
	return metricFileDTO{
		ID:    m.ID,
		MType: m.MType,
		Delta: m.Delta,
		Value: m.Value,
		Hash:  m.Hash,
	}
}

func dtoToMetric(dto metricFileDTO) model.Metric {
	return model.Metric{
		ID:    dto.ID,
		MType: dto.MType,
		Delta: dto.Delta,
		Value: dto.Value,
		Hash:  dto.Hash,
	}
}

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

	var metricList []metricFileDTO
	for _, m := range metrics {
		metricList = append(metricList, metricToDTO(m))
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

	var dtos []metricFileDTO
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&dtos); err != nil {
		return nil, err
	}

	metrics := make([]model.Metric, 0, len(dtos))
	for _, dto := range dtos {
		metrics = append(metrics, dtoToMetric(dto))
	}

	return metrics, nil
}
