package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
)
var (
    InfoLogger  *log.Logger
    ErrorLogger *log.Logger
)

func init() {
    InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}


func main() {
	const kubecostEndpoint = "http://a9b82b2ca8f5442f4bb118af0d6901fa-1974666064.ap-south-1.elb.amazonaws.com:9090"
	filePath := "efficiency.xlsx"

	
	f := excelize.NewFile() 

	
	defaultSheetName := "Sheet1"   
	newSheetName := "Cluster"
	if err := f.SetSheetName(defaultSheetName, newSheetName); err != nil {
		fmt.Println("Error renaming default sheet:", err)
		return
	}

	
	if err := f.SaveAs(filePath); err != nil {     
		fmt.Println("Error creating file:", err)
		return
	}



	FetchAndWriteClusterData(kubecostEndpoint, filePath)
	
	FetchAndWriteNodeData(kubecostEndpoint, filePath)

	FetchAndWritePodData(kubecostEndpoint, filePath)

	FetchAndWriteNamespaceData(kubecostEndpoint, filePath)

	FetchAndWriteServiceData(kubecostEndpoint , filePath)

	FetchAndWriteDeploymentData(kubecostEndpoint, filePath)

	FetchAndWriteControllerData(kubecostEndpoint, filePath)

	FetchAndWriteControllerKindData(kubecostEndpoint, filePath)

}
