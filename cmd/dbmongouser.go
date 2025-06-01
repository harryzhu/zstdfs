package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func mongoGuessValueTypeByKey(key string) (t string) {
	t = "string"
	k4Int64v := []string{"size", "mtime"}

	if Contains(k4Int64v, key) {
		t = "int64"
		return t
	}

	if strings.HasPrefix(key, "stats_") || strings.HasPrefix(key, "num_") {
		t = "int"
		return t
	}

	if strings.HasPrefix(key, "is_") || strings.HasSuffix(key, "_count") {
		t = "int"
		return t
	}

	if strings.HasPrefix(key, "int_") {
		t = "int"
		return t
	}

	if strings.HasPrefix(key, "int64_") {
		t = "int64"
		return t
	}

	if strings.HasPrefix(key, "f64_") {
		t = "float64"
		return t
	}

	if strings.HasPrefix(key, "strs_") {
		t = "strs"
		return t
	}

	if strings.HasPrefix(key, "ints_") {
		t = "ints"
		return t
	}

	return t
}

func mongoSave(user string, id string, k, v string) bool {
	if IsAnyEmpty(user, id, k) {
		return false
	}
	valType := mongoGuessValueTypeByKey(k)

	collUser := mgodb.Collection(user)

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{k, v}}}}
	if valType == "int" {
		update = bson.D{{"$set", bson.D{{k, Str2Int(v)}}}}
	}
	if valType == "int64" {
		update = bson.D{{"$set", bson.D{{k, Int2Int64(Str2Int(v))}}}}
	}
	if valType == "float64" {
		update = bson.D{{"$set", bson.D{{k, Str2Float64(v)}}}}
	}
	if valType == "strs" {
		update = bson.D{{"$set", bson.D{{k, Str2Strings(v, ",")}}}}
	}
	if valType == "ints" {
		update = bson.D{{"$set", bson.D{{k, Str2Ints(v, ",")}}}}
	}

	opt := options.UpdateOne().SetUpsert(true)
	_, err := collUser.UpdateOne(context.TODO(), filter, update, opt)

	if err != nil {
		PrintError("mongoSave", err)
		return false
	}
	return true
}

func mongoGet(user string, id string) (meta map[string]string) {
	var result bson.M
	meta = make(map[string]string)

	if IsAnyEmpty(user, id) {
		return meta
	}

	collUser := mgodb.Collection(user)
	filter := bson.D{{"_id", id}}

	err := collUser.FindOne(context.TODO(), filter).Decode(&result)
	PrintError("mongoGet", err)
	if err == nil {
		for k, v := range result {
			meta[k] = fmt.Sprintf("%v", v)
		}
	}

	return meta
}

func mongoGetByKey(user string, key string, val string) (meta map[string]string) {
	var result bson.M
	meta = make(map[string]string)

	if IsAnyEmpty(user, key, val) {
		return meta
	}

	collUser := mgodb.Collection(user)
	filter := bson.D{{key, val}}

	err := collUser.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		PrintError("mongoGetByKey", err)
		return meta
	}
	for k, v := range result {
		meta[k] = fmt.Sprintf("%v", v)
	}

	return meta
}

func mongoRandomGetURI(user string, total int) (files []string) {
	var allFiles []string

	allFilesCacheFile := fmt.Sprintf("%s/mongoRandomGet/files_all.dat", user)

	if GobLoad(allFilesCacheFile, &allFiles, FunctionCacheExpires) == false {
		var rows []bson.M
		filter := bson.D{{"_id", bson.Regex{Pattern: "(.mp4)$", Options: "i"}}}
		collUser := mgodb.Collection(user)
		results, err := collUser.Find(context.TODO(), filter)
		if err != nil {
			PrintError("mongoRandomGet", err)
			return files
		}

		err = results.All(context.TODO(), &rows)
		PrintError("mongoListFiles", err)

		var rowuri string
		for _, row := range rows {
			//rowid = row["_id"].(string)
			rowuri = row["uri"].(string)
			allFiles = append(allFiles, rowuri)
		}

		GobDump(allFilesCacheFile, allFiles)
	}

	maxIndex := len(allFiles)
	indexes := GetRandomInts(total, 10, maxIndex)
	DebugInfo("mongoRandomGet----max", maxIndex, ":", indexes)
	for _, idx := range indexes {
		files = append(files, allFiles[idx])
	}

	return files
}

func mongoListFiles(user, prefix string, optSort bson.D) (dirs, files []string) {
	var rows []bson.M
	var lines []string
	if user == "" {
		return lines, lines
	}

	filter := bson.D{{}}
	if prefix != "" {
		DebugInfo("prefix before regex", prefix)
		regxprefix := strings.Join([]string{"^(", prefix, ")"}, "")
		filter = bson.D{{"_id", bson.Regex{Pattern: regxprefix, Options: "i"}}}
	}
	DebugInfo("mongoListFiles:filter", filter)

	collUser := mgodb.Collection(user)

	results, err := collUser.Find(context.TODO(), filter, options.Find().SetSort(optSort))
	if err != nil {
		PrintError("mongoListFiles", err)
		return lines, lines
	}

	err = results.All(context.TODO(), &rows)
	PrintError("mongoListFiles", err)

	dds := make(map[string]int, 100)
	var rowid string
	for _, row := range rows {
		rowid = row["_id"].(string)
		rowid = strings.TrimPrefix(rowid, prefix)
		rowid = strings.Trim(rowid, "/")
		if strings.Contains(rowid, "/") {
			dd := strings.Split(rowid, "/")
			dds[dd[0]] = 1
		} else {
			files = append(files, rowid)
		}
	}
	//DebugInfo("mongoListFiles:dirs", dirs)
	//DebugInfo("mongoListFiles:files", files)
	for k, _ := range dds {
		if k != "" {
			dirs = append(dirs, k)
		}
	}

	return dirs, files
}

func mongoTagFiles(user, tagname string) (files []string) {
	var rows []bson.M
	var lines []string
	if user == "" || tagname == "" {
		return lines
	}
	filesCacheFile := fmt.Sprintf("%s/mongoTagFiles/files_%s.dat", user, GetXxhash([]byte(tagname)))

	if GobLoad(filesCacheFile, &files, FunctionCacheExpires) == false {
		//tagname = strings.ReplaceAll(tagname, "&", ".*")
		regxTagname := strings.Join([]string{"(", tagname, ")"}, "")
		filter := bson.D{
			{"tags", bson.D{{"$exists", true}}},
			{"tags", bson.D{{"$ne", ""}}},
			{"tags", bson.Regex{Pattern: regxTagname, Options: "i"}},
		}

		optSort := bson.D{{"_id", 1}, {"mtime", -1}}

		collUser := mgodb.Collection(user)
		results, err := collUser.Find(context.TODO(), filter, options.Find().SetSort(optSort))
		if err != nil {
			PrintError("mongoTagFiles.10", err)
			return lines
		}

		err = results.All(context.TODO(), &rows)
		PrintError("mongoTagFiles.20", err)

		var rowid string
		for _, row := range rows {
			rowid = row["_id"].(string)
			files = append(files, rowid)
		}

		GobDump(filesCacheFile, files)
	}

	return files
}

func mongoTagList(user string, idPrefix string) (tags []string) {
	var lines []string
	if user == "" {
		return lines
	}
	tagsCacheFile := fmt.Sprintf("%s/mongoTagList/tags.dat", user)
	tagCountCacheFile := fmt.Sprintf("%s/mongoTagList/tags_count.dat", user)

	if len(idPrefix) > 0 {
		tagsCacheFile = fmt.Sprintf("%s/mongoTagList/tags_%s.dat", user, GetXxhash([]byte(idPrefix)))
		tagCountCacheFile = fmt.Sprintf("%s/mongoTagList/tags_count_%s.dat", user, GetXxhash([]byte(idPrefix)))

	}

	if GobLoad(tagsCacheFile, &tags, FunctionCacheExpires) == false {
		filter := bson.D{
			{"tags", bson.D{{"$exists", true}}},
			{"tags", bson.D{{"$ne", ""}}},
		}

		if len(idPrefix) > 0 {
			regxIDPrefix := strings.Join([]string{"(", idPrefix, ")"}, "")
			filter = bson.D{
				{"$and",
					bson.A{
						bson.D{{"tags", bson.D{{"$exists", true}}}},
						bson.D{{"tags", bson.D{{"$ne", ""}}}},
						bson.D{{"_id", bson.Regex{Pattern: regxIDPrefix, Options: "i"}}},
					},
				},
			}
		}

		opts := options.Find().SetProjection(bson.D{{"tags", 1}})

		collUser := mgodb.Collection(user)
		//err := collUser.Distinct(context.TODO(), "tags", filter).Decode(&lines)
		results, err := collUser.Find(context.TODO(), filter, opts)
		if err != nil {
			PrintError("mongoTagList.10", err)
			return lines
		}

		var rows []bson.M
		err = results.All(context.TODO(), &rows)
		PrintError("mongoTagList.20", err)

		tagcount := make(map[string]int)
		var line string
		for _, row := range rows {
			line = row["tags"].(string)
			line = TagLineFormat(line)
			words := strings.Split(line, ",")
			for _, word := range words {
				word = strings.TrimSpace(word)
				if word != "" {
					tagcount[word] += 1
				}
			}
		}
		for k, _ := range tagcount {
			tags = append(tags, k)
		}

		GobDump(tagsCacheFile, tags)
		GobDump(tagCountCacheFile, tagcount)
	}

	return tags
}

func mongoTagCount(user, tagname string) (tagcount map[string]int) {
	if IsAnyEmpty(user, tagname) {
		return tagcount
	}

	tagCountCacheFile := fmt.Sprintf("%s/mongoTagList/tags_count.dat", user)
	if GobLoad(tagCountCacheFile, &tagcount, FunctionCacheExpires) == false {
		mongoTagList(user, "")
	}
	GobLoad(tagCountCacheFile, &tagcount, FunctionCacheExpires)

	if tagcount == nil {
		tagcount[tagname] = len(mongoTagFiles(user, tagname))
	}
	return tagcount
}

func mongoCaptionList(user string) (tags []string) {
	var lines []string
	if user == "" {
		return lines
	}
	tagsCacheFile := fmt.Sprintf("%s/mongoCaptionList/caption.dat", user)
	tagCountCacheFile := fmt.Sprintf("%s/mongoCaptionList/caption_count.dat", user)

	if GobLoad(tagsCacheFile, &tags, FunctionCacheExpires) == false {
		filter := bson.D{
			{"caption", bson.D{{"$exists", true}}},
			{"caption", bson.D{{"$ne", ""}}},
		}
		opts := options.Find().SetProjection(bson.D{{"caption", 1}})

		collUser := mgodb.Collection(user)
		results, err := collUser.Find(context.TODO(), filter, opts)
		if err != nil {
			PrintError("mongoCaptionList.10", err)
			return lines
		}

		var rows []bson.M
		err = results.All(context.TODO(), &rows)
		PrintError("mongoCaptionList.20", err)

		tagcount := make(map[string]int)
		var line string
		for _, row := range rows {
			line = row["caption"].(string)
			line = strings.ReplaceAll(line, ".", ",")
			line = strings.ReplaceAll(line, ";", ",")
			words := strings.Split(line, ",")
			for _, word := range words {
				word = strings.TrimSpace(word)
				if len(word) < 3 || strings.Count(word, " ") > 10 {
					continue
				}

				tagcount[word] += 1

			}
		}

		for k, _ := range tagcount {
			tags = append(tags, k)
			//DebugInfo("mongoCaptionList.30", k, ":", v)
		}

		GobDump(tagsCacheFile, tags)
		GobDump(tagCountCacheFile, tagcount)
	}

	return tags
}

func mongoCaptionCount(user, tagname string) (tagcount map[string]int) {
	if IsAnyEmpty(user, tagname) {
		return tagcount
	}

	tagCountCacheFile := fmt.Sprintf("%s/mongoCaptionList/caption_count.dat", user)
	if GobLoad(tagCountCacheFile, &tagcount, FunctionCacheExpires) == false {
		mongoCaptionList(user)
	}
	GobLoad(tagCountCacheFile, &tagcount, FunctionCacheExpires)

	if tagcount == nil {
		tagcount[tagname] = len(mongoCaptionFiles(user, tagname))
	}

	return tagcount
}

func mongoCaptionFiles(user, captionWord string) (files []string) {
	var rows []bson.M
	var lines []string
	if user == "" || captionWord == "" {
		return lines
	}
	filesCacheFile := fmt.Sprintf("%s/mongoCaptionFiles/files_%s.dat", user, GetXxhash([]byte(captionWord)))

	if GobLoad(filesCacheFile, &files, FunctionCacheExpires) == false {
		captionWord = strings.ReplaceAll(captionWord, "&", ".*")
		regxCaptionWord := strings.Join([]string{"(", captionWord, ")"}, "")

		filter := bson.D{
			{"caption", bson.D{{"$exists", true}}},
			{"caption", bson.D{{"$ne", ""}}},
			{"caption", bson.Regex{Pattern: regxCaptionWord, Options: "i"}},
		}

		if strings.Index(captionWord, "/") > 0 && strings.Index(captionWord, "/") < len(captionWord) {
			idPrefixCaption := strings.Split(captionWord, "/")
			if len(idPrefixCaption) == 2 {
				regxCaptionWord = strings.Join([]string{"(", idPrefixCaption[1], ")"}, "")
				regxIDPrefix := strings.Join([]string{"(", idPrefixCaption[0], ")"}, "")
				filter = bson.D{
					{"_id", bson.Regex{Pattern: regxIDPrefix, Options: "i"}},
					{"caption", bson.D{{"$exists", true}}},
					{"caption", bson.D{{"$ne", ""}}},
					{"caption", bson.Regex{Pattern: regxCaptionWord, Options: "i"}},
				}
			}
		}

		optSort := bson.D{{"_id", 1}, {"mtime", -1}}

		opts := options.Find().SetSort(optSort)
		opts.SetProjection(bson.D{{"_id", 1}})

		collUser := mgodb.Collection(user)
		results, err := collUser.Find(context.TODO(), filter, opts)
		if err != nil {
			PrintError("mongoCaptionFiles.10", err)
			return lines
		}

		err = results.All(context.TODO(), &rows)
		PrintError("mongoCaptionFiles.20", err)

		var rowid string
		for _, row := range rows {
			rowid = row["_id"].(string)
			files = append(files, rowid)
		}
		GobDump(filesCacheFile, files)
	}

	return files
}

func mongoUserStats(user string) (stats map[string]string) {
	if user == "" {
		return stats
	}
	stats = make(map[string]string)
	collUser := mgodb.Collection(user)
	docCount, err := collUser.EstimatedDocumentCount(context.TODO(), nil)
	PrintError("mongoUserStats:docCount", err)

	stats["doc_count"] = Int64ToString(docCount)
	stats["unique_doc_count"] = Int2Str(len(mongoDistinctByKey(user, "_fsum")))

	DebugInfo("mongoUserStats", stats)
	return stats
}

func mongoDistinctByKey(uname string, key string) (files []string) {
	filter := bson.D{{
		key, bson.D{{"$exists", true}, {"$ne", ""}},
	}}
	collUser := mgodb.Collection(uname)

	err := collUser.Distinct(context.TODO(), key, filter).Decode(&files)
	PrintError("mongoDistinctByKey", err)

	return files
}

func mongoUserCollectionInit(user string) bool {
	testFile := StaticDir + "/test.jpg"

	finfo, err := os.Stat(testFile)
	FatalError("EntitySaveSmoke", err)
	mtime := finfo.ModTime().UTC().Unix()

	entity := NewEntity(user, testKey).WithFile(testFile)

	meta := make(map[string]string)
	meta["author"] = user
	meta["size"] = Int64ToString(finfo.Size())
	meta["mtime"] = Int64ToString(mtime)
	meta["fsha256"] = SHA256File(testFile)
	//
	meta["tags"] = TagLineFormat("壁纸，自然;杭州；游湖#泛舟/休闲")
	meta["strs_tags"] = meta["tags"]
	meta["ints_days"] = "3,  7  , 21"
	meta["dot_color"] = "dot-purple"
	meta["mime"] = "image/jpeg"
	meta["caption"] = "file, caption, sample"
	//
	meta["stats_digg_count"] = "100"
	meta["stats_comment_count"] = "200"
	meta["stats_collect_count"] = "300"
	meta["stats_share_count"] = "400"
	meta["stats_download_count"] = "500"
	meta["is_public"] = "1"
	meta["is_ban"] = "0"
	meta["num_prefix"] = "1000"
	meta["stats_prefix"] = "2000"

	for k, v := range meta {
		entity.WithMeta(k, v)
	}

	entity.Save()

	return true
}

func mongoBatchWriteFiles(files []string) bool {
	metaAllowKey := []string{"_id", "size", "mime", "mtime", "_fsum"}
	BulkLoadDir = ToUnixSlash(BulkLoadDir)
	var err error
	var ett Entity
	var fp *os.File
	var val []byte
	for _, file := range files {
		file = ToUnixSlash(file)
		fp, err = os.Open(file)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		val, err = io.ReadAll(fp)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		fp.Close()

		ID := strings.TrimPrefix(ToUnixSlash(strings.TrimPrefix(file, BulkLoadDir)), "/")

		ett = NewEntity(BulkLoadUser, ID).WithFile(file)
		for k := range ett.Meta {
			if !Contains(metaAllowKey, k) {
				delete(ett.Meta, k)
			}
		}
		//DebugInfo("mongoBatchWriteFiles", file, " <= ", ID)
		ett.Meta["_fsum"] = string(SumBlake3(val))
		ett.SaveWithoutData()
		//DebugInfo("mongoBatchWriteFiles", ID, "<= ", file)
	}
	DebugInfo("mongoBatchWriteFiles: files: ", len(files))
	return true
}
