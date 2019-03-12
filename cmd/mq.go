package cmd

// import (
// 	"errors"
// 	"strings"

// 	"github.com/nsqio/go-nsq"
// )

// func PrepareMasterMessageQueue() error {
// 	mq_conf := strings.Trim(CFG.Global.Mq_address, " ")
// 	if strings.Index(mq_conf, ":") < 1 {
// 		Logger.Fatal("ERROR: in the config.toml, global>mq_address should be ip:port")
// 		return errors.New("in the config.toml, global>mq_address should be ip:port")
// 	} else {
// 		Logger.Info("mq_address: ", mq_conf)
// 	}

// 	NSQP, err := nsq.NewProducer(mq_conf, nsq.NewConfig())
// 	if err != nil {
// 		Logger.Fatal(err)
// 		return errors.New("cannot create message queue producer.")
// 	}
// 	NSQP.SetLogger(nil, 0)

// 	NSQP.Publish("hazhufs", []byte("hello")) //topic， message

// 	return nil
// }

// func MasterPublish(b string) error {
// 	NSQP, _ := nsq.NewProducer("172.16.32.49:4150", nsq.NewConfig())

// 	NSQP.Publish("hazhufs", []byte(b))
// 	return nil
// }
