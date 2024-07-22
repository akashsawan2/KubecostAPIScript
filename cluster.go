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
		fmt.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "cluster")
	q.Set("idle", "false")
	q.Set("accumulate", "true")
	u.RawQuery = q.Encode()

	
	newURL := u.String()

	
	resp, err := http.Get(newURL)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	fmt.Println("Code:", result["code"])

	data := result["data"].([]interface{})

	
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("Error opening Excel file:", err)
		return
	}

	
	header := []string{"Name", "Cluster", "Region", "Window Start", "Window End","Total Cost" ,"Total Efficiency"}
	sheetName := "Cluster"
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) // A1, B1, C1, etc.
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			fmt.Println("Error writing Excel header:", err)
			return
		}
	}

	
	row := 2
	for _, element := range data {
		clusterMap := element.(map[string]interface{})

		
		for _, clusterData := range clusterMap {
			clusterOne := clusterData.(map[string]interface{})

			name := clusterOne["name"].(string)

			properties := clusterOne["properties"].(map[string]interface{})
			cluster := properties["cluster"].(string)

			labels := properties["labels"].(map[string]interface{})
			region := labels["topology_kubernetes_io_region"].(string)

			window := clusterOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			var totalCost float64
			if cost,ok := clusterOne["totalCost"].(float64); ok {
				totalCost = cost
			} else {
				fmt.Println("Total cost is not a float64")
				totalCost = 0 
			}


			var totalEfficiency float64
			if efficiency, ok := clusterOne["totalEfficiency"].(float64); ok {
				totalEfficiency = efficiency * 100
			} else {
				fmt.Println("Total Efficiency is not a float64")
				totalEfficiency = 0
			}

			record := []interface{}{name, cluster, region, windowStart, windowEnd , fmt.Sprintf("%.2f", totalCost) ,fmt.Sprintf("%.2f", totalEfficiency) }
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row)
				if err := f.SetCellValue(sheetName, cell, val); err != nil {
					fmt.Println("Error writing record to Excel:", err)
					return
				}
			}
			row++
		}
	}

	
	if err := f.SaveAs(filePath); err != nil {
		fmt.Println("Error saving Excel file:", err)
		return
	}

	fmt.Println("Cluster data successfully written to Excel file")
}
