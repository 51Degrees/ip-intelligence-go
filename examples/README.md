# 51Degrees Ip Intelligence API Examples

![51Degrees](https://51degrees.com/img/logo.png?utm_source=github&utm_medium=repository&utm_campaign=c_open_source&utm_content=readme_main "Data rewards the curious")
**IP Intelligence Go Examples**

## Introduction

This directory contains usage examples of the [ip-intelligence-go](https://github.com/51degrees/ip-intelligence-go) module.

## Pre-requisites

To run these examples you will need a data file and example evidence. To fetch these assets follow the instructions in
the [ip-intelligence-data](https://github.com/51Degrees/ip-intelligence-data) repo and put in the root of this repository.

### Software

The dependency library `ip-intelligence-go` uses CGO, so make sure `CGO_ENABLED=1` - this is the default, unless you override it.

### Windows

If you are on Windows, make sure that:

- The path to the `MinGW-x64` `bin` folder is included in the `PATH`. By default, the path should be
  `C:\msys64\ucrt64\bin`
- Go environment variable `CGO_ENABLED` is set to `1`

```
go env -w CGO_ENABLED=1
```

## Running the examples

- All examples under `ipi_onpremise` directory are console program examples and are run using `go run`.

Below is a table that describes the examples:

| Example                                                          | Description                                                                                                                               |
|------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| ipi_onpremise/getting_started/getting_sarted.go                  | An example showing how to initialize the IPI engine, minimum required parameters, calling the engine and printing the result              |
| ipi_onpremise/offline_processing/offline_processing.go           | Example showing how to get values from the engine in weighted value format; writing the obtained values to a yaml file                    |
| ipi_onpremise/performance/performance.go                         | A benchmarking example to measure the speed of data processing in single and multi-threaded modes                                         |
| ipi_onpremise/reload_from_file/reload_from_file.go               | An example that demonstrates how a data file can be reloaded while serving IP Intelligence requests                                       |
| ipi_onpremise/update_polling_interval/update_polling_interval.go | An example doing periodic polling for the updated data file                                                                               |

## Run examples

```
go run ./examples/example_dir
```
For further details of how to run each example, please read more in the comment section located at the top of each example file.

To provide a different path to a data file or evidence file use environment variables, e.g.
```bash
DATA_FILE=../51Degrees-EnterpriseIpiV41.ipi go run ./examples/getting_started
DATA_FILE=../51Degrees-EnterpriseIpiV41.ipi EVIDENCE_YAML=../20000_ipi_evidence_records.yml go run ./examples/offline_processing
DATA_FILE=../51Degrees-EnterpriseIpiV41.ipi EVIDENCE_YAML=../20000_ipi_evidence_records.yml go run ./examples/update_polling_interval
```
