package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/xuri/excelize/v2"
)


func FetchAndWriteControllerKindData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "controllerKind")
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

	InfoLogger.Println("Status Code for ControllerKind :", result["code"])

	data := result["data"].([]interface{})


	f, err := excelize.OpenFile(filePath)
	if err != nil {
		ErrorLogger.Println("Error opening file:", err)
		return
	}


	if _, err := f.NewSheet("controllerKind"); err != nil {
		ErrorLogger.Println("Error creating 'controllerKind' sheet:", err)
		return
	}


	header := []string{"ControllerKind", "Region" ,"Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency","Ram Efficiency","Total Efficiency"}
	for i, h := range header {
		cell := fmt.Sprintf("%s%d", string('A'+i), 1) 
		if err := f.SetCellValue("controllerKind", cell, h); err != nil {
			ErrorLogger.Println("Error writing header:", err)
			return
		}
	}


	row := 2
	for _, element := range data {
		controllerKindMap := element.(map[string]interface{})


		for _, controllerKindData := range controllerKindMap {
			controllerKindOne := controllerKindData.(map[string]interface{})

			name := controllerKindOne["name"].(string)


			properties := controllerKindOne["properties"].(map[string]interface{})

			var region string
			if name != "__idle__"{
				labels := properties["labels"].(map[string]interface{})
			
				region = labels["topology_kubernetes_io_region"].(string)
			}


			window := controllerKindOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			cpuCost := controllerKindOne["cpuCost"].(float64)
			gpuCost := controllerKindOne["gpuCost"].(float64)
			ramCost := controllerKindOne["ramCost"].(float64)
			pvCost  := controllerKindOne["pvCost"].(float64)
			networkCost := controllerKindOne["networkCost"].(float64)
			loadBalancerCost := controllerKindOne["loadBalancerCost"].(float64)
			sharedCost := controllerKindOne["sharedCost"].(float64)

			totalCost := controllerKindOne["totalCost"].(float64)

			cpuEfficiency := controllerKindOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := controllerKindOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := controllerKindOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100


			record := []interface{}{name, region , windowStart, windowEnd, cpuCost,gpuCost,ramCost,pvCost,networkCost,loadBalancerCost,sharedCost,totalCost,cpuEfficiency,ramEfficiency,totalEfficiency}
			for i, val := range record {
				cell := fmt.Sprintf("%s%d", string('A'+i), row) 
				if err := f.SetCellValue("controllerKind", cell, val); err != nil {
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

	InfoLogger.Println("controllerKind Data successfully written")
}
