package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var grafanaUrl string = os.Getenv("GRAFANA_URL")
var influxdbUrl string = os.Getenv("INFLUXDB_URL")

func main() {
	orgs, err := GetOrgs()
	if err != nil {
		fmt.Print(err)
	}
	for _, org := range orgs {
		if org.Id == 1 {
			continue
		}
		if org.Name != strings.ToLower(org.Name) {
			newName := strings.ToLower(org.Name)
			log.Printf("Updating name for org %s to %s", org.Name, newName)
			if err := UpdateOrg(org.Id, m.UpdateOrgCommand{Name: newName}); err != nil {
				log.Printf("%v", err)
			}
			org.Name = newName
		}

		if ds, _ := GetDataSourceByName(org.Name, "ScreepsPlus-Graphite"); ds == nil {
			data := m.AddDataSourceCommand{
				Name:            "ScreepsPlus-Graphite",
				Type:            "graphite",
				Access:          "direct",
				Url:             "https://screepspl.us/grafana/carbonapi",
				WithCredentials: true,
				JsonData:        &simplejson.Json{},
			}
			data.JsonData.Set("graphiteVersion", "1.1.x")
			if ds, err := AddDataSource(org.Name, data); err != nil {
				log.Printf("DataSource add failed for %s %v\n", org.Name, err)
			} else {
				log.Printf("DataSource %s added for %s\n", ds.Name, org.Name)
			}
		}

		if ds, _ := GetDataSourceByName(org.Name, "ScreepsPlus-InfluxDB"); ds == nil {
			data := m.AddDataSourceCommand{
				Name:            "ScreepsPlus-InfluxDB",
				Type:            "influxdb",
				Access:          "proxy",
				Url:             influxdbUrl,
				WithCredentials: true,
			}
			if ds, err := AddDataSource(org.Name, data); err != nil {
				log.Printf("DataSource add failed for %s %v\n", org.Name, err)
			} else {
				log.Printf("DataSource %s added for %s\n", ds.Name, org.Name)
			}
		}
	}
}

func GetOrgs() ([]m.OrgDTO, error) {
	ret := make([]m.OrgDTO, 0)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/orgs", grafanaUrl), nil)
	req.Header.Add("Token-Claim-Sub", "admin")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func GetDataSourceByName(user string, name string) (*m.DataSource, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/datasources/name/%s", grafanaUrl, name), nil)
	req.Header.Add("Token-Claim-Sub", user)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	switch res.StatusCode {
	case 200:
		ret := m.DataSource{}
		err := dec.Decode(&ret)
		if err != nil {
			log.Printf("Error Decoding: %v", err)
			return nil, err
		}
		return &ret, nil
	default:
		ret := struct{ Message string }{}
		err := dec.Decode(&ret)
		if err != nil {
			log.Printf("Error Decoding: %v", err)
			return nil, err
		}
		log.Printf("Error Getting Datasource %s for %s: %s\n", name, user, ret.Message)
		return nil, fmt.Errorf("Error: %s", ret.Message)
	}
}

func AddDataSource(user string, data m.AddDataSourceCommand) (*m.DataSource, error) {
	jsonStr, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/datasources", grafanaUrl), bytes.NewBuffer(jsonStr))
	req.Header.Add("Token-Claim-Sub", user)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(jsonStr)))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	switch res.StatusCode {
	case 200:
		ret := m.DataSource{}
		err := dec.Decode(&ret)
		if err != nil {
			log.Printf("Error Decoding: %v", err)
			return nil, err
		}
		return &ret, nil
	default:
		ret := struct{ Message string }{}
		err := dec.Decode(&ret)
		if err != nil {
			log.Printf("Error Decoding: %v", err)
			return nil, err
		}
		log.Printf("Error Adding Datasource for %s: %s\n", user, ret.Message)
		return nil, fmt.Errorf("Error: %s", ret.Message)
	}
}

func UpdateOrg(orgId int64, data m.UpdateOrgCommand) error {
	jsonStr, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/orgs/%d", grafanaUrl, orgId), bytes.NewBuffer(jsonStr))
	req.Header.Add("Token-Claim-Sub", "admin")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(jsonStr)))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	switch res.StatusCode {
	case 200:
		return nil
	default:
		ret := struct{ Message string }{}
		err := dec.Decode(&ret)
		if err != nil {
			log.Printf("Error Decoding: %v", err)
			return err
		}
		log.Printf("Error Updating org %d: %s\n", orgId, ret.Message)
		return fmt.Errorf("Error: %s", ret.Message)
	}
}
