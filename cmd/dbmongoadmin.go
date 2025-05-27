package cmd

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

//use zstdfs
//db.createUser({user: "zstdfs",pwd: "zstdfs",roles: [ { role: "dbOwner", db: "zstdfs" } ]})
//show users

func mongoConnect() *mongo.Database {
	conn := GetEnv("zstdfs_mongo", "mongodb://localhost:27017/zstdfs")
	db, err := mongo.Connect(options.Client().ApplyURI(conn))
	FatalError("ERROR: mongoConnect. Please set the env var: zstdfs_mongo, it should be: mongodb://username:password@your-mongodb-ip:27017/zstdfs", err)
	zdb := db.Database("zstdfs")
	return zdb
}

func NewObjectIDFromTimestamp(ts time.Time) bson.ObjectID {
	oid := bson.NewObjectIDFromTimestamp(ts)
	return oid
}

func mongoAdminSetIfEmpty(id, k, v string) bool {
	collAdmin := mgodb.Collection(AdminBucket)

	var result bson.M
	filter := bson.D{{"_id", id}}
	err := collAdmin.FindOne(context.TODO(), filter).Decode(&result)
	PrintError("mongoAdminSetIfEmpty", err)

	if result[k] == nil {
		DebugInfo("mongoAdminSetIfEmpty:init set", result[k])
		update := bson.D{{"$set", bson.D{{k, v}}}}
		opt := options.UpdateOne().SetUpsert(true)
		collAdmin.UpdateOne(context.TODO(), filter, update, opt)
	}
	//
	return true
}

func mongoAdminUpsert(id, k, v string) bool {
	collAdmin := mgodb.Collection(AdminBucket)

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{k, v}}}}
	opt := options.UpdateOne().SetUpsert(true)

	collAdmin.UpdateOne(context.TODO(), filter, update, opt)
	return true
}

func mongoAdminListCollections() (lines []string) {
	filters := bson.M{}

	colls, err := mgodb.ListCollectionNames(context.TODO(), filters, nil)
	DebugInfo("mongoAdminListCollections", err, colls)

	for _, col := range colls {
		lines = append(lines, col)
	}
	DebugInfo("mongoAdminListCollections", lines)
	return lines
}

func mongoAdminCreateIndex(user string) bool {
	collUser := mgodb.Collection(user)

	indexes := make(map[string]int)
	indexes["mtime"] = -1
	indexes["size"] = 1
	indexes["dot_color"] = 1
	indexes["is_ban"] = 1
	indexes["is_public"] = -1
	indexes["stats_digg_count"] = -1
	indexes["stats_collect_count"] = -1
	indexes["stats_share_count"] = -1
	indexes["stats_comment_count"] = -1
	indexes["stats_download_count"] = -1
	indexes["tags"] = -1
	indexes["caption"] = -1

	for key, val := range indexes {
		indexModel := mongo.IndexModel{
			Keys: bson.D{
				{key, val},
			}}
		_, err := collUser.Indexes().CreateOne(context.TODO(), indexModel)
		FatalError("mongoAdminInitIndex:"+key, err)
	}

	return true
}
