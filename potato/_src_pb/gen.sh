#!/bin/bash
protoc --go_out=plugins=grpc:../pb/volume volume.proto 
#protoc --go_out=plugins=grpc:../../../potatoctl/client/pb/volume volume.proto