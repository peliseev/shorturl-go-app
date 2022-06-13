package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/peliseev/shorturl-go-app/domain"
)

type ShortURLService struct {
	db         *mongo.Client
	collection *mongo.Collection
}

func NewShortURLService(db *mongo.Client) *ShortURLService {
	return &ShortURLService{db, db.Database("shorturl").Collection("urls")}
}

func (sus *ShortURLService) CreateShortURL(ctx context.Context, shortURL *domain.ShortURL) (string, error) {
	err := sus.collection.FindOne(ctx, bson.D{{Key: "short_url", Value: shortURL.ShortURL}}).Decode(shortURL)
	if err == nil {
		return shortURL.ShortURL, nil
	}

	shortURL.CreatedAt = time.Now()
	shortURL.UpdatedAt = time.Now()

	_, err = sus.collection.InsertOne(ctx, shortURL)
	if err != nil {
		log.Print("Error while 'InsertOne' operation")
		return "", err
	}
	return shortURL.ShortURL, nil
}

func (sus *ShortURLService) GetOriginURL(ctx context.Context, shortURL string) (*domain.ShortURL, error) {
	var su domain.ShortURL
	filter := bson.M{
		"short_url": shortURL,
	}
	update := bson.M{
		"$inc": bson.M{"count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}
	err := sus.collection.FindOneAndUpdate(ctx, filter, update).Decode(&su)
	return &su, err
}

func (sus *ShortURLService) GetURLFollowCount(ctx context.Context, shortURL string) (*domain.ShortURL, error) {
	var su domain.ShortURL
	filter := bson.M{
		"short_url": shortURL,
	}
	err := sus.collection.FindOne(ctx, filter).Decode(&su)
	return &su, err
}
