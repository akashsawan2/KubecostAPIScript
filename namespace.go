package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"log"
	"os"
	"encoding/csv"
)


func FetchAndWriteNamespaceData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
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

	InfoLogger.Println("Status Code for Namespace :", result["code"])

	data := result["data"].([]interface{})

	today := time.Now().Format("02January2006")
	fmt.Println(today)

	fileName := "Namespace-" + today + ".csv"

	csvfile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("Failed to open the file",err)
	}

	defer csvfile.Close()

	
	header := []string{"Namespace", "Region", "Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost","Cpu Efficiency","Ram Efficiency","Total Efficiency"}
	
	writer := csv.NewWriter(csvfile)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		log.Fatalln("Fatal to write headers in the csv file",err)
	}

	

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

			cpuCost := namespaceOne["cpuCost"].(float64)
			gpuCost := namespaceOne["gpuCost"].(float64)
			ramCost := namespaceOne["ramCost"].(float64)
			pvCost  := namespaceOne["pvCost"].(float64)
			networkCost := namespaceOne["networkCost"].(float64)
			loadBalancerCost := namespaceOne["loadBalancerCost"].(float64)
			sharedCost := namespaceOne["sharedCost"].(float64)

			totalCost := namespaceOne["totalCost"].(float64)

			cpuEfficiency := namespaceOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := namespaceOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := namespaceOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100

			record := []string{namespace, region, windowStart, windowEnd,fmt.Sprintf("%f",cpuCost),fmt.Sprintf("%f",gpuCost),fmt.Sprintf("%f",ramCost),fmt.Sprintf("%f",pvCost),fmt.Sprintf("%f",networkCost),fmt.Sprintf("%f",loadBalancerCost),fmt.Sprintf("%f",sharedCost),fmt.Sprintf("%f",totalCost),fmt.Sprintf("%f",cpuEfficiency),fmt.Sprintf("%f",ramEfficiency),fmt.Sprintf("%f",totalEfficiency)}
			if err := writer.Write(record); err != nil{
				fmt.Println("Error writing record")
			}
		}
	}




	InfoLogger.Println("Namespace successfully written")
}
