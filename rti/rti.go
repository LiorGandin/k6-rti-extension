package rti

import (
    "github.com/dop251/goja"
    "go.k6.io/k6/js/modules"
    rti "github.com/rticommunity/rticonnextdds-connector-go"
    "log"
)

// RTIModule is the main structure for the RTI module.
type RTIModule struct {
    connector *connector.Connector
}

// Init initializes the RTI module.
func (r *RTIModule) Init(configFilePath, configName string) {
    var err error
    r.connector, err = rti.NewConnector(configName, configFilePath)
    if err != nil {
        log.Fatalf("Failed to create RTI Connector: %v", err)
    }
}

// GetRealTimeData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeData() string {
    // Replace with actual RTI logic using r.connector
    // This is a placeholder example.
    return "Real-time data"
}

// Register the RTI module
func init() {
    modules.Register("k6/x/rti", new(RTIModule))
}

// Export RTI functions to JavaScript runtime
func (r *RTIModule) XGetRealTimeData(call goja.FunctionCall) goja.Value {
    rt := call.Runtime
    result := r.GetRealTimeData()
    return rt.ToValue(result)
}
