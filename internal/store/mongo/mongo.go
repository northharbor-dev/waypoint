package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/northharbor-dev/waypoint/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var (
	ErrAlreadyClaimed = errors.New("work item already claimed or not in not_started state")
	ErrNotFound       = errors.New("work item not found")
)

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
}

func New(ctx context.Context, uri, database string) (*MongoStore, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return &MongoStore{
		client: client,
		db:     client.Database(database),
	}, nil
}

func (s *MongoStore) workItems() *mongo.Collection {
	return s.db.Collection("work_items")
}

func (s *MongoStore) phases() *mongo.Collection {
	return s.db.Collection("phases")
}

func (s *MongoStore) ListWorkItems(ctx context.Context, project string) ([]models.WorkItem, error) {
	cur, err := s.workItems().Find(ctx, bson.M{"project": project})
	if err != nil {
		return nil, err
	}
	var items []models.WorkItem
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *MongoStore) GetWorkItem(ctx context.Context, project, id string) (*models.WorkItem, error) {
	var item models.WorkItem
	err := s.workItems().FindOne(ctx, bson.M{"_id": id, "project": project}).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *MongoStore) ClaimWorkItem(ctx context.Context, project, id, claimedBy string) (*models.WorkItem, error) {
	now := time.Now()
	var item models.WorkItem
	err := s.workItems().FindOneAndUpdate(ctx,
		bson.M{"_id": id, "project": project, "status": models.StatusNotStarted},
		bson.M{"$set": bson.M{
			"status":     models.StatusInProgress,
			"claimed_by": claimedBy,
			"started_at": now,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrAlreadyClaimed
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *MongoStore) ReleaseWorkItem(ctx context.Context, project, id string) error {
	res, err := s.workItems().UpdateOne(ctx,
		bson.M{"_id": id, "project": project},
		bson.M{
			"$set": bson.M{
				"status":     models.StatusNotStarted,
				"updated_at": time.Now(),
			},
			"$unset": bson.M{
				"claimed_by":       "",
				"started_at":       "",
				"completed_at":     "",
				"duration_seconds": "",
			},
		},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) UpdateStatus(ctx context.Context, project, id string, status models.Status, note string) error {
	set := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}
	if note != "" {
		set["blocker_note"] = note
	}
	res, err := s.workItems().UpdateOne(ctx,
		bson.M{"_id": id, "project": project},
		bson.M{"$set": set},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) CompleteWorkItem(ctx context.Context, project, id string) error {
	item, err := s.GetWorkItem(ctx, project, id)
	if err != nil {
		return err
	}
	now := time.Now()
	update := bson.M{
		"status":       models.StatusDone,
		"completed_at": now,
		"updated_at":   now,
	}
	if item.StartedAt != nil {
		dur := int64(now.Sub(*item.StartedAt).Seconds())
		update["duration_seconds"] = dur
	}
	_, err = s.workItems().UpdateOne(ctx,
		bson.M{"_id": id, "project": project},
		bson.M{
			"$set":   update,
			"$unset": bson.M{"claimed_by": ""},
		},
	)
	return err
}

func (s *MongoStore) SeedProject(ctx context.Context, project string, items []models.WorkItem, phases []models.Phase) error {
	if _, err := s.workItems().DeleteMany(ctx, bson.M{"project": project}); err != nil {
		return err
	}
	if _, err := s.phases().DeleteMany(ctx, bson.M{"project": project}); err != nil {
		return err
	}

	if len(items) > 0 {
		now := time.Now()
		docs := make([]any, len(items))
		for i := range items {
			items[i].Status = models.StatusNotStarted
			items[i].UpdatedAt = now
			items[i].Project = project
			docs[i] = items[i]
		}
		if _, err := s.workItems().InsertMany(ctx, docs); err != nil {
			return err
		}
	}

	if len(phases) > 0 {
		docs := make([]any, len(phases))
		for i := range phases {
			phases[i].Project = project
			docs[i] = phases[i]
		}
		if _, err := s.phases().InsertMany(ctx, docs); err != nil {
			return err
		}
	}

	return nil
}

func (s *MongoStore) ListPhases(ctx context.Context, project string) ([]models.Phase, error) {
	cur, err := s.phases().Find(ctx, bson.M{"project": project})
	if err != nil {
		return nil, err
	}
	var result []models.Phase
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MongoStore) UpsertWorkItem(ctx context.Context, item models.WorkItem) error {
	_, err := s.workItems().UpdateOne(ctx,
		bson.M{"_id": item.ID},
		bson.M{
			"$set": bson.M{
				"title":        item.Title,
				"phase":        item.Phase,
				"owner":        item.Owner,
				"role":         item.Role,
				"dependencies": item.Dependencies,
				"project":      item.Project,
				"updated_at":   time.Now(),
			},
			"$setOnInsert": bson.M{
				"status": models.StatusNotStarted,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) DeleteWorkItem(ctx context.Context, project, id string) error {
	res, err := s.workItems().DeleteOne(ctx, bson.M{"_id": id, "project": project})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) UpsertPhase(ctx context.Context, phase models.Phase) error {
	_, err := s.phases().ReplaceOne(ctx,
		bson.M{"_id": phase.ID},
		phase,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
