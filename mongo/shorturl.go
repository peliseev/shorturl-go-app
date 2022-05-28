package mongo

import (
	"context"
	"log"

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
	err := sus.collection.FindOne(ctx, bson.D{{"short_url", shortURL.ShortURL}}).Decode(shortURL)
	if err == nil {
		return shortURL.ShortURL, nil
	}
	_, err = sus.collection.InsertOne(ctx, shortURL)
	if err != nil {
		log.Print("Error while 'InsertOne' operation")
		return "", err
	}
	return shortURL.ShortURL, nil
}

func (sus *ShortURLService) GetOriginUrl(ctx context.Context, shortURL string) (*domain.ShortURL, error) {
	return &domain.ShortURL{}, nil
}
