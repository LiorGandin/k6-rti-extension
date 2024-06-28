package rti

import (
    "github.com/dop251/goja"
    "go.k6.io/k6/js/modules"
    rtiGo "github.com/rticommunity/rticonnextdds-connector-go"
    "log"
)

// RTIModule is the main structure for the RTI module.
type RTIModule struct {
    connector *rtiGo.Connector
}

// Init initializes the RTI module.
func (r *RTIModule) Init(configFilePath, configName string) {
    var err error
    r.connector, err = rtiGo.NewConnector(configName, configFilePath)
    if err != nil {
        log.Fatalf("Failed to create RTI Connector: %v", err)
    }
}

// GetRealTimeData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeData() string {
    if r.connector == nil {
	return "RTI Connector not initialized"
    }

    input, _ := r.connector.GetInput("MySubscriber::MySquareReader")
    if input == nil {
	return "Failed to get input"
    }

    input.Take()
    numOfSamples, _ := input.Samples.GetLength()
    for i := 0; i<numOfSamples; i++ {
	data, _ := input.Samples.GetString(i, "color")
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
    vm := goja.New()
    result := r.GetRealTimeData()
    res, _ := vm.RunString(result)
    return res
}

// Export Init function to JavaScript runtime
func (r *RTIModule) XInit(call goja.FunctionCall) goja.Value {
    configFilePath := call.Argument(0).String()
    configName := call.Argument(1).String()
    r.Init(configFilePath, configName)
    return nil
}
