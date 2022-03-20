/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"

	//"strings"

	"github.com/harryzhu/potatofs/util"
	"github.com/spf13/cobra"
)

var (
	DIR      string
	FILTER   string
	USERNAME string
	ALBUM    string
	TAGS     string
	COMMENT  string
)

// fputCmd represents the fput command
var fputCmd = &cobra.Command{
	Use:   "fput",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fput called")
		if util.ValidateANS(USERNAME) == false {
			log.Fatal("username cannot include: space, /,\\")
		}

		if util.ValidateANS(ALBUM) == false {
			log.Fatal("album cannot include: space, /,\\ ")
		}

		// if strings.Contains(USERNAME, "/") || strings.Contains(ALBUM, "/") {
		// 	log.Fatal("--username and --album cannot contain '/'")
		// }
		// if strings.Contains(USERNAME, "\\") || strings.Contains(ALBUM, "\\") {
		// 	log.Fatal("--username and --album cannot contain '\\'")
		// }
		// if strings.Contains(USERNAME, " ") || strings.Contains(ALBUM, " ") {
		// 	log.Fatal("--username and --album cannot contain ' '")
		// }

		Fput()
	},
}

func init() {
	rootCmd.AddCommand(fputCmd)

	fputCmd.Flags().StringVar(&DIR, "dir", "/Volumes/HDD4/downloadonly/seven", "--dir=/path/of/your/files")
	fputCmd.Flags().StringVar(&FILTER, "filter", "*.jpg", "--filter=*.jpg")
	fputCmd.Flags().StringVar(&USERNAME, "username", "u0", "--username=user's name")
	fputCmd.Flags().StringVar(&ALBUM, "album", "a0", "--album=sport")
	fputCmd.Flags().StringVar(&TAGS, "tags", "", "--tags=\"tag1,tag2,tag3,tag4\"")
	fputCmd.Flags().StringVar(&COMMENT, "comment", "", "--comment=comment-of-file")

}
