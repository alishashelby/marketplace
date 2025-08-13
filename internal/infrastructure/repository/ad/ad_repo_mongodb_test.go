package ad

import (
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"testing"
)

func TestAdRepoMongoDB_Save(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		ad := &entity.Ad{
			ID:    uuid.New(),
			Title: "test",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := repo.Save(ad)

		assert.NoError(t, err)
	})

	mt.Run("Failure", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		ad := &entity.Ad{
			ID:    uuid.New(),
			Title: "test",
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{}))
		err := repo.Save(ad)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrorFailedToSaveAd)
	})
}

func TestAdRepoMongoDB_FindAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		opts := &entity.Options{
			Page:     1,
			Limit:    10,
			SortBy:   entity.SortByPrice,
			OrderBy:  entity.OrderByDesc,
			MinPrice: 100,
			MaxPrice: 200,
		}
		expected := []*entity.Ad{
			{
				ID:    uuid.New(),
				Price: 200,
			},
			{
				ID:    uuid.New(),
				Price: 150,
			},
		}

		first := mtest.CreateCursorResponse(1, "db.ads", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expected[0].ID},
			{Key: "price", Value: expected[0].Price},
		})
		second := mtest.CreateCursorResponse(1, "db.ads", mtest.NextBatch, bson.D{
			{Key: "_id", Value: expected[1].ID},
			{Key: "price", Value: expected[1].Price},
		})
		killCursors := mtest.CreateCursorResponse(0, "db.ads", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		ads, err := repo.FindAll(opts)

		assert.NoError(t, err)
		assert.Len(t, ads, 2)
		assert.Equal(t, expected[0], ads[0])
		assert.Equal(t, expected[1], ads[1])
	})

	mt.Run("Failure - not found", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		opts := &entity.Options{
			Page:  1,
			Limit: 10,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.ads", mtest.FirstBatch))
		ads, err := repo.FindAll(opts)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrorAdsNotFound)
		assert.Nil(t, ads)
	})

	mt.Run("Success - ascending order", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		opts := &entity.Options{
			Page:    1,
			Limit:   10,
			SortBy:  entity.SortByPrice,
			OrderBy: entity.OrderByAsc,
		}
		expected := []*entity.Ad{
			{
				ID:    uuid.New(),
				Price: 150,
			},
			{
				ID:    uuid.New(),
				Price: 200,
			},
		}

		first := mtest.CreateCursorResponse(1, "db.ads", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expected[0].ID},
			{Key: "price", Value: expected[0].Price},
		})
		second := mtest.CreateCursorResponse(0, "db.ads", mtest.NextBatch, bson.D{
			{Key: "_id", Value: expected[1].ID},
			{Key: "price", Value: expected[1].Price},
		})
		killCursors := mtest.CreateCursorResponse(0, "db.ads", mtest.NextBatch)

		mt.AddMockResponses(first, second, killCursors)
		ads, err := repo.FindAll(opts)

		assert.NoError(t, err)
		assert.Len(t, ads, 2)
		assert.Equal(t, expected[0], ads[0])
		assert.Equal(t, expected[1], ads[1])
	})

	mt.Run("Failure - error in find command", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		opts := &entity.Options{
			Page:  1,
			Limit: 10,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{}))

		ads, err := repo.FindAll(opts)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrorAdsNotFound)
		assert.Nil(t, ads)
	})

	mt.Run("Failure - cursor decoding error", func(mt *mtest.T) {
		repo := NewAdRepoMongoDB(mt.DB)
		opts := &entity.Options{
			Page:  1,
			Limit: 10,
		}

		invalid := bson.D{
			{Key: "_id", Value: "invalid"},
		}

		first := mtest.CreateCursorResponse(1, "db.ads", mtest.FirstBatch, invalid)
		killCursors := mtest.CreateCursorResponse(0, "db.ads", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		ads, err := repo.FindAll(opts)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding key _id")
		assert.Nil(t, ads)
	})
}
