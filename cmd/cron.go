package cmd

import (
	pbm "hazhufs/pb/master_pb"
	"strings"

	"google.golang.org/grpc"
)

func runSync() error {
	GenerateSyncNeededListFromMeta()
	return nil
}

func runHeartbeat() error {
	Heartbeat()
	return nil
}

func run_Job_VolumeMasterStreamSendFile() error {
	addr := strings.Join([]string{CFG.Master.Ip, CFG.Master.Port}, ":")
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
	defer conn.Close()
	if err != nil {
		Logger.Error(err)
		return err
	}
	c := pbm.NewMasterServiceClient(conn)
	Job_VolumeMasterStreamSendFile(c)
	return nil
}
