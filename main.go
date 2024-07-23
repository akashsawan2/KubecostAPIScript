package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	inputURL := "http://a9b82b2ca8f5442f4bb118af0d6901fa-1974666064.ap-south-1.elb.amazonaws.com:9090"
	filePath := "output.xlsx"

	
	f := excelize.NewFile() 

	
	defaultSheetName := "Sheet1"   
	newSheetName := "Cluster"
	if err := f.SetSheetName(defaultSheetName, newSheetName); err != nil {
		fmt.Println("Error renaming default sheet:", err)
		return
	}

	
	if err := f.SaveAs(filePath); err != nil {     
		fmt.Println("Error creating Excel file:", err)
		return
	}



	FetchAndWriteClusterData(inputURL, filePath)
	
	FetchAndWriteNodeData(inputURL, filePath)

	FetchAndWritePodData(inputURL, filePath)

	FetchAndWriteNamespaceData(inputURL, filePath)

	FetchAndWriteServiceData(inputURL , filePath)

	FetchAndWriteDeploymentData(inputURL, filePath)

	FetchAndWriteControllerData(inputURL, filePath)

	FetchAndWriteControllerKindData(inputURL, filePath)

}
