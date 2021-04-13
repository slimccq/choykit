// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"fmt"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createTestDBConn() *mongo.Client {
	username := "admin"
	password := "cuKpVrfZzUvg"
	uri := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:27017/?connect=direct", username, password)

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Panicf("%v", err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Panicf("%v", err)
	}
	return client
}

func createMongoStore(label string) Storage {
	db := "testdb"
	cli := createTestDBConn()
	return NewMongoDBStore(cli, context.TODO(), db, label, DefaultSeqStep)
}

func TestMongoStore_Incr(t *testing.T) {

}
