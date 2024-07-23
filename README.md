# Kubecost Allocation API Fetcher

This Go script fetches data from the Kubecost Allocation API to retrieve efficiency metrics.

## Prerequisites

- Go (Golang) installed on your machine
- Kubecost Dashboard API endpoint

## Setup

1. Clone the repository to your local machine.
2. Navigate into the project directory.
3. Open the `main.go` file.
4. Replace the placeholder Kubecost dashboard endpoint with your actual endpoint in the `main.go` file.

## Running the Code


To run the script without building using go run:

Note : You should be inside the project directory.

```sh
go run .
```


## Building and Running the Executable

To build the executable:

Note : You should be inside the project directory.

```sh
go build .
```

Then, run the built executable:
```sh
./kubecost-efficiency-fetcher
```
