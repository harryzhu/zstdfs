package cmd

import (
	"context"
	"fmt"
	"strings"
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
	indexes["caption.en"] = -1
	indexes["caption.cn"] = -1
	//
	indexes["uri"] = 11
	//

	indexModel := mongo.IndexModel{}
	for key, val := range indexes {
		if val == 99 {
			indexModel = mongo.IndexModel{
				Keys: bson.D{
					{"caption", "text"},
				}}
		} else if val == 11 {
			indexModel = mongo.IndexModel{
				Keys: bson.D{
					{key, 1},
				},
				Options: options.Index().SetUnique(true),
			}
		} else {
			indexModel = mongo.IndexModel{
				Keys: bson.D{
					{key, val},
				}}
		}

		_, err := collUser.Indexes().CreateOne(context.TODO(), indexModel)
		FatalError("mongoAdminInitIndex:"+key, err)
	}

	return true
}

func mongoAdminResetKeyStats(uname string, key string) error {
	DebugInfo("mongoAdminResetKeyStats", uname, ":", key)
	keyName := ""
	if strings.HasPrefix(key, "_") {
		keyName = strings.Join([]string{"count", key}, "")
	} else {
		keyName = strings.Join([]string{"count", key}, "_")
	}

	collAdmin := mgodb.Collection(AdminBucket)

	regxprefix := strings.Join([]string{"^(", uname, "::", key, "::)"}, "")
	filter := bson.D{
		{"_id", bson.Regex{Pattern: regxprefix, Options: "i"}},
	}

	update := bson.D{{"$set", bson.D{{keyName, 0}}}}

	_, err := collAdmin.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		PrintError("mongoAdminResetKeyStats.10", err)
		return err
	}
	return nil
}

func mongoAdminUpdateKeyStats(uname string, key string) error {
	DebugInfo("mongoAdminUpdateKeyStats", uname, ":", key)
	collUser := mgodb.Collection(uname)
	filter := bson.D{{}}
	update := bson.D{{}}

	keyCount := make(map[string]int)

	results, err := collUser.Find(context.TODO(), filter, nil)
	if err != nil {
		PrintError("mongoAdminUpdateKeyStats.10", err)
		return err
	}

	var rows []bson.M
	err = results.All(context.TODO(), &rows)
	PrintError("mongoAdminUpdateKeyStats.20", err)

	statKey := ""
	for _, row := range rows {
		rowfsha256 := fmt.Sprint(row["fsha256"])
		if len(rowfsha256) > 16 {
			statKey = strings.Join([]string{uname, key, rowfsha256}, "::")
			keyCount[statKey] += 1
		}
	}

	keyName := ""
	if strings.HasPrefix(key, "_") {
		keyName = strings.Join([]string{"count", key}, "")
	} else {
		keyName = strings.Join([]string{"count", key}, "_")
	}
	optUpdate := options.UpdateOne().SetUpsert(true)
	collAdmin := mgodb.Collection(AdminBucket)
	for k, v := range keyCount {
		if v > 1 {
			filter = bson.D{{"_id", k}}
			update = bson.D{{"$set", bson.D{{keyName, v}}}}
			_, err := collAdmin.UpdateOne(context.TODO(), filter, update, optUpdate)
			if err != nil {
				PrintError("mongoUpdateKeyStats", err)
				return err
			}
		}
	}

	return nil
}

func mongoAdminGetKeyStats(uname string, key string) (urls []map[string]int) {
	keyName := ""
	if strings.HasPrefix(key, "_") {
		keyName = strings.Join([]string{"count", key}, "")
	} else {
		keyName = strings.Join([]string{"count", key}, "_")
	}

	collAdmin := mgodb.Collection(AdminBucket)
	regxprefix := strings.Join([]string{"^(", uname, "::", key, "::)"}, "")
	filter := bson.D{
		{"_id", bson.Regex{Pattern: regxprefix, Options: "i"}},
		{keyName, bson.D{{"$gt", 1}}},
	}

	opts := options.Find().SetSort(bson.D{{keyName, -1}})

	results, err := collAdmin.Find(context.TODO(), filter, opts)
	if err != nil {
		PrintError("mongoAdminGetKeyStats.10", err)
		return urls
	}

	var rows []bson.M
	err = results.All(context.TODO(), &rows)
	PrintError("mongoAdminGetKeyStats.20", err)

	unamekey := strings.Join([]string{uname, "::", key, "::"}, "")

	for _, row := range rows {
		if fmt.Sprint(row["_id"]) != "" && row["_id"] != nil {
			idcount := make(map[string]int)
			mkey := fmt.Sprint(row["_id"])[len(unamekey):]
			idcount[mkey] = Str2Int(fmt.Sprintf("%v", row[keyName]))
			urls = append(urls, idcount)
		}
	}
	return urls
}
