package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func NewLogger(coll *mongo.Collection) *Logger {
	return &Logger{
		coll: coll,
	}
}

type Logger struct {
	coll *mongo.Collection
}

func (l *Logger) SendLog(log map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	print("-------------------")
	_, err := l.coll.InsertOne(ctx, log)
	print(err)
	return err
}
