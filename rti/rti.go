package rti

import (
    "github.com/dop251/goja"
    "go.k6.io/k6/js/modules"
    rtiGo "github.com/rticommunity/rticonnextdds-connector-go"
    "log"
    "encoding/json"
	"strconv"
	"time"
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

    input, _ := r.connector.GetInput("MySubscriber::MyReader")
    if input == nil {
	return "Failed to get input"
    }
    r.connector.Wait(-1)
    input.Take()
    numOfSamples, _ := input.Samples.GetLength()
    for i := 0; i<numOfSamples; i++ {
	valid, _ := input.Infos.IsValid(i)
	if valid {
		data, err := input.Samples.GetJSON(i)
		if err != nil {
			return err.Error()
		} else {
			return string(data)
		}
	}
    }

    return "No data available"
}

// GetRealTimeFracturedData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeFracturedData(messageLength int, isDurableOrReliable bool) string {
    if r.connector == nil {
	return "RTI Connector not initialized"
    }

    input, _ := r.connector.GetInput("MySubscriber::MyReader")
    if input == nil {
	return "Failed to get input"
    }
	bytesRecieved := 0
	var data []byte
	var receivedByte byte
	var err error
	for {
		r.connector.Wait(-1)
		input.Take()
		if isDurableOrReliable {
			receivedByte, err = input.Samples.GetByte(bytesRecieved, "b")
		} else {
			receivedByte, err = input.Samples.GetByte(0, "b")
		}
		bytesRecieved++
		if err != nil {
			return err.Error()
		}
		data = append(data, []byte{receivedByte}...)
		if bytesRecieved == messageLength {
			break
		}
	}

    return string(data)
}

// WriteRealTimeData writes data to the DataWriter.
func (r *RTIModule) WriteRealTimeData(jsonData string) string {
    if r.connector == nil {
        return "RTI Connector not initialized"
    }

    output, _ := r.connector.GetOutput("MyPublisher::MyWriter")
    if output == nil {
        return "Failed to get output"
    }

    var result map[string]interface{}
    marshalErr := json.Unmarshal([]byte(jsonData), &result)
    if marshalErr != nil {
	return "Failed to UnMarshal data: " + marshalErr.Error()
    }
    data, _ := json.Marshal(result)
	
    output.Instance.SetJSON(data)
    err := output.Write()
    if err != nil {
        return "Failed to write data: " + err.Error()
    }
	
    byteCount := len(data)
    return strconv.Itoa(byteCount)
}

// WriteRealTimeDataByRate writes data to the DataWriter by rate.
func (r *RTIModule) WriteRealTimeDataByRate(jsonData string, rate int, size int) string {
    if r.connector == nil {
        return "RTI Connector not initialized"
    }

    output, _ := r.connector.GetOutput("MyPublisher::MyWriter")
    if output == nil {
        return "Failed to get output"
    }

    var result map[string]interface{}
    marshalErr := json.Unmarshal([]byte(jsonData), &result)
    if marshalErr != nil {
		return "Failed to UnMarshal data: " + marshalErr.Error()
    }
    data, _ := json.Marshal(result)
	
	for i := 0; i<len(data); i+=size*rate {
		for j := 0; j<size; j++ {
			if i + j > len(data) {
				return "All Data Has Been Written Successfully"
			}
			output.Instance.SetByte("b", data[i+j])
			err := output.Write()
			if err != nil {
				return "Failed to write data: " + err.Error()
			}
			time.Sleep(time.Duration(rate/size)*time.Second)
		}
	}
	
    return "All Data Has Been Written Successfully"
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

func (r *RTIModule) XGetRealTimeFracturedData(call goja.FunctionCall) goja.Value {
    vm := goja.New()
	messageLength := call.Argument(0).ToInteger()
	isDurableOrReliable := call.Argument(0).ToBoolean()
    result := r.GetRealTimeFracturedData(int(messageLength), isDurableOrReliable)
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

func (r *RTIModule) XWriteRealTimeData(call goja.FunctionCall) goja.Value {
    vm := goja.New()
    jsonData := call.Argument(0).String()
    res, _ := vm.RunString(r.WriteRealTimeData(jsonData))
    return res
}

func (r *RTIModule) XWriteRealTimeDataByRate(call goja.FunctionCall) goja.Value {
    vm := goja.New()
    jsonData := call.Argument(0).String()
	rate := call.Argument(1).ToInteger()
	size := call.Argument(2).ToInteger()
    res, _ := vm.RunString(r.WriteRealTimeDataByRate(jsonData, int(rate), int(size)))
    return res
}
