package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWriteNamespaceData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	
	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "namespace")
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

	
	if _, err := f.NewSheet("Namespace"); err != nil {
		fmt.Println("Error creating 'Namespace' sheet:", err)
		return
	}

	
	header := []string{"Name", "Namespace", "Region", "Window Start", "Window End","Total Cost","Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) // A1, B1, C1, etc.
		if err := f.SetCellValue("Namespace", cell, h); err != nil {
			fmt.Println("Error writing Excel header:", err)
			return
		}
	}

	
	row := 2
	for _, element := range data {
		namespaceMap := element.(map[string]interface{})

		
		for _, namespaceData := range namespaceMap {
			namespaceOne := namespaceData.(map[string]interface{})

			name := namespaceOne["name"].(string)


			properties := namespaceOne["properties"].(map[string]interface{})
			namespace := properties["namespace"].(string)

			var region string
			if name != "__idle__"{
				labels := properties["labels"].(map[string]interface{})
			
				region = labels["topology_kubernetes_io_region"].(string)
			}

			window := namespaceOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			var totalCost float64
			if cost, ok := namespaceOne["totalCost"].(float64); ok {
				totalCost = cost
			} else {
				fmt.Println("Error fetching Cost")
			}


			var totalEfficiency float64
			if efficiency, ok := namespaceOne["totalEfficiency"].(float64); ok {
				totalEfficiency = efficiency * 100
			} else {
				fmt.Println("Error fetching Efficiency")
			}

			record := []interface{}{name, namespace, region, windowStart, windowEnd,fmt.Sprintf("%.2f", totalCost) ,fmt.Sprintf("%.2f", totalEfficiency)}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) 
				if err := f.SetCellValue("Namespace", cell, val); err != nil {
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

	fmt.Println("Namespace successfully written to Excel file")
}
