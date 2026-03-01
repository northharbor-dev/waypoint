package store

import (
	"context"

	"github.com/northharbor-dev/waypoint/internal/models"
)

type Store interface {
	ListWorkItems(ctx context.Context, project string) ([]models.WorkItem, error)
	GetWorkItem(ctx context.Context, project string, id string) (*models.WorkItem, error)
	ClaimWorkItem(ctx context.Context, project string, id string, claimedBy string) (*models.WorkItem, error)
	ReleaseWorkItem(ctx context.Context, project string, id string) error
	UpdateStatus(ctx context.Context, project string, id string, status models.Status, note string) error
	CompleteWorkItem(ctx context.Context, project string, id string) error
	SeedProject(ctx context.Context, project string, items []models.WorkItem, phases []models.Phase) error
	ListPhases(ctx context.Context, project string) ([]models.Phase, error)

	UpsertWorkItem(ctx context.Context, item models.WorkItem) error
	DeleteWorkItem(ctx context.Context, project string, id string) error
	UpsertPhase(ctx context.Context, phase models.Phase) error

	Close(ctx context.Context) error
}
