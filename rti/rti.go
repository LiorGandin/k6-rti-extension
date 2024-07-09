package rti

import (
    "github.com/dop251/goja"
    "go.k6.io/k6/js/modules"
    rtiGo "github.com/rticommunity/rticonnextdds-connector-go"
    "log"
    "encoding/json"
    "strconv"
    "time"
    "sync"
)

// RTIModule is the main structure for the RTI module.
type RTIModule struct {
    connector *rtiGo.Connector
    muWriters sync.Mutex
    muReaders sync.Mutex
    wgWriters sync.WaitGroup
    wgReaders sync.WaitGroup
}

// Init initializes the RTI module.
func (r *RTIModule) Init(configFilePath, configName string, numGoroutinesWriters int, numGoroutinesReaders int) {
    var err error
    r.connector, err = rtiGo.NewConnector(configName, configFilePath)
	
    // Adding the number of goroutines to the WaitGroup
    r.wgWriters.Add(numGoroutinesWriters)
    r.wgReaders.Add(numGoroutinesReaders)
	
    if err != nil {
        log.Fatalf("Failed to create RTI Connector: %v", err)
    }
}

// GetRealTimeData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeData() string {
    r.muReaders.Lock()
    if r.connector == nil {
        defer r.muReaders.Unlock()
	return "RTI Connector not initialized"
    }

    input, _ := r.connector.GetInput("MySubscriber::MyReader")
    if input == nil {
	defer r.muReaders.Unlock()
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
		    defer r.muReaders.Unlock()
		    return err.Error()
		} else {
		    defer r.muReaders.Unlock()
		    return string(data)
		}
	}
    }
    defer r.muReaders.Unlock()
    return "No data available"
}

// GetRealTimeFracturedData is an example function that retrieves real-time data.
func (r *RTIModule) GetRealTimeFracturedData(messageLength int, isDurableOrReliable bool) []byte {
    r.muReaders.Lock()
	if r.connector == nil {
	    defer r.muReaders.Unlock()
	    return []byte("RTI Connector not initialized")
	}
	
    	input, _ := r.connector.GetInput("MySubscriber::MyReader")
	if input == nil {
	    defer r.muReaders.Unlock()
	    return []byte("Failed to get input")
	}
	bytesRecieved := 0
	var data []byte
	var receivedByte byte
	var err error
	r.connector.Wait(-1)
	input.Take()
	numOfSamples, _ := input.Samples.GetLength()
	for i := 0; i < numOfSamples; i++ {
		valid, _ := input.Infos.IsValid(i)
		if valid {
			if isDurableOrReliable {
				receivedByte, err = input.Samples.GetByte(bytesRecieved, "b")
			} else {
				receivedByte, err = input.Samples.GetByte(0, "b")
			}
			bytesRecieved++
			if err != nil {
			    defer r.muReaders.Unlock()
			    return []byte(err.Error())
			}
			data = append(data, []byte{receivedByte}...)
		}
		if bytesRecieved == messageLength {
		    defer r.muReaders.Unlock()
		    return data
		}
	}
	defer r.muReaders.Unlock()
	return []byte("Was unable to get data in it's entirety")
}

// WriteRealTimeData writes data to the DataWriter.
func (r *RTIModule) WriteRealTimeData(jsonData string) string {
    r.muWriters.Lock()
    if r.connector == nil {
	defer r.muWriters.Unlock()
        return "RTI Connector not initialized"
    }

    output, _ := r.connector.GetOutput("MyPublisher::MyWriter")
    if output == nil {
	defer r.muWriters.Unlock()
        return "Failed to get output"
    }

    var result map[string]interface{}
    marshalErr := json.Unmarshal([]byte(jsonData), &result)
    if marshalErr != nil {
	defer r.muWriters.Unlock()
	return "Failed to UnMarshal data: " + marshalErr.Error()
    }
    data, _ := json.Marshal(result)
	
    output.Instance.SetJSON(data)
    err := output.Write()
    if err != nil {
	defer r.muWriters.Unlock()
        return "Failed to write data: " + err.Error()
    }
    defer r.muWriters.Unlock()
    byteCount := len(data)
    return strconv.Itoa(byteCount)
}

// WriteRealTimeDataByRate writes data to the DataWriter by rate.
func (r *RTIModule) WriteRealTimeDataByRate(jsonData string, rate int, size int) string {
    r.muWriters.Lock()
    if r.connector == nil {
	defer r.muWriters.Unlock()
        return "RTI Connector not initialized"
    }

    output, _ := r.connector.GetOutput("MyPublisher::MyWriter")
    if output == nil {
	defer r.muWriters.Unlock()
        return "Failed to get output"
    }

    var result map[string]interface{}
    marshalErr := json.Unmarshal([]byte(jsonData), &result)
    if marshalErr != nil {
	defer r.muWriters.Unlock()
	return "Failed to UnMarshal data: " + marshalErr.Error()
    }
    data, _ := json.Marshal(result)
	
    for i := 0; i<len(data); i+=size*rate {
	for j := 0; j<size; j++ {
	    if i + j > len(data) {
		defer r.muWriters.Unlock()
		return "All Data Has Been Written Successfully"
	    }
	    output.Instance.SetByte("b", data[i+j])
	    err := output.Write()
	    if err != nil {
		defer r.muWriters.Unlock()
		return "Failed to write data: " + err.Error()
	    }
	    time.Sleep(time.Duration(rate/size)*time.Second)
	}
     }
    defer r.muWriters.Unlock()
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
    r.wgReaders.Done()
    r.wgReaders.Wait()
    res, _ := vm.RunString(result)
    return res
}

func (r *RTIModule) XGetRealTimeFracturedData(call goja.FunctionCall) goja.Value {
    vm := goja.New()
    messageLength := call.Argument(0).ToInteger()
    isDurableOrReliable := call.Argument(0).ToBoolean()
    result := string(r.GetRealTimeFracturedData(int(messageLength), isDurableOrReliable))
    r.wgReaders.Done()
    r.wgReaders.Wait()
    res, _ := vm.RunString(result)
    return res
}

// Export Init function to JavaScript runtime
func (r *RTIModule) XInit(call goja.FunctionCall) goja.Value {
    configFilePath := call.Argument(0).String()
    configName := call.Argument(1).String()
    numGoroutinesWriters := call.Argument(2).ToInteger()
    numGoroutinesReaders := call.Argument(3).ToInteger()
    r.Init(configFilePath, configName, int(numGoroutinesWriters), int(numGoroutinesReaders))
    return nil
}

func (r *RTIModule) XWriteRealTimeData(call goja.FunctionCall) goja.Value {
    vm := goja.New()
    jsonData := call.Argument(0).String()
    res, _ := vm.RunString(r.WriteRealTimeData(jsonData))
    r.wgWriters.Done()
    r.wgWriters.Wait()
    return res
}

func (r *RTIModule) XWriteRealTimeDataByRate(call goja.FunctionCall) goja.Value {
    vm := goja.New()
    jsonData := call.Argument(0).String()
    rate := call.Argument(1).ToInteger()
    size := call.Argument(2).ToInteger()
    res, _ := vm.RunString(r.WriteRealTimeDataByRate(jsonData, int(rate), int(size)))
    r.wgWriters.Done()
    r.wgWriters.Wait()
    return res
}
