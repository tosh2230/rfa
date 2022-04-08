package main

import (
	"fmt"
	"os"

	"github.com/tosh223/rfa/search"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "rfa",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.PersistentFlags().GetString("project-id")
		location, _ := cmd.PersistentFlags().GetString("location")
		twitterID, _ := cmd.PersistentFlags().GetString("twitter-id")
		sizeStr, _ := cmd.PersistentFlags().GetString("search-size")

		var rfa search.Rfa
		rfa.ProjectID = projectID
		rfa.Location = location
		rfa.TwitterID = twitterID
		rfa.Size = sizeStr
		rfa.Search()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("project-id", "p", "", "GCP Project ID")
	rootCmd.PersistentFlags().StringP("location", "l", "us", "BigQuery location")
	rootCmd.PersistentFlags().StringP("twitter-id", "u", "", "Twitter ID")
	rootCmd.PersistentFlags().StringP("search-size", "s", "1", "search size")
}
