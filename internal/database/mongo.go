// Package database owns the MongoDB connection lifecycle.
package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Open establishes one reusable client and verifies connectivity before startup.
// Errors intentionally omit driver details, which can contain connection secrets.
func Open(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(3 * time.Second))
	if err != nil {
		return nil, errors.New("invalid MongoDB connection configuration")
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = client.Disconnect(cleanup)
		return nil, errors.New("MongoDB is unavailable; check the database service and connection configuration")
	}
	return client, nil
}
