package cmd

import (
	"context"
	"strings"

	//"strings"

	//"encoding/json"
	"fmt"
	//"io/ioutil"
	//"os"
	//"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	//"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func mongoAggCountByKey(uname string, key string, min_count, max_count int) (files []string) {
	fmt.Print("-----mongoAggCountByKey----")
	collUser := mgodb.Collection(uname)

	matchStage := bson.D{
		{"$match", bson.D{
			{key, bson.D{{"$gte", min_count}, {"$lt", max_count}}},
		}}}

	sortStage := bson.D{
		{"$sort", bson.D{
			{key, 1},
		}}}

	DebugInfo("mongoAggCountByKey:matchStage", matchStage)
	DebugInfo("mongoAggCountByKey:sortStage", sortStage)

	cursor, err := collUser.Aggregate(context.TODO(), mongo.Pipeline{matchStage, sortStage})
	PrintError("mongoAggCountByKey", err)

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		PrintError("mongoAggCountByKey", err)
	}
	for _, result := range results {
		files = append(files, fmt.Sprintf("%v", result["_id"]))
		//DebugInfo("mongoAggCountByKey", result)
	}
	return files
}

func mongoAggFilesByKey(uname string, key string, val string) (files []string) {
	fmt.Print("-----mongoAggFilesByKey----")

	kv := strings.Join([]string{key, val}, ":")
	filesCacheFile := fmt.Sprintf("%s/mongoAggFilesByKey/files_%s.dat", uname, GetXxhash([]byte(kv)))

	if GobLoad(filesCacheFile, &files, FunctionCacheExpires) == false {
		collUser := mgodb.Collection(uname)
		matchStage := bson.D{
			{"$match", bson.D{
				{key, bson.D{{"$eq", val}}},
			}}}

		valType := mongoGuessValueTypeByKey(key)
		if valType == "int" {
			matchStage = bson.D{
				{"$match", bson.D{
					{key, bson.D{{"$eq", Str2Int(val)}}},
				}}}
		}
		if valType == "int64" {
			matchStage = bson.D{
				{"$match", bson.D{
					{key, bson.D{{"$eq", Int2Int64(Str2Int(val))}}},
				}}}
		}
		if valType == "float64" {
			matchStage = bson.D{
				{"$match", bson.D{
					{key, bson.D{{"$eq", Str2Float64(val)}}},
				}}}
		}

		sortStage := bson.D{
			{"$sort", bson.D{
				{key, 1},
			}}}

		DebugInfo("mongoAggFilesByKey:matchStage", matchStage)
		DebugInfo("mongoAggFilesByKey:sortStage", sortStage)

		cursor, err := collUser.Aggregate(context.TODO(), mongo.Pipeline{matchStage, sortStage})
		PrintError("mongoAggFilesByKey", err)

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			PrintError("mongoAggFilesByKey", err)
		}
		for _, result := range results {
			files = append(files, fmt.Sprintf("%v", result["_id"]))
			//DebugInfo("mongoAggFilesByKey", result)
		}
		if len(files) > 0 {
			GobDump(filesCacheFile, files)
		}

	}

	return files
}
