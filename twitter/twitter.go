package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tosh223/rfa/gcpsecretmanager"

	"github.com/ChimeraCoder/anaconda"
)

type CfgList struct {
	ConsumerKey       string `json:"consumer_key,omitempty"`
	ConsumerSecret    string `json:"consumer_secret,omitempty"`
	AccessToken       string `json:"access_token,omitempty"`
	AccessTokenSecret string `json:"access_token_secret,omitempty"`
}

type Rslt struct {
	ScreenName    string
	CreatedAt     time.Time
	MediaUrlHttps []string
}

const secretVersion string = "latest"

func GetConfig(pj string, secretID string) (cfg CfgList, err error) {
	data, err := getFromSecretManager(pj, secretID, secretVersion)
	if err != nil {
		return
	}
	err = cfg.setTwitterConfig(data)
	if err != nil {
		return
	}
	return
}

func (cfg *CfgList) Search(user *string, count int, lastExecutedAt time.Time) (rslts []Rslt, err error) {
	const hashTag string = "RingFitAdventure"
	const filterIn string = "twimg"
	// const filterEx string = "retweets"
	var keyword string

	anaconda.SetConsumerKey(cfg.ConsumerKey)
	anaconda.SetConsumerSecret(cfg.ConsumerSecret)
	api := anaconda.NewTwitterApi(cfg.AccessToken, cfg.AccessTokenSecret)

	if lastExecutedAt == (time.Time{}) {
		// keyword = fmt.Sprintf("from:%s #%s filter:%s -filter:%s", *user, hashTag, filterIn, filterEx)
		keyword = fmt.Sprintf("from:%s #%s filter:%s", *user, hashTag, filterIn)
	} else {
		searchTime := lastExecutedAt.Add(1 * time.Second)
		// keyword = fmt.Sprintf("from:%s #%s filter:%s -filter:%s since:%s", *user, hashTag, filterIn, filterEx, searchTime.Format("2006-01-02_15:04:05_MST"))
		keyword = fmt.Sprintf("from:%s #%s filter:%s since:%s", *user, hashTag, filterIn, searchTime.Format("2006-01-02_15:04:05_MST"))
	}

	v := url.Values{}
	v.Set("count", strconv.Itoa(count))

	searchResult, err := api.GetSearch(keyword, v)
	if err != nil {
		return
	}
	fmt.Println("Search results:")
	fmt.Println(searchResult)

	for _, tweet := range searchResult.Statuses {
		var urls []string
		for _, medium := range tweet.ExtendedEntities.Media {
			urls = append(urls, medium.Media_url_https)
		}
		createdAt, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.CreatedAt)
		rslt := Rslt{
			ScreenName:    tweet.User.ScreenName,
			CreatedAt:     createdAt,
			MediaUrlHttps: urls,
		}
		rslts = append(rslts, rslt)
	}

	return
}

func GetImage(url string) (file *os.File, err error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	urlSliced := strings.Split(url, "/")
	fileName := fmt.Sprintf(urlSliced[len(urlSliced)-1])
	file, err = ioutil.TempFile("", fileName)
	if err != nil {
		return nil, err
	}
	io.Copy(file, response.Body)

	return file, err
}

func getFromSecretManager(pj string, secID string, ver string) (data []byte, err error) {
	var secMgr gcpsecretmanager.SecretManager

	secMgr.ProjectID = pj
	secMgr.SecretID = secID
	secMgr.Version = ver

	data, err = secMgr.Access()
	if err != nil {
		return
	}

	return
}

func (cfg *CfgList) setTwitterConfig(b []byte) (err error) {
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return
	}
	if cfg.ConsumerKey == "" {
		err = fmt.Errorf("twitter.CfgList.ConsumerKey is empty")
		return
	}
	if cfg.ConsumerSecret == "" {
		err = fmt.Errorf("twitter.CfgList.ConsumerSecret is empty")
		return
	}
	if cfg.AccessToken == "" {
		err = fmt.Errorf("twitter.CfgList.AccessToken is empty")
		return
	}
	if cfg.AccessTokenSecret == "" {
		err = fmt.Errorf("twitter.CfgList.AccessTokenSecret is empty")
		return
	}

	return
}
