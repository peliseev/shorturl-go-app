package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShortURL struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	User      string             `bson:"user"`
	ShortURL  string             `bson:"short_url"`
	OriginURL string             `bson:"origin_url"`
}

type ShortURLService interface {
	CreateShortURL(context.Context, *ShortURL) (string, error)

	GetOriginUrl(context.Context, string) (*ShortURL, error)
}
