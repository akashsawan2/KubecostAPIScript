package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWriteControllerData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "controller")
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

	InfoLogger.Println("Status Code for Controller: ", result["code"])

	data := result["data"].([]interface{})


	f, err := excelize.OpenFile(filePath)
	if err != nil {
		ErrorLogger.Println("Error opening file:", err)
		return
	}


	if _, err := f.NewSheet("Controller"); err != nil {
		ErrorLogger.Println("Error creating 'Controller' sheet:", err)
		return
	}


	header := []string{"Controller", "Region", "Namespace" ,"Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency","Ram Efficiency","Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) 
		if err := f.SetCellValue("Controller", cell, h); err != nil {
			ErrorLogger.Println("Error writing header:", err)
			return
		}
	}


	row := 2
	for _, element := range data {
		ControllerMap := element.(map[string]interface{})


		for _, ControllerData := range ControllerMap {
			ControllerOne := ControllerData.(map[string]interface{})

			name := ControllerOne["name"].(string)
			if name =="__unallocated__" {
				continue
			}

			properties := ControllerOne["properties"].(map[string]interface{})

			var region string
			if name != "__idle__"{
				labels := properties["labels"].(map[string]interface{})
			
				region = labels["topology_kubernetes_io_region"].(string)
			}

			namespace_labels := properties["namespaceLabels"].(map[string]interface{})
			namespace_controller := namespace_labels["kubernetes_io_metadata_name"]

			window := ControllerOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			cpuCost := ControllerOne["cpuCost"].(float64)
			gpuCost := ControllerOne["gpuCost"].(float64)
			ramCost := ControllerOne["ramCost"].(float64)
			pvCost  := ControllerOne["pvCost"].(float64)
			networkCost := ControllerOne["networkCost"].(float64)
			loadBalancerCost := ControllerOne["loadBalancerCost"].(float64)
			sharedCost := ControllerOne["sharedCost"].(float64)

			totalCost := ControllerOne["totalCost"].(float64)

			cpuEfficiency := ControllerOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := ControllerOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := ControllerOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100


			record := []interface{}{name, region, namespace_controller , windowStart, windowEnd,cpuCost,gpuCost,ramCost,pvCost,networkCost,loadBalancerCost,sharedCost,totalCost,cpuEfficiency,ramEfficiency,totalEfficiency}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) 
				if err := f.SetCellValue("Controller", cell, val); err != nil {
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

	InfoLogger.Println("Controller Data successfully written")
}
