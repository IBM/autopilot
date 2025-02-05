package utils

// All GPU tests pass
var GPUHealthPassLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/gpuhealth": "PASS"
			}
		}
	}
`

// At least one GPU test fails. No info about the severity of the failure
var GPUHealthWarnLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/gpuhealth": "WARN"
			}
		}
	}
`

var GPUHealthEmptyLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/gpuhealth": ""
				}
		}
	}
`

var GPUHealthTestingLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/gpuhealth": "TESTING"
				}
		}
	}
`

// Some health check failed. Can be any health check
var NodeHealthWarnLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/nodehealth": "WARN"
			}
		}
	}
`

var NodeHealthEmptyLabel string = `
	{
		"metadata": {
			"labels": {
				"autopilot.ibm.com/nodehealth": ""
			}
		}
	}
`
