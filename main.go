package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var metadataURL = "http://rancher-metadata/latest/"

func main() {
	// Parse the flags
	first := flag.Bool("first", false, "Select first entry only")
	csv := flag.Bool("csv", false, "Return as comma separated values")
	flag.Parse()
	tail := flag.Args()
	if len(tail) != 3 {
		os.Stderr.WriteString("Roundup v0.7\n")
		os.Stderr.WriteString("Query the rancher-metadata by label\n")
		os.Stderr.WriteString("Invalid number of arguments\n")
		os.Stderr.WriteString("Usage: roundup [options] {hosts|stacks|services|containers} predicate value-name \n")
		os.Stderr.WriteString("Example:\n")
		os.Stderr.WriteString("  Fetch first container name that has the label \"arangodb_cluster_name\" with the value \"cluster1\":\n")
		os.Stderr.WriteString("  roundup -first containers arangodb_cluster_name=cluster1 name \n")
		os.Stderr.WriteString("Options:\n")

		flag.PrintDefaults()
		return
	}
	fmt.Print(query(tail[0], tail[1], tail[2], first, csv))
}
func query(obj string, predicateString string, valueName string, first *bool, csv *bool) string {
	ret := make([]string, 0, 0)
	predicates := make(map[string]string)
	for _, i := range strings.Split(predicateString, ",") {
		predicate := strings.Split(i, "=")
		predicates[predicate[0]] = predicate[1]
	}
	npredicates := len(predicates)
	req, err := http.NewRequest(http.MethodGet, metadataURL+obj, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var objects []LabeledObject
	json.Unmarshal(buf, &objects)
	for _, o := range objects {
		criteria := 0
		for k, v := range predicates {
			if o.Labels[k] == v {
				criteria++
			}
		}
		if criteria == npredicates {
			req, err := http.NewRequest(http.MethodGet, metadataURL+obj+"/"+o.Name+"/"+valueName, nil)
			if err != nil {
				panic(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			ret = append(ret, string(buf))
			if *first {
				return ret[0]
			}
		}
	}
	if *csv {
		return strings.Join(ret, ",")
	}

	return strings.Join(ret, "\n")
}
