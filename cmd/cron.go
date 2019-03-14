package cmd

func runSync() error {
	GenerateSyncNeededListFromMeta()
	return nil
}
