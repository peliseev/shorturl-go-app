package domain

import (
	"context"
	"time"
)

// ShortURL is main domain entity, which represent
// map between shortURL and originURL
type ShortURL struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	User      string    `bson:"user"`
	ShortURL  string    `bson:"short_url"`
	OriginURL string    `bson:"origin_url"`
	Count     int       `bson:"count"`
}

// ShortURLService interface for interact within ShortURL entity.
type ShortURLService interface {
	CreateShortURL(context.Context, *ShortURL) (string, error)
	GetOriginURL(context.Context, string) (*ShortURL, error)
	GetURLFollowCount(context.Context, string) (*ShortURL, error)
}
