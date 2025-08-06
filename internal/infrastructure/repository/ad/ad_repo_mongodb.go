package ad

import (
	"context"
	"errors"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	ErrorAdsNotFound    = errors.New("ads not found")
	ErrorFailedToSaveAd = errors.New("failed to save ad")
)

const (
	collectionName = "ads"
)

type AdRepoMongoDB struct {
	collection *mongo.Collection
	client     *mongo.Client
}

func NewAdRepoMongoDB(db *mongo.Database) *AdRepoMongoDB {
	return &AdRepoMongoDB{
		collection: db.Collection(collectionName),
		client:     db.Client(),
	}
}

func (r *AdRepoMongoDB) Save(ad *entity.Ad) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, ad)
	if err != nil {
		return ErrorFailedToSaveAd
	}

	return nil
}

func (r *AdRepoMongoDB) FindAll(ops *entity.Options) ([]*entity.Ad, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if ops.MinPrice > 0 || ops.MaxPrice > 0 {
		priceFilter := bson.M{}
		if ops.MinPrice > 0 {
			priceFilter["$gte"] = ops.MinPrice
		}
		if ops.MaxPrice > 0 {
			priceFilter["$lte"] = ops.MaxPrice
		}
		filter[entity.SortByPrice] = priceFilter
	}

	orderBy := entity.OrderByDesc
	if ops.OrderBy == entity.OrderByAsc {
		orderBy = entity.OrderByAsc
	}
	sortBy := entity.SortByCreatedAt
	if ops.SortBy == entity.SortByPrice {
		sortBy = entity.SortByPrice
	}
	sortOps := bson.D{{Key: sortBy, Value: orderBy}}

	findOps := options.Find().
		SetSort(sortOps).
		SetSkip(int64((ops.Page - 1) * ops.Limit)).
		SetLimit(int64(ops.Limit))

	cursor, err := r.collection.Find(ctx, filter, findOps)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ads []*entity.Ad
	if err = cursor.All(ctx, &ads); err != nil {
		return nil, err
	}

	if len(ads) == 0 {
		return nil, ErrorAdsNotFound
	}

	return ads, nil
}
