package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWritePodData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	
	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "pod")
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

	
	if _, err := f.NewSheet("Pod"); err != nil {
		fmt.Println("Error creating 'Pod' sheet:", err)
		return
	}

	
	header := []string{"Name", "Pod", "Region","Namespace" ,"Window Start", "Window End","Total Cost","Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) // A1, B1, C1, etc.
		if err := f.SetCellValue("Pod", cell, h); err != nil {
			fmt.Println("Error writing Excel header:", err)
			return
		}
	}

	
	row := 2
	for _, element := range data {
		podMap := element.(map[string]interface{})

		
		for _, podData := range podMap {
			podOne := podData.(map[string]interface{})

			name := podOne["name"].(string)
			if name =="__unallocated__" {
				continue
			}

			properties := podOne["properties"].(map[string]interface{})
			pod := properties["pod"].(string)

			var region string
			if name != "__idle__"{
				labels := properties["labels"].(map[string]interface{})
			
				region = labels["topology_kubernetes_io_region"].(string)
			}
			
			namespace_labels := properties["namespaceLabels"].(map[string]interface{})
			namespace_pod := namespace_labels["kubernetes_io_metadata_name"]

			window := podOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			var totalCost float64
			if cost, ok := podOne["totalCost"].(float64); ok {
				totalCost = cost
			} else {
				fmt.Println("Error fetching Cost")
			}


			var totalEfficiency float64
			if efficiency, ok := podOne["totalEfficiency"].(float64); ok {
				totalEfficiency = efficiency * 100
			} else {
				fmt.Println("Error fetching Efficiency")
				totalEfficiency = 0
			}

			record := []interface{}{name, pod, region,namespace_pod ,windowStart, windowEnd,fmt.Sprintf("%.2f", totalCost) ,fmt.Sprintf("%.2f", totalEfficiency)}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) // A2, B2, C2, etc.
				if err := f.SetCellValue("Pod", cell, val); err != nil {
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

	fmt.Println("Pod Data successfully written to Excel file")
}
