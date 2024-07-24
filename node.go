package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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


	today := time.Now().Format("02January2006")
	fmt.Println(today)

	fileName := "node-" + today + ".csv"

	csvfile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("Failed to open the file",err)
	}

	defer csvfile.Close()

	headers := []string{"Cluster", "Region","Window Start","Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency","Ram Efficiency","Total Efficiency"}

	writer := csv.NewWriter(csvfile)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		log.Fatalln("Fatal to write headers in the csv file",err)
	}


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


			record := []string{node, region, windowStart, windowEnd,fmt.Sprintf("%f",cpuCost),fmt.Sprintf("%f",gpuCost),fmt.Sprintf("%f",ramCost),fmt.Sprintf("%f",pvCost),fmt.Sprintf("%f",networkCost),fmt.Sprintf("%f",loadBalancerCost),fmt.Sprintf("%f",sharedCost),fmt.Sprintf("%f",totalCost),fmt.Sprintf("%f",cpuEfficiency),fmt.Sprintf("%f",ramEfficiency),fmt.Sprintf("%f",totalEfficiency)}
			if err := writer.Write(record); err != nil{
				fmt.Println("error writing record")
			}
		}
	}


	InfoLogger.Println("Node data successfully written")
}
