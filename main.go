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

var metadataURL string

var allowableObjects = []string{"hosts", "stacks", "services", "containers"}

func main() {
	// Parse the flags
	first := flag.Bool("first", false, "Select first entry only")
	csv := flag.Bool("csv", false, "Return as comma separated values")
	hostname := flag.String("hostname", "rancher-metadata", "Custom rancher-metadata hostname")
	flag.Parse()

	metadataURL = fmt.Sprintf("http://%v/latest/", *hostname)

	tail := flag.Args()

	valid := true
	errorMessage := ""
	if len(tail) != 3 {
		valid = false
		errorMessage = "Invlalid number of arguments"
	}

	// The object class specified must be an allowed object
	if valid {
		validObjClass := false
		objClass := tail[0]
		for i := range allowableObjects {
			if objClass == allowableObjects[i] {
				validObjClass = true
				break
			}
		}
		valid = validObjClass
		if !valid {
			errorMessage = objClass + " is not one of " + strings.Join(allowableObjects, ",") + ".\n"
		}
	}

	if !valid {
		message := "Roundup v0.7\n" +
			"Query the rancher-metadata by label\n" +
			"%v\n" +
			"Usage: roundup [options] {hosts|stacks|services|containers} predicate value-name\nExample:\n" +
			"  Fetch first container name that has the label \"arangodb_cluster_name\" with the value \"cluster1\":\n" +
			"    roundup -first containers arangodb_cluster_name=cluster1 name\n" +
			"Options:\n"
		os.Stderr.WriteString(fmt.Sprintf(message, errorMessage))
		flag.PrintDefaults()
		return
	}

	if ret, err := query(tail[0], tail[1], tail[2], first, csv); err == nil {
		fmt.Print(ret)
	} else {
		panic(err)
	}
}
func query(obj string, predicateString string, valueName string, first *bool, csv *bool) (string, error) {

	// Loop through the predicates splitting them up
	predicates := make(map[string]string)
	for _, i := range strings.Split(predicateString, ",") {
		predicate := strings.Split(i, "=")
		predicates[predicate[0]] = predicate[1]
	}
	npredicates := len(predicates)

	// Request all of the objects of the class specified as JSON from the metadata URL
	buf, err := httpGet(obj, true)
	if err != nil {
		return "", err
	}

	// Deserialize the objects into LabeledObject structures that contain the name and labels.
	var objects []LabeledObject
	json.Unmarshal(buf, &objects)

	// Loop through the labeled objects
	ret := make([]string, 0, 0)
	for _, o := range objects {

		// Check each predicate
		criteria := 0
		for k, v := range predicates {
			if o.Labels[k] == v {
				criteria++
			}
		}

		// If all predicates are not met then continue
		if criteria != npredicates {
			continue
		}

		// Assemble the result
		if buf, err := httpGet(obj+"/"+o.Name+"/"+valueName, false); err == nil {
			ret = append(ret, string(buf))

			// If the "first" flag was specified just return the first one we found.
			if *first {
				return ret[0], nil
			}
		} else {
			return "", err
		}
	}
	// Format and return the result
	seperator := "\n"
	if *csv {
		seperator = ","
	}

	return strings.Join(ret, seperator), nil
}

func httpGet(relativeURI string, acceptJSON bool) (buffer []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, metadataURL+relativeURI, nil)
	if err != nil {
		os.Stderr.WriteString("Failed to create request for: " + metadataURL + relativeURI)
		return nil, err
	}
	if acceptJSON {
		req.Header.Add("Accept", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		os.Stderr.WriteString("HTTP GET failed for: " + metadataURL + relativeURI)
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		os.Stderr.WriteString("Failed to read response body")
		return nil, err
	}
	return buf, nil
}
