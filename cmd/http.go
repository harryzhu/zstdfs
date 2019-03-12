// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	// "bytes"
	// "encoding/gob"
	//"encoding/json"
	//"errors"
	// "fmt"
	pbv "hazhufs/pb/volume_pb"

	"net/http"
	"strconv"
	"strings"
	"time"

	//"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	r = mux.NewRouter()
)

// masterCmd represents the master command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "http",
	Long:  `Example: hazhufs http" `,
	Run: func(cmd *cobra.Command, args []string) {
		PrepareHttpCacheDatabase()

		srv := &http.Server{
			Handler: r,
			Addr:    strings.Join([]string{CFG.Http.Ip, CFG.Http.Port}, ":"),
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		}

		Logger.Fatal(srv.ListenAndServe())
	},
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	//w.Header().Set("Content-Length", _h_size)
	w.WriteHeader(200)
	w.Write([]byte("Welcome"))
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qKey := vars["key"]
	Logger.Debug("Key: ", qKey)

	key := strings.ToLower(qKey)
	f := GetFileFromRPC(key)
	if f == nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(404)
		w.Write([]byte("404 NOT FOUND"))
	} else {
		fmeta := strings.Split(string(f.Meta), ";")
		h_mime := "image/jpeg"
		h_size := "0"
		h_expires := "7200"
		if len(fmeta) == 2 {
			if len(fmeta[0]) > 1 {
				h_mime = fmeta[0]
			}
			if len(fmeta[1]) > 1 {
				h_size = fmeta[1]
			}
		} else {
			Logger.Debug("no mime and size meta was found.")
			h_size = strconv.Itoa(len(f.Data))
		}

		if CFG.Http.Expire_seconds != "" {
			h_expires = CFG.Http.Expire_seconds
		}

		w.Header().Set("Content-Type", h_mime)
		w.Header().Set("Cache-Control", strings.Join([]string{"max-age", h_expires}, "="))
		w.Header().Set("Content-Length", h_size)
		w.WriteHeader(200)
		w.Write(f.Data)
	}

}

func init() {
	rootCmd.AddCommand(httpCmd)

	//logger.Info("masterIp")

	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/s/{key}", FileHandler)
	//r.HandleFunc("/list/{database}/{bucket}", PhotoListHandler)

	// r.HandleFunc("/articles/{category}/", ArticlesCategoryHandler)
	// r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
	//r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
}

func GetFileFromRPC(key string) *pbv.File {

	addr := strings.Join([]string{CFG.Http.Beip, CFG.Http.Beport}, ":")
	Logger.Info("----", addr)
	conn, _ := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(GRPCMAXMSGSIZE)))
	defer conn.Close()
	client := pbv.NewVolumeServiceClient(conn)
	//	pong, err := client.Ping(context.Background(), &pbv.Pong{Factor: "SmokeTest"})

	//	Logger.Info("Ping Test: ", pong.Message)

	file, err := client.ReadFile(context.Background(), &pbv.File{Key: key})

	if err != nil {
		Logger.Error("cannot read file from RPC. ", err)
		return nil
	} else {
		Logger.Info("Read the file:")
	}
	return file
}
