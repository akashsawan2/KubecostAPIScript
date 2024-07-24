package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWriteClusterData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "cluster")
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

	InfoLogger.Println("Status Code for Cluster : ", result["code"])

	data := result["data"].([]interface{})

	
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		ErrorLogger.Println("Error opening file:", err)
		return
	}

	
	header := []string{"Cluster", "Region", "Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency", "Ram Efficiency" ,"Total Efficiency"}
	sheetName := "Cluster"
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			ErrorLogger.Println("Error writing header:", err)
			return
		}
	}

	
	row := 2
	for _, element := range data {
		clusterMap := element.(map[string]interface{})

		
		for _, clusterData := range clusterMap {
			clusterOne := clusterData.(map[string]interface{})


			properties := clusterOne["properties"].(map[string]interface{})
			cluster := properties["cluster"].(string)

			labels := properties["labels"].(map[string]interface{})
			region := labels["topology_kubernetes_io_region"].(string)

			window := clusterOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)


			cpuCost := clusterOne["cpuCost"].(float64)
			gpuCost := clusterOne["gpuCost"].(float64)
			ramCost := clusterOne["ramCost"].(float64)
			pvCost  := clusterOne["pvCost"].(float64)
			networkCost := clusterOne["networkCost"].(float64)
			loadBalancerCost := clusterOne["loadBalancerCost"].(float64)
			sharedCost := clusterOne["sharedCost"].(float64)

			totalCost := clusterOne["totalCost"].(float64)
			
			cpuEfficiency := clusterOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := clusterOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := clusterOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100

			record := []interface{}{cluster, region, windowStart, windowEnd ,cpuCost,gpuCost,ramCost,pvCost,networkCost,loadBalancerCost,sharedCost,totalCost,cpuEfficiency,ramEfficiency,totalEfficiency }
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row)
				if err := f.SetCellValue(sheetName, cell, val); err != nil {
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

	InfoLogger.Println("Cluster data successfully written")
}
