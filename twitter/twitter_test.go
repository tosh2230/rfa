package twitter

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestGetFromSecretManager(t *testing.T) {
	// Require
	// 1. Set Env TEST_PROJECT_ID
	// 2. Set SecretManager projects/<project>/secrets/rfa-test

	var err error

	pj := os.Getenv("TEST_PROJECT_ID")
	if pj == "" {
		err = fmt.Errorf("Env TEST_PROJECT_ID is empty.")
		t.Error(err)
	}

	var want = "TestGetFromSecretManager"
	var inputSecretID = "rfa-test"
	var inputSecretVersion = "latest"

	outByte, err := getFromSecretManager(pj, inputSecretID, inputSecretVersion)
	if err != nil {
		t.Error(err)
	}
	out := fmt.Sprintf("%s", outByte)

	if out != want {
		err = errors.Errorf("Test data is different.\n[out ]: %s\n[want]: %s", out, want)
		t.Error(err)
	}
}

func TestSetTwitterConfig(t *testing.T) {
	var err error

	var out CfgList
	var want CfgList = CfgList{
		ConsumerKey:       "TestSetTwitterConfig.consumer_key",
		ConsumerSecret:    "TestSetTwitterConfig.consumer_secret",
		AccessToken:       "TestSetTwitterConfig.access_token",
		AccessTokenSecret: "TestSetTwitterConfig.access_token_secret",
	}

	input, err := ioutil.ReadFile("testdata/TestSetTwitterConfig.json")
	if err != nil {
		t.Error(err)
	}

	err = out.setTwitterConfig(input)
	if err != nil {
		t.Error(err)
	}

	if out != want {
		err = errors.Errorf("twitter.CfgList is different.\n[out ]: %v\n[want]: %v", out, want)
		t.Error(err)
	}
}
