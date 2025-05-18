package db

import (
	"context"
	"sync"
	"time"

	conf "github.com/iceymoss/go-hichat-api/pkg/config"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoConn = make(map[string]*mongo.Client)
var mongoMutex sync.RWMutex

func GetMongoConn() *mongo.Client {
	mongoMutex.RLock()
	conn, ok := mongoConn["main"]
	mongoMutex.RUnlock()
	if !ok {
		mongoMutex.Lock()
		mongoUri := conf.ServiceConf.Mongo.Link
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri).SetMaxPoolSize(120))
		if err != nil {
			zLog.Error(err.Error())
			return nil
		}

		mongoConn["main"] = client
		conn = client
		mongoMutex.Unlock()
	}

	return conn
}
