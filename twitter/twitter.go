package twitter

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"gopkg.in/ini.v1"
)

type ConfigList struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
}

type Result struct {
	ScreenName    string
	CreatedAt     string
	MediaUrlHttps string
}

var Config ConfigList
var Section string = "DEFAULT"
var IniFile string = ".twitter/config.ini"

func init() {
	home := os.Getenv("HOME")
	ini_path := fmt.Sprintf("%s/%s", home, IniFile)
	cfg, err := ini.Load(ini_path)
	if err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	Config = ConfigList{
		consumerKey:       cfg.Section(Section).Key("CONSUMER_KEY").String(),
		consumerSecret:    cfg.Section(Section).Key("CONSUMER_SECRET").String(),
		accessToken:       cfg.Section(Section).Key("ACCESS_TOKEN").String(),
		accessTokenSecret: cfg.Section(Section).Key("ACCESS_TOKEN_SECRET").String(),
	}
}

func Search(user *string, count int) []Result {
	var results []Result
	anaconda.SetConsumerKey(Config.consumerKey)
	anaconda.SetConsumerSecret(Config.consumerSecret)
	api := anaconda.NewTwitterApi(Config.accessToken, Config.accessTokenSecret)

	v := url.Values{}
	v.Set("count", strconv.Itoa(count))

	var from string = ""
	if len(*user) != 0 {
		from = fmt.Sprintf("from:%s ", *user)
	}
	keyword := fmt.Sprintf("%s#RingFitAdventure -filter:retweets filter:twimg", from)
	searchResult, _ := api.GetSearch(keyword, v)
	for _, tweet := range searchResult.Statuses {
		result := Result{
			ScreenName:    tweet.User.ScreenName,
			CreatedAt:     tweet.CreatedAt,
			MediaUrlHttps: tweet.Entities.Media[0].Media_url_https,
		}
		results = append(results, result)
	}

	return results
}
