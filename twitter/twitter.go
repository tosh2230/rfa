package twitter

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"gopkg.in/ini.v1"
)

type CfgList struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
}

type Rslt struct {
	ScreenName    string
	CreatedAt     string
	MediaUrlHttps []string
}

var Cfg CfgList
var Section string = "DEFAULT"
var IniFile string = ".twitter/config.ini"

func init() {
	home := os.Getenv("HOME")
	iniPath := fmt.Sprintf("%s/%s", home, IniFile)
	iniCfg, err := ini.Load(iniPath)
	if err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	Cfg = CfgList{
		consumerKey:       iniCfg.Section(Section).Key("CONSUMER_KEY").String(),
		consumerSecret:    iniCfg.Section(Section).Key("CONSUMER_SECRET").String(),
		accessToken:       iniCfg.Section(Section).Key("ACCESS_TOKEN").String(),
		accessTokenSecret: iniCfg.Section(Section).Key("ACCESS_TOKEN_SECRET").String(),
	}
}

func Search(user *string, count int, lastExecutedAt time.Time) []Rslt {
	const hashTag string = "RingFitAdventure"
	const filterIn string = "twimg"
	const filterEx string = "retweets"
	var rslts []Rslt

	anaconda.SetConsumerKey(Cfg.consumerKey)
	anaconda.SetConsumerSecret(Cfg.consumerSecret)
	api := anaconda.NewTwitterApi(Cfg.accessToken, Cfg.accessTokenSecret)

	keyword := fmt.Sprintf("from:%s #%s filter:%s -filter:%s", *user, hashTag, filterIn, filterEx)
	// keyword := fmt.Sprintf("from:%s #%s filter:%s -filter:%s since:%s", *user, hashTag, filterIn, filterEx, lastExecutedAt.Format("2006-01-02_15:04:05_MST"))
	v := url.Values{}
	v.Set("count", strconv.Itoa(count))

	searchResult, _ := api.GetSearch(keyword, v)
	for _, tweet := range searchResult.Statuses {
		var urls []string
		for _, medium := range tweet.ExtendedEntities.Media {
			urls = append(urls, medium.Media_url_https)
		}
		rslt := Rslt{
			ScreenName:    tweet.User.ScreenName,
			CreatedAt:     tweet.CreatedAt,
			MediaUrlHttps: urls,
		}
		rslts = append(rslts, rslt)
	}

	return rslts
}

func GetImage(url string) *os.File {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	urlSliced := strings.Split(url, "/")
	fileName := fmt.Sprintf(urlSliced[len(urlSliced)-1])
	file, _ := ioutil.TempFile("", fileName)
	io.Copy(file, response.Body)

	return file
}
