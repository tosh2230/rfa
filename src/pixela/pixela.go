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

const secretVersion string = "latest"

func GetConfig(pj string, secretID string) (cfg CfgList, err error) {
	data, err := getFromSecretManager(pj, secretID, secretVersion)
	if err != nil {
		return cfg, nil
	}
	err = cfg.setPixelaConfig(data)
	if err != nil {
		return
	}
	return
}

func (cfg *CfgList) Grow(CreatedAt time.Time) (resp *http.Response, err error) {
	url := "https://pixe.la/v1/users/" + cfg.User + "/graphs/" + cfg.GraphId
	jsonStr := `{"date":"` + CreatedAt.Format("20060102") + `","quantity":"1"}`

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-USER-TOKEN", cfg.Token)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, err
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
		err = fmt.Errorf("pixela.CfgList.User is empty")
		return
	}
	if cfg.GraphId == "" {
		err = fmt.Errorf("pixela.CfgList.GraphId is empty")
		return
	}
	if cfg.Token == "" {
		err = fmt.Errorf("pixela.CfgList.Token is empty")
		return
	}

	return
}
