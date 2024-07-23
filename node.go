package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWriteNodeData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation" 
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "node")
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

	InfoLogger.Println("Status Code for Node:", result["code"])

	data := result["data"].([]interface{})


	f, err := excelize.OpenFile(filePath)
	if err != nil {
		ErrorLogger.Println("Error opening file:", err)
		return
	}


	if _, err := f.NewSheet("Node"); err != nil {
		ErrorLogger.Println("Error creating 'Pod' sheet:", err)
		return
	}


	header := []string{"Name", "Node", "Region", "Window Start", "Window End","Total Cost","Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) // A1, B1, C1, etc.
		if err := f.SetCellValue("Node", cell, h); err != nil {
			ErrorLogger.Println("Error writing header:", err)
			return
		}
	}


	row := 2
	for _, element := range data {
		nodeMap := element.(map[string]interface{})


		for _, nodeData := range nodeMap {
			nodeOne := nodeData.(map[string]interface{})

			name := nodeOne["name"].(string)

			properties := nodeOne["properties"].(map[string]interface{})
			node := properties["node"].(string)
			
			var region string
			if name != "__idle__"{
				labels := properties["labels"].(map[string]interface{})
			
				region = labels["topology_kubernetes_io_region"].(string)
			}
			

			window := nodeOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)


			var totalCost float64
			if cost, ok := nodeOne["totalCost"].(float64); ok {
				totalCost = cost
			} else {
				ErrorLogger.Println("Error fetching Cost")
				totalCost = 0
			}

			var totalEfficiency float64
			if efficiency, ok := nodeOne["totalEfficiency"].(float64); ok {
				totalEfficiency = efficiency * 100
			} else {
				ErrorLogger.Println("Error fetching efficiency")
				totalEfficiency = 0
			}

			record := []interface{}{name, node, region, windowStart, windowEnd,fmt.Sprintf("%.2f", totalCost) ,fmt.Sprintf("%.2f", totalEfficiency)}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) // A2, B2, C2, etc.
				if err := f.SetCellValue("Node", cell, val); err != nil {
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

	InfoLogger.Println("Node data successfully written")
}
