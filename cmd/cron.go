package cmd

func runSync() error {
	GenerateSyncNeededListFromMeta()
	return nil
}

func runHeartbeat() error {
	Heartbeat()
	return nil
}
