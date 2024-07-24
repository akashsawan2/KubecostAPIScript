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


func init() {
    InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func FetchAndWriteDeploymentData(inputURL string, filePath string) {

	u, err := url.Parse(inputURL)
	if err != nil {
		ErrorLogger.Println("Error parsing URL:", err)
		return
	}


	u.Path = "/model/allocation"
	q := u.Query()
	q.Set("window", "24h")
	q.Set("aggregate", "deployment")
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

	InfoLogger.Println("Status Code for Deployment:", result["code"])

	data := result["data"].([]interface{})

	today := time.Now().Format("02January2006")
	fmt.Println(today)

	fileName := "Deployment-" + today + ".csv"

	csvfile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("Failed to open the file",err)
	}

	defer csvfile.Close()



	header := []string{"Deployment", "Region", "Namespace" ,"Window Start", "Window End","Cpu Cost","Gpu Cost","Ram Cost","PV Cost","Network Cost","LoadBalancer Cost","Shared Cost","Total Cost" ,"Cpu Efficiency","Ram Efficiency","Total Efficiency"}

	writer := csv.NewWriter(csvfile)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		log.Fatalln("Fatal to write headers in the csv file",err)
	}



	for _, element := range data {
		deploymentMap := element.(map[string]interface{})


		for _, DeploymentData := range deploymentMap {
			deploymentOne := DeploymentData.(map[string]interface{})

			name := deploymentOne["name"].(string)
			if name =="__unallocated__" {
				continue
			}

			properties := deploymentOne["properties"].(map[string]interface{})

			labels := properties["labels"].(map[string]interface{})
			region := labels["topology_kubernetes_io_region"].(string)

			namespace_deployment, ok := labels["kubernetes_io_metadata_name"].(string)
			if !ok {
				namespace_deployment = ""
			}			

			window := deploymentOne["window"].(map[string]interface{})
			windowStart := window["start"].(string)
			windowEnd := window["end"].(string)

			cpuCost := deploymentOne["cpuCost"].(float64)
			gpuCost := deploymentOne["gpuCost"].(float64)
			ramCost := deploymentOne["ramCost"].(float64)
			pvCost  := deploymentOne["pvCost"].(float64)
			networkCost := deploymentOne["networkCost"].(float64)
			loadBalancerCost := deploymentOne["loadBalancerCost"].(float64)
			sharedCost := deploymentOne["sharedCost"].(float64)

			totalCost := deploymentOne["totalCost"].(float64)

			cpuEfficiency := deploymentOne["cpuEfficiency"].(float64)
			cpuEfficiency = cpuEfficiency * 100

			ramEfficiency := deploymentOne["ramEfficiency"].(float64)
			ramEfficiency = ramEfficiency * 100

			totalEfficiency := deploymentOne["totalEfficiency"].(float64)
			totalEfficiency = totalEfficiency * 100

			record := []string{name, region, namespace_deployment , windowStart, windowEnd,fmt.Sprintf("%f",cpuCost),fmt.Sprintf("%f",gpuCost),fmt.Sprintf("%f",ramCost),fmt.Sprintf("%f",pvCost),fmt.Sprintf("%f",networkCost),fmt.Sprintf("%f",loadBalancerCost),fmt.Sprintf("%f",sharedCost),fmt.Sprintf("%f",totalCost),fmt.Sprintf("%f",cpuEfficiency),fmt.Sprintf("%f",ramEfficiency),fmt.Sprintf("%f",totalEfficiency)}
			if err := writer.Write(record); err != nil{
				fmt.Println("Error writing record")
			}

		}
	}

	InfoLogger.Println("Deployment Data successfully written")
}
