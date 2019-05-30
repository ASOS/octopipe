package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func logAndExitf(message string, a ...interface{}) {
	fmt.Printf(message+"\n", a...)
	os.Exit(1)
}

func doOctopusRequest(body []byte, uri string, method string) (responsebody []byte, status int) {

	httpreq, _ := http.NewRequest(method, uri, nil)
	httpreq.Header.Set("X-Octopus-ApiKey", apiKey)

	if method == "GET" {

		response, err := client.Do(httpreq)
		if err != nil {
			logAndExitf("Error executing request %s to %s:\n%s", method, uri, err.Error())
		}

		defer response.Body.Close()

		responsebody, err = ioutil.ReadAll(response.Body)
		if err != nil {
			logAndExitf("Error reading response body of GET to %s:\n%s", uri, err.Error())
		}

		return body, response.StatusCode

	} else if method == "PUT" || method == "POST" {
		httpreq, _ := http.NewRequest(method, uri, bytes.NewBuffer(body))
		httpreq.Header.Set("X-Octopus-ApiKey", apiKey)

		response, err := client.Do(httpreq)
		if err != nil {
			logAndExitf("Error executing request %s to %s:\n%s", method, uri, err.Error())
		}

		defer response.Body.Close()

		responsebody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logAndExitf("Error reading response body of PUT to %s:\n%s", uri, err.Error())
		}

		return responsebody, response.StatusCode
	}

	return []byte("Method " + method + " not supported"), 0
}

func verifyVariableType(svtype variable) (vtype string, err error) {
	for _, validType := range validVariableTypes {
		if validType == svtype.Type {
			return svtype.Type, nil
		}
	}

	if svtype.Type == "" {
		return "String", nil
	}

	var errorstring string
	for _, vvar := range validVariableTypes {
		errorstring = errorstring + vvar + ","
	}
	errorstring = strings.TrimSuffix(errorstring, ",")
	return "", errors.New("Variable type '" + svtype.Type + "' for variable '" + svtype.Name + "' is not valid. Valid types are " + errorstring)
}

func verifySyntaxType(satype step) (atype string, err error) {
	for _, validType := range validScriptSyntaxTypes {
		if validType == satype.Type {
			return satype.Type, nil
		}
	}

	var errorstring string
	for _, vsyn := range validScriptSyntaxTypes {
		errorstring = errorstring + vsyn + ","
	}
	errorstring = strings.TrimSuffix(errorstring, ",")
	return "", errors.New("Script syntax type '" + satype.Type + "' for process step '" + satype.Name + "' is not valid. Syntax types are case sensitive. Valid types are " + errorstring)
}

func getProjectSlug(name string) (slug string) {
	name = strings.ToLower(name)
	regex := regexp.MustCompile("[^0-9A-Za-z]+")

	slugged := regex.ReplaceAll([]byte(name), []byte("-"))

	return string(slugged)
}

func getLifecycle(ls []octopusLifecycle, name string) (l octopusLifecycle, err error) {
	for _, l := range ls {
		if l.Name == name {
			return l, nil
		}
	}
	return octopusLifecycle{}, errors.New("Lifecycle with name " + name + " not found")
}

func getProjectGroup(pgs []octopusProjectGroup, name string) (pg octopusProjectGroup, err error) {
	for _, pg := range pgs {
		if pg.Name == name {
			return pg, nil
		}
	}
	return octopusProjectGroup{}, errors.New("Project group with name " + name + " not found")
}
