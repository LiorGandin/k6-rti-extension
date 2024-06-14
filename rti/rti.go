package rti

import (
    "github.com/dop251/goja"
    "go.k6.io/k6/js/modules"
    "github.com/rticommunity/rticonnextdds-connector-go/pkg/connector"
    "log"
)

// RTIModule is the main structure for the RTI module.
type RTIModule struct {
    connector *connector.Connector
}

// Init initializes the RTI module.
func (r *RTIModule) Init(configFilePath, configName string) {
    var err error
    r.connector, err = connector.NewConnector(configName, configFilePath)
    if err != nil {
        log.Fatalf("Failed to create RTI Connector: %v", err)
    }
}

// GetRealTimeData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeData() string {
    if r.connector == nil {
        return "RTI Connector not initialized"
    }

    // Assuming you have a DataReader for some topic
    input := r.connector.GetInput("MySubscriber::MyReader")
    if input == nil {
        return "Failed to get input"
    }

    input.Take()
    samples, _ := input.Samples.ValidDataIterator()

    for sample := range samples {
        data, _ := sample.GetJSON()
        return string(data)
    }

    return "No data available"
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
