package cmd

var chanKV = make(chan map[string][]byte, 10)

func PutChanKV(k string, v []byte) error {
	m := make(map[string][]byte, 1)
	m[k] = v
	chanKV <- m
	return nil
}
