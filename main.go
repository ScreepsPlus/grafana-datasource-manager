package main

import (
    "bytes"
	"encoding/json"
	"fmt"
    "io/ioutil"
	"net/http"
	"os"
    m "github.com/grafana/grafana/pkg/models"
)

var grafanaUrl string = os.Getenv("GRAFANA_URL")

func main() {
	fmt.Print("Checking orgs...\n")
	orgs, err := GetOrgs()
	if err != nil { 
		fmt.Print(err)	
	}
	fmt.Printf("Found %d\n", len(orgs))
	for _, org := range orgs {
		if org.Id == 1 { 
			continue 
		}
		ds, err := GetDataSourceByName(org.Name, "ScreepsPlus-Graphite")
		if err != nil { 
			fmt.Print(err)
			continue
		}
		if ds.Name == "" {
			data := m.AddDataSourceCommand{
				Name: "ScreepsPlus-Graphite",
				Type: "graphite",
				Access: "direct",
				Url: "https://screepspl.us/grafana/carbonapi",
				WithCredentials: true,
			}
			ds, _ := AddDataSource(org.Name, data)
			fmt.Printf("DataSource added %s %v\n", org.Name, ds)
		}
	}
}

func GetOrgs() ([]m.OrgDTO,error) {
    ret := make([]m.OrgDTO,0)
    req, _ := http.NewRequest("GET",fmt.Sprintf("%s/api/orgs", grafanaUrl),nil)
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

func GetDataSourceByName(user string, name string) (m.DataSource,error) {
    ret := m.DataSource{}
    req, _ := http.NewRequest("GET",fmt.Sprintf("%s/api/datasources/name/%s", grafanaUrl, name),nil)
    req.Header.Add("Token-Claim-Sub", user)
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
    	return ret, err
    }
    defer res.Body.Close()
    data, err := ioutil.ReadAll(res.Body)
    err = json.Unmarshal(data, &ret)
    if err != nil { fmt.Printf("%s %v\n",string(data),err)}
    return ret, err   
}

func AddDataSource(user string, data m.AddDataSourceCommand) (m.DataSource,error) {
    jsonStr, _ := json.Marshal(data)
    req, _ := http.NewRequest("POST",fmt.Sprintf("%s/api/datasources", grafanaUrl),bytes.NewBuffer(jsonStr))
    req.Header.Add("Token-Claim-Sub", user)
    ret := m.DataSource{}
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Content-Length", string(len(jsonStr)))
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
    	return ret, err
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    err = json.Unmarshal(body, &ret)
    if err != nil { fmt.Printf("%s %v\n",string(body),err)}
    return ret, err
}