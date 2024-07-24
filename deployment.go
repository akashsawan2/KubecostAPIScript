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


	header := []string{"Deployment", "Region", "Namespace" ,"Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost" ,"Cpu Efficiency","Ram Efficiency","Total Efficiency"}
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

			cpuCost := deploymentOne["cpuCost"].(float64)
			gpuCost := deploymentOne["gpuCost"].(float64)
			ramCost := deploymentOne["ramCost"].(float64)
			pvCost  := deploymentOne["pvCost"].(float64)
			networkCost := deploymentOne["networkCost"].(float64)
			loadBalancerCost := deploymentOne["loadBalancerCost"].(float64)
			sharedCost := deploymentOne["sharedCost"].(float64)

			totalCost := deploymentOne["totalCost"].(float64)

			cpuEfficiency := deploymentOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := deploymentOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := deploymentOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100

			record := []interface{}{name, region, namespace_deployment , windowStart, windowEnd,cpuCost,gpuCost,ramCost,pvCost,networkCost,loadBalancerCost,sharedCost,totalCost,cpuEfficiency,ramEfficiency,totalEfficiency}
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
