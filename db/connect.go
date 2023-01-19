package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToCollection(mongoURI string, database string, collection string) (*mongo.Collection, error) {
	conCtx, conCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer conCancel()
	cl, err := mongo.Connect(conCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer pingCancel()
	err = cl.Ping(pingCtx, nil)
	if err != nil {
		return nil, err
	}
	db := cl.Database(database)
	return db.Collection(collection), nil
}
