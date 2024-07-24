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

	header := []string{"Node", "Region", "Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency","Ram Efficiency" ,"Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1)
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
			if name != "__idle__" {
				labels := properties["labels"].(map[string]interface{})

				region = labels["topology_kubernetes_io_region"].(string)
			}

			window := nodeOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)


			cpuCost := nodeOne["cpuCost"].(float64)
			gpuCost := nodeOne["gpuCost"].(float64)
			ramCost := nodeOne["ramCost"].(float64)
			pvCost  := nodeOne["pvCost"].(float64)
			networkCost := nodeOne["networkCost"].(float64)
			loadBalancerCost := nodeOne["loadBalancerCost"].(float64)
			sharedCost := nodeOne["sharedCost"].(float64)

			totalCost := nodeOne["totalCost"].(float64)

			cpuEfficiency := nodeOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := nodeOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := nodeOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100

			record := []interface{}{node, region, windowStart, windowEnd,cpuCost,gpuCost,ramCost,pvCost,networkCost,loadBalancerCost,sharedCost,totalCost,cpuEfficiency,ramEfficiency,totalEfficiency}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row)
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
