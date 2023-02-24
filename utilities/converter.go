package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)
 

type GethTraceData struct {
    Jsonrpc	string
    Id	string
    Result	[]BlockData
}

type BlockData struct {
    Result Block
}

type Block struct {
    From	string
    Gas	string
    GasUsed	string
    To	string
    Input	string
    Output string
    Calls	[]Block
    Value string
    Type string
}

type ErigonTraceData struct {
    Jsonrpc     string	`json:"jsonrpc"`
    Id  string	`json:"id"`
    Result      []ErigonBlockData	`json:"result"`
}

type ErigonBlockData struct {
    Action    ErigonAction	`json:"action"`
    BlockHash    string	`json:"blockHash"`
    BlockNumber    string	`json:"blockNumber"`
    Result ErigonBlockResult	`json:"result"`
    Subtraces    string	`json:"subtraces"`
    TraceAddress    []string	`json:"traceAddress"`
    TransactionHash    string	`json:"transactionHash"`
    TransactionPosition    string	`json:"transactionPosition"`
    Type    string	`json:"type"`
}

type ErigonBlockResult struct {
    GasUsed    string	`json:"gasUsed"`
    Output    string	`json:"output"`
}

type ErigonAction struct {
    From    string	`json:"from"`
    CallType    string	`json:"callType"`
    Gas    string	`json:"gas"`
    Input    string	`json:"input"`
    To    string	`json:"to"`
    Value    string	`json:"value"`
}

func organizeData(gethTraceBlockData Block) ErigonBlockData{
    var oneBlock ErigonBlockData
    var action ErigonAction
    action.From = gethTraceBlockData.From
    action.CallType = strings.ToLower(gethTraceBlockData.Type)
    action.Gas = gethTraceBlockData.Gas
    action.Input = gethTraceBlockData.Input
    action.To = gethTraceBlockData.To
    action.Value = gethTraceBlockData.Value
    if len(action.Value) == 0 {
	    action.Value = "0x0"
    }
    oneBlock.Action = action
    oneBlock.Result = ErigonBlockResult{gethTraceBlockData.GasUsed, gethTraceBlockData.Output}
    oneBlock.Type = "call"
    if len(oneBlock.Result.Output) == 0 {
	    oneBlock.Result.Output = "0x"
    }

    oneBlock.BlockHash = ""
    oneBlock.BlockNumber = ""
    oneBlock.Subtraces = "0"
    oneBlock.TransactionHash = ""
    oneBlock.TransactionPosition = ""
    
    return oneBlock
}

func handleSubCalls(gethTraceBlockData Block) {
    for j := 0; j < len(gethTraceBlockData.Calls); j++ {
        callBlock := organizeData(gethTraceBlockData.Calls[j])
        erigonTraceData.Result = append(erigonTraceData.Result, callBlock)
        if len(gethTraceBlockData.Calls[j].Calls) > 0 {
            handleSubCalls(gethTraceBlockData.Calls[j])
        }
    }
}

var erigonTraceData ErigonTraceData

func main() {
    // Open our jsonFile
    jsonFile, err := os.Open("geth_output_call_tracer_version.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        fmt.Println(err)
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

    // read our opened xmlFile as a byte array.
    byteValue, _ := ioutil.ReadAll(jsonFile)

    // we initialize our Users array
    var gethTraceData GethTraceData

    // we unmarshal our byteArray which contains our
    // jsonFile's content into 'users' which we defined above
    json.Unmarshal(byteValue, &gethTraceData)
    
    erigonTraceData.Jsonrpc = gethTraceData.Jsonrpc
    erigonTraceData.Id = "0"

    for i := 0; i < len(gethTraceData.Result); i++ {
        gethTraceBlockData := gethTraceData.Result[i].Result
        oneBlock := organizeData(gethTraceBlockData)
        erigonTraceData.Result = append(erigonTraceData.Result, oneBlock)
        
        handleSubCalls(gethTraceBlockData)
    }

    // for i := 0; i < len(erigonTraceData.Result); i++ {
	// fmt.Println("from: " + erigonTraceData.Result[i].Action.From)
    // }
    file, _ := json.MarshalIndent(erigonTraceData, "", " ")

    _ = ioutil.WriteFile("erigon_output1.json", file, 0644)

    fmt.Println("Erigon output is successfully generated.")

}