package pixela

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tosh223/rfa/gcpsecretmanager"
)

type CfgList struct {
	User    string `json:"user,omitempty"`
	GraphId string `json:"graph_id,omitempty"`
	Token   string `json:"token,omitempty"`
}

const secretID string = "rfa_pixela"
const secretVersion string = "latest"

func GetConfig(pj string) (cfg CfgList, err error) {
	data, err := getFromSecretManager(pj, secretID, secretVersion)
	if err != nil {
		return
	}
	err = cfg.setPixelaConfig(data)
	if err != nil {
		return
	}
	return
}

func (cfg *CfgList) Grow(CreatedAtStr string) (err error) {
	url := "https://pixe.la/v1/users/" + cfg.User + "/graphs/" + cfg.GraphId
	createdAt, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", CreatedAtStr)
	jsonStr := `{"date":"` + createdAt.Format("20060102") + `","quantity":"1"}`

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return err
	}
	req.Header.Set("X-USER-TOKEN", cfg.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
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

func (cfg *CfgList) setPixelaConfig(b []byte) (err error) {
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return
	}
	if cfg.User == "" {
		err = fmt.Errorf("twitter.CfgList.ConsumerKey is empty")
		return
	}
	if cfg.GraphId == "" {
		err = fmt.Errorf("twitter.CfgList.ConsumerSecret is empty")
		return
	}
	if cfg.Token == "" {
		err = fmt.Errorf("twitter.CfgList.AccessToken is empty")
		return
	}

	return
}
