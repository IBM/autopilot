package remapped

import (
	"testing"
)

// TestParseRemappedRowsOutput tests the parseRemappedRowsOutput function
func TestParseRemappedRowsOutput(t *testing.T) {
	// Test case 1: No remapped rows (Pending : No)
	output := `==============NVSMI LOG==============

Timestamp                                 : Mon Jul 26 14:30:00 2021
Driver Model
    Current                              : N/A
    Pending                              : N/A

Attached GPUs                            : 1
=======================
    GPU 00000000:00:00.0
        Product Name                     : NVIDIA A100-SXM4-40GB
        Product Brand                    : NVIDIA
        Display Mode                     : Disabled
        Display Active                   : Disabled
        Persistence Mode                 : Enabled
        MIG Mode
            Current                      : N/A
            Pending                      : N/A
        Accounting Mode                  : Disabled
        Accounting Mode Buffer Size      : 4000
        Driver Model
            Current                      : N/A
            Pending                      : N/A
        Serial Number                    : 1234567890
        GPU UUID                         : GPU-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
        Minor Number                     : 0
        VBIOS Version                    : 90.00.00.00.00
        MultiGPU Board                   : No
        Board ID                         : 0x0
        GPU Part Number                  : 12345-678-A0
        Inforom Version
            Image Version                : G001.0000.00.01
            OEM Object                   : 1.1
            ECC Object                   : 4.1
            Power Management Object      : N/A
        GPU Operation Mode
            Current                      : N/A
            Pending                      : N/A
        GPU Virtualization Mode
            Virtualization Mode          : Pass-Through
        IBMNVLINK Version                : N/A
        IBMNVLINK Link                   : N/A
        IBMNVLINK Status                 : N/A
        IBMNVLINK Error                  : N/A
        IBMNVLINK Warning                : N/A
        Fan Speed                        : 30 %
        Performance State                : P0
        Clocks Throttle Reasons
            None                         : Active
        FB Memory Usage
            Total                        : 40536 MiB
            Reserved                     : 624 MiB
            Used                         : 0 MiB
            Free                         : 39912 MiB
        BAR1 Memory Usage
            Total                        : 65536 MiB
            Used                         : 0 MiB
            Free                         : 65536 MiB
        Compute Mode                     : Default
        Utilization
            Gpu                          : 0 %
            Memory                       : 0 %
            Encoder                      : 0 %
            Decoder                      : 0 %
        Encoder Stats
            Active Sessions              : 0
            Average FPS                  : 0
            Average Latency              : 0
        FBC Stats
            Sessions Count               : 0
            Average FPS                  : 0
            Average Latency              : 0
        Ecc Mode
            Current                      : Enabled
            Pending                      : Enabled
        ECC Errors
            Volatile
                Single Bit
                    Device Memory        : N/A
                    Register File       : N/A
                    L1 Cache             : N/A
                    L2 Cache             : N/A
                    Texture Memory       : N/A
                    Texture Shared       : N/A
                    CBU                  : N/A
                Double Bit
                    Device Memory        : N/A
                    Register File        : N/A
                    L1 Cache             : N/A
                    L2 Cache             : N/A
                    Texture Memory       : N/A
                    Texture Shared       : N/A
                    CBU                  : N/A
            Aggregate
                Single Bit
                    Device Memory        : 0
                    Register File        : 0
                    L1 Cache             : 0
                    L2 Cache             : 0
                    Texture Memory       : 0
                    Texture Shared       : 0
                    CBU                  : 0
                Double Bit
                    Device Memory        : 0
                    Register File        : 0
                    L1 Cache             : 0
                    L2 Cache             : 0
                    Texture Memory       : 0
                    Texture Shared       : 0
                    CBU                  : 0
        Retired Pages
            Single Bit ECC               : 0
            Double Bit ECC               : 0
            Pending                      : No
        Remapped Rows
            Correctable Error            : 0
            Uncorrectable Error          : 0
            Pending                      : No
            Remapping Failure Occurred   : No
        Temperature
            GPU Current Temp             : 30 C
            GPU Shutdown Temp            : 90 C
            GPU Slowdown Temp           : 87 C
            GPU Max Operating Temp       : 83 C
            Memory Current Temp          : 30 C
            Memory Max Operating Temp    : 95 C
        Power Readings
            Power Management             : Supported
            Power Draw                   : 45.25 W
            Power Limit                  : 400.00 W
            Default Power Limit          : 400.00 W
            Enforced Power Limit         : 400.00 W
            Min Power Limit              : 150.00 W
            Max Power Limit              : 400.00 W
        Clocks
            Graphics                     : 1200 MHz
            SM                           : 1200 MHz
            Memory                       : 1215 MHz
            Video                        : 1200 MHz
        Applications Clocks
            Graphics                     : 1200 MHz
            Memory                       : 1215 MHz
        Default Applications Clocks
            Graphics                     : 1200 MHz
            Memory                       : 1215 MHz
        Max Clocks
            Graphics                     : 1200 MHz
            SM                           : 1200 MHz
            Memory                       : 1215 MHz
            Video                        : 1200 MHz
        Clock Policy
            Auto Boost                   : On
            Auto Boost Default          : On
        Processes                        : None

`

	hasRemappedRows, err := parseRemappedRowsOutput(output)
	if err != nil {
		t.Errorf("parseRemappedRowsOutput returned error: %v", err)
	}

	if hasRemappedRows != false {
		t.Errorf("Expected hasRemappedRows to be false, got %t", hasRemappedRows)
	}

	// Test case 2: Has remapped rows (Pending : Yes)
	output = `==============NVSMI LOG==============

Timestamp                                 : Mon Jul 26 14:30:00 2021
Driver Model
    Current                              : N/A
    Pending                              : N/A

Attached GPUs                            : 1
=======================
    GPU 00000000:00:00.0
        Product Name                     : NVIDIA A100-SXM4-40GB
        Product Brand                    : NVIDIA
        Display Mode                     : Disabled
        Display Active                   : Disabled
        Persistence Mode                 : Enabled
        MIG Mode
            Current                      : N/A
            Pending                      : N/A
        Accounting Mode                  : Disabled
        Accounting Mode Buffer Size      : 4000
        Driver Model
            Current                      : N/A
            Pending                      : N/A
        Serial Number                    : 1234567890
        GPU UUID                         : GPU-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
        Minor Number                     : 0
        VBIOS Version                    : 90.00.00.00.00
        MultiGPU Board                   : No
        Board ID                         : 0x0
        GPU Part Number                  : 12345-678-A0
        Inforom Version
            Image Version                : G001.0000.00.01
            OEM Object                   : 1.1
            ECC Object                   : 4.1
            Power Management Object      : N/A
        GPU Operation Mode
            Current                      : N/A
            Pending                      : N/A
        GPU Virtualization Mode
            Virtualization Mode          : Pass-Through
        IBMNVLINK Version                : N/A
        IBMNVLINK Link                   : N/A
        IBMNVLINK Status                 : N/A
        IBMNVLINK Error                  : N/A
        IBMNVLINK Warning                : N/A
        Fan Speed                        : 30 %
        Performance State                : P0
        Clocks Throttle Reasons
            None                         : Active
        FB Memory Usage
            Total                        : 40536 MiB
            Reserved                     : 624 MiB
            Used                         : 0 MiB
            Free                         : 39912 MiB
        BAR1 Memory Usage
            Total                        : 65536 MiB
            Used                         : 0 MiB
            Free                         : 65536 MiB
        Compute Mode                     : Default
        Utilization
            Gpu                          : 0 %
            Memory                       : 0 %
            Encoder                      : 0 %
            Decoder                      : 0 %
        Encoder Stats
            Active Sessions              : 0
            Average FPS                  : 0
            Average Latency              : 0
        FBC Stats
            Sessions Count               : 0
            Average FPS                  : 0
            Average Latency              : 0
        Ecc Mode
            Current                      : Enabled
            Pending                      : Enabled
        ECC Errors
            Volatile
                Single Bit
                    Device Memory        : N/A
                    Register File       : N/A
                    L1 Cache             : N/A
                    L2 Cache             : N/A
                    Texture Memory       : N/A
                    Texture Shared       : N/A
                    CBU                  : N/A
                Double Bit
                    Device Memory        : N/A
                    Register File        : N/A
                    L1 Cache             : N/A
                    L2 Cache             : N/A
                    Texture Memory       : N/A
                    Texture Shared       : N/A
                    CBU                  : N/A
            Aggregate
                Single Bit
                    Device Memory        : 0
                    Register File        : 0
                    L1 Cache             : 0
                    L2 Cache             : 0
                    Texture Memory       : 0
                    Texture Shared       : 0
                    CBU                  : 0
                Double Bit
                    Device Memory        : 0
                    Register File        : 0
                    L1 Cache             : 0
                    L2 Cache             : 0
                    Texture Memory       : 0
                    Texture Shared       : 0
                    CBU                  : 0
        Retired Pages
            Single Bit ECC               : 0
            Double Bit ECC               : 0
            Pending                      : No
        Remapped Rows
            Correctable Error            : 0
            Uncorrectable Error          : 0
            Pending                      : Yes
            Remapping Failure Occurred   : No
        Temperature
            GPU Current Temp             : 30 C
            GPU Shutdown Temp            : 90 C
            GPU Slowdown Temp           : 87 C
            GPU Max Operating Temp       : 83 C
            Memory Current Temp          : 30 C
            Memory Max Operating Temp    : 95 C
        Power Readings
            Power Management             : Supported
            Power Draw                   : 45.25 W
            Power Limit                  : 400.00 W
            Default Power Limit          : 400.00 W
            Enforced Power Limit         : 400.00 W
            Min Power Limit              : 150.00 W
            Max Power Limit              : 400.00 W
        Clocks
            Graphics                     : 1200 MHz
            SM                           : 1200 MHz
            Memory                       : 1215 MHz
            Video                        : 1200 MHz
        Applications Clocks
            Graphics                     : 1200 MHz
            Memory                       : 1215 MHz
        Default Applications Clocks
            Graphics                     : 1200 MHz
            Memory                       : 1215 MHz
        Max Clocks
            Graphics                     : 1200 MHz
            SM                           : 1200 MHz
            Memory                       : 1215 MHz
            Video                        : 1200 MHz
        Clock Policy
            Auto Boost                   : On
            Auto Boost Default          : On
        Processes                        : None

`

	hasRemappedRows, err = parseRemappedRowsOutput(output)
	if err != nil {
		t.Errorf("parseRemappedRowsOutput returned error: %v", err)
	}

	if hasRemappedRows != true {
		t.Errorf("Expected hasRemappedRows to be true, got %t", hasRemappedRows)
	}
}