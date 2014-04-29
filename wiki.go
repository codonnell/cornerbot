package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const apiUrl = "http://en.wikipedia.org/w/api.php"

var urlQuery = url.Values{
	"action": []string{"query"},
	"prop":   []string{"info"},
	"inprop": []string{"url"},
	"format": []string{"json"},
}

var randomQuery = url.Values{
	"action":      []string{"query"},
	"list":        []string{"random"},
	"rnlimit":     []string{"1"},
	"rnnamespace": []string{"0"},
	"format":      []string{"json"},
}

type UrlResponse struct {
	Query struct {
		Pages map[string]struct {
			Fullurl string
		}
	}
}

type RandomResponse struct {
	Query struct {
		Random []struct {
			Id float64
		}
	}
}

func RandomPageID() (int, error) {
	client := &http.Client{}
	u, err := url.Parse(apiUrl)
	if err != nil {
		log.Print("RandomPageID(): problem parsing url %s", err)
		return 0, err
	}
	u.RawQuery = randomQuery.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Print("RandomPageID(): problem creating new request %s", err.Error())
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("RandomPageID(): problem getting response %s", err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	var jsonData RandomResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&jsonData); err != nil {
		log.Print("RandomPageID(): problem parsing json %s", err)
		return 0, err
	}
	return int(jsonData.Query.Random[0].Id), nil
}

func PageUrl(id int) (string, error) {
	client := &http.Client{}
	u, err := url.Parse(apiUrl)
	if err != nil {
		log.Print("PageUrl(): problem parsing url %s", err)
		return "", err
	}
	query := make(url.Values)
	for k, v := range urlQuery {
		query[k] = v
	}
	query.Add("pageids", strconv.Itoa(id))
	u.RawQuery = query.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Print("PageUrl(): problem creating new request %s", err.Error())
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("PageUrl(): problem getting response %s", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	var jsonData UrlResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&jsonData); err != nil {
		log.Print("PageUrl(): problem parsing json %s", err)
		return "", err
	}
	val, ok := jsonData.Query.Pages[strconv.Itoa(id)]
	if ok != true {
		log.Print("PageUrl(): no pageid found")
		return "", errors.New("no pageid found")
	}
	return val.Fullurl, nil
}
