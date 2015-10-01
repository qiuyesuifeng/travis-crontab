package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// See: http://docs.travis-ci.com/api

var (
	travisURL          string
	travisAcceptHeader string
	travisAuthHeader   string
)

type travisBuild struct {
	ID     int64  `json:"id"`
	State  string `json:"state"`
	Branch string `json:"branch"`
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

func getTravisToken(reqURL string, token string) string {
	v := url.Values{}
	v.Set("github_token", token)
	body := ioutil.NopCloser(strings.NewReader(v.Encode()))

	req, err := http.NewRequest("POST", reqURL, body)
	checkErr(err)

	req.Header.Set("User-Agent", "TravisMockAgent")
	req.Header.Set("Accept", travisAcceptHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	data := map[string]string{}
	err = json.Unmarshal(b, &data)
	checkErr(err)

	return data["access_token"]
}

func getTravisLastBuildID(reqURL string, token string, branch string) int64 {
	req, err := http.NewRequest("GET", reqURL, nil)
	checkErr(err)

	req.Header.Set("User-Agent", "TravisMockAgent")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	datas := []travisBuild{}
	err = json.Unmarshal(b, &datas)
	checkErr(err)

	for _, data := range datas {
		if data.Branch == branch {
			return data.ID
		}
	}

	return 0
}

func rebuildTravis(reqURL string, token string) string {
	req, err := http.NewRequest("POST", reqURL, nil)
	checkErr(err)

	req.Header.Set("User-Agent", "TravisMockAgent")
	req.Header.Set("Accept", travisAcceptHeader)
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	return string(b)
}

func init() {
	travisURL = "https://api.travis-ci.org"
	travisAcceptHeader = "application/vnd.travis-ci.2+json"
}

func main() {
	flag.Usage = usage
	token := flag.String("t", "", "Github Token")
	repo := flag.String("r", "", "Github Repo[username/repo]")
	branch := flag.String("b", "", "Github Repo Branch")
	flag.Parse()

	if *token == "" || *repo == "" || *branch == "" {
		usage()
	}

	travisToken := getTravisToken(fmt.Sprintf("%s/auth/github", travisURL), *token)
	travisLastBuildID := getTravisLastBuildID(fmt.Sprintf("%s/repos/%s/builds", travisURL, *repo), travisToken, *branch)
	travisResult := rebuildTravis(fmt.Sprintf("%s/builds/%d/restart", travisURL, travisLastBuildID), travisToken)

	travisRebuildURL := fmt.Sprintf("https://travis-ci.org/%s/builds/%d", *repo, travisLastBuildID)
	fmt.Println(time.Now(), travisRebuildURL, travisResult)
}
