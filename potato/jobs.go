package potato

func BdbCompaction() {
	if isBDBValueLogGCNeeded == true {
		isBDBValueLogGCNeeded = false
		bdb_compact()
		isBDBValueLogGCNeeded = true
	}
}
