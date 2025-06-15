package cmd

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func mongoAggCountByKey(uname string, key string, mincount, maxcount int) (files []string) {
	t1 := GetNowUnixMillo()
	collUser := mgodb.Collection(uname)

	matchStage := bson.D{
		{"$match", bson.D{
			{key, bson.D{{"$gte", mincount}, {"$lt", maxcount}}},
		}}}

	sortStage := bson.D{
		{"$sort", bson.D{
			{key, -1},
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
	ShowRunningTime(t1, "mongoAggCountByKey")
	return files
}

func mongoAggFilesByKey(uname string, key string, val string) (files []string) {
	t1 := GetNowUnixMillo()
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
				{key, -1},
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

	ShowRunningTime(t1, "mongoAggFilesByKey")
	return files
}

func mongoAggSumByKey(uname string, key string) (totalSum int64) {
	if uname == "" || key == "" {
		return 0
	}

	cacheKey := bcacheKeyJoin(uname, "mongoAggSumByKey", key)
	bcval := bcacheGet(cacheKey)
	bckv := make(map[string]int64, 1)
	if bcval != nil {
		err := jsonDec(bcval, &bckv)
		PrintError("mongoAggSumByKey.10", err)
		totalSum = bckv["user_size_sum"]
		DebugInfo("mongoAggSumByKey:cache:bckv", bckv)
		return totalSum
	}

	t1 := GetNowUnixMillo()
	collUser := mgodb.Collection(uname)

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":           nil,
				"uniqueAmounts": bson.M{"$addToSet": "$size"},
			},
		},

		{
			"$unwind": "$uniqueAmounts",
		},

		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$uniqueAmounts"},
			},
		},
	}

	cursor, err := collUser.Aggregate(context.TODO(), pipeline)
	PrintError("mongoAggSumByKey.20", err)

	var result struct {
		Total int64 `bson:"total"`
	}
	if cursor.Next(context.TODO()) {
		if err := cursor.Decode(&result); err != nil {
			PrintError("mongoAggSumByKey.30", err)
		}
		totalSum = result.Total
	}
	ShowRunningTime(t1, "mongoAggSumByKey")
	DebugInfo("mongoAggSumByKey: totalSum", totalSum)
	DebugInfo("mongoAggSumByKey: totalSum", totalSum/Int2Int64(MB), " MB")

	bckv["user_size_sum"] = totalSum
	bcacheSet(cacheKey, jsonEnc(bckv))
	return totalSum
}
