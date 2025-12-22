package bootstrap

import (
	"database/sql"
	"net/http"

	"github.com/iPatrushevSergey/metrics/internal/filestorage"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

// App represents the application with resources needed for shutdown
type App struct {
	Server        *http.Server
	DB            *sql.DB
	Repository    repository.MetricRepository
	FileStorage   *filestorage.FileStorage
	PeriodicSaver *filestorage.PeriodicSaver
}
