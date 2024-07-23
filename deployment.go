package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
)


func init() {
    InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func FetchAndWriteDeploymentData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "deployment")
	q.Set("accumulate", "true")
	u.RawQuery = q.Encode()


	newURL := u.String()


	resp, err := http.Get(newURL)
	if err != nil {
		ErrorLogger.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ErrorLogger.Println("Error reading response body:", err)
		return
	}


	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		ErrorLogger.Println("Error unmarshalling JSON:", err)
		return
	}

	InfoLogger.Println("Status Code for Deployment:", result["code"])

	data := result["data"].([]interface{})


	f, err := excelize.OpenFile(filePath)
	if err != nil {
		ErrorLogger.Println("Error opening file:", err)
		return
	}


	if _, err := f.NewSheet("Deployment"); err != nil {
		ErrorLogger.Println("Error creating 'Deployment' sheet:", err)
		return
	}


	header := []string{"Name", "Region", "Namespace" ,"Window Start", "Window End","Total Cost" ,"Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) 
		if err := f.SetCellValue("Deployment", cell, h); err != nil {
			ErrorLogger.Println("Error writing header:", err)
			return
		}
	}


	row := 2
	for _, element := range data {
		deploymentMap := element.(map[string]interface{})


		for _, DeploymentData := range deploymentMap {
			deploymentOne := DeploymentData.(map[string]interface{})

			name := deploymentOne["name"].(string)
			if name =="__unallocated__" {
				continue
			}

			properties := deploymentOne["properties"].(map[string]interface{})

			labels := properties["labels"].(map[string]interface{})
			region := labels["topology_kubernetes_io_region"].(string)

			namespace_deployment := labels["kubernetes_io_metadata_name"]

			window := deploymentOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			var totalCost float64
			if cost, ok := deploymentOne["totalCost"].(float64);ok {
				totalCost = cost
			} else {
				ErrorLogger.Println("Error fetching cost data")
			}

			var totalEfficiency float64
			if efficiency, ok := deploymentOne["totalEfficiency"].(float64); ok {
				totalEfficiency = efficiency * 100
			} else {
				ErrorLogger.Println("Error fetching efficiency data")
			}

			record := []interface{}{name, region, namespace_deployment , windowStart, windowEnd,fmt.Sprintf("%.2f", totalCost) ,fmt.Sprintf("%.2f", totalEfficiency)}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) 
				if err := f.SetCellValue("Deployment", cell, val); err != nil {
					ErrorLogger.Println("Error writing record :", err)
					return
				}
			}
			row++
		}
	}


	if err := f.SaveAs(filePath); err != nil {
		ErrorLogger.Println("Error saving file:", err)
		return
	}

	InfoLogger.Println("Deployment Data successfully written")
}
