package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "bytes"
    "net/http"
    "github.com/gin-gonic/gin"
)

const QuickNodeURL = "https://black-dark-meadow.quiknode.pro/1a15b493e4c3bc9e6aab158ea100180b05a1944a/"


type GethBlockTraceData struct {
    Jsonrpc	string
    Id	string
    Result	[]BlockData
}

type GethTransactionTraceData struct {
    Jsonrpc	string
    Id	string
    Result  Block
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
    Id  int	`json:"id"`
    Result      []ErigonBlockData	`json:"result"`
}

type ErigonBlockData struct {
    Action    ErigonAction	`json:"action"`
    BlockHash    string	`json:"blockHash"`
    BlockNumber    int	`json:"blockNumber"`
    Result ErigonBlockResult	`json:"result"`
    Subtraces    int	`json:"subtraces"`
    TraceAddress    []int	`json:"traceAddress"`
    TransactionHash    string	`json:"transactionHash"`
    TransactionPosition    int	`json:"transactionPosition"`
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

type EthGetLogParam struct {
    FromBlock string    `json:"fromBlock"`
    ToBlock string  `json:"toBlock"`
    Address string  `json:"address"`
    Topics string   `json:"topics"`
    BlockHash string    `json:"blockhash"`
}

type EthGetTransactionParam struct {
    Hash string `json:"hash"`
}

type EthGetBlockByNumberParam struct {
    BlockNumber string  `json:"blockNumber"`
    DetailFlag bool `json:"detailFlag"`
}

type EthGetTransactionReceiptParam struct {
    Hash string `json:"hash"`
}

type ApiLogParam struct {
    Method string   `json:"method"`
    Params EthGetLogParam   `json:"params"`
    Id int  `json:"id"`
    Jsonrpc string  `json:"jsonrpc"`
}

type ApiTransactionParam struct {
    Method string   `json:"method"`
    Params EthGetTransactionParam   `json:"params"`
    Id int  `json:"id"`
    Jsonrpc string  `json:"jsonrpc"`
}

type ApiBlockParam struct {
    Method string   `json:"method"`
    Params EthGetBlockByNumberParam `json:"params"`
    Id int  `json:"id"`
    Jsonrpc string  `json:"jsonrpc"`
}

type ApiReceiptParam struct {
    Method string   `json:"method"`
    Params EthGetTransactionReceiptParam    `json:"params"`
    Id int  `json:"id"`
    Jsonrpc string  `json:"jsonrpc"`
}

func organizeData(gethTraceBlockData Block, index int, callStack []int) ErigonBlockData{
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

    subCallCount := len(gethTraceBlockData.Calls)

    oneBlock.BlockHash = ""
    oneBlock.BlockNumber = 16636490
    oneBlock.Subtraces = subCallCount
    oneBlock.TransactionHash = ""
    oneBlock.TransactionPosition = index

    // if callStack != -1 {
    oneBlock.TraceAddress = callStack
    // }
    
    return oneBlock
}

func handleSubCalls(gethTraceBlockData Block, index int, callStack []int) {
    for j := 0; j < len(gethTraceBlockData.Calls); j++ {
        traceAddress := append(callStack, j);
        callBlock := organizeData(gethTraceBlockData.Calls[j], index, traceAddress)
        
        erigonTraceData.Result = append(erigonTraceData.Result, callBlock)
        if len(gethTraceBlockData.Calls[j].Calls) > 0 { 
            handleSubCalls(gethTraceBlockData.Calls[j], index, traceAddress)
        }
    }
}

var erigonTraceData ErigonTraceData

var zeroErigonTraceData = &ErigonTraceData{}
func (a *ErigonTraceData) Reset() {
    *a = *zeroErigonTraceData
}

func convertBlockTraceData(gethTraceData GethBlockTraceData) {
    erigonTraceData.Reset()

    erigonTraceData.Jsonrpc = gethTraceData.Jsonrpc
    erigonTraceData.Id = 0
    initial_subtrace := []int{}

    for i := 0; i < len(gethTraceData.Result); i++ {
        gethTraceBlockData := gethTraceData.Result[i].Result
        oneBlock := organizeData(gethTraceBlockData, i, initial_subtrace)
        erigonTraceData.Result = append(erigonTraceData.Result, oneBlock)
        
        handleSubCalls(gethTraceBlockData, i, initial_subtrace)
    }
}

func convertTransactionTraceData(gethTraceData GethTransactionTraceData) {
    erigonTraceData.Reset()

    erigonTraceData.Jsonrpc = gethTraceData.Jsonrpc
    erigonTraceData.Id = 0
    initial_subtrace := []int{}

    gethTraceBlockData := gethTraceData.Result
    oneBlock := organizeData(gethTraceBlockData, 0, initial_subtrace)
    erigonTraceData.Result = append(erigonTraceData.Result, oneBlock)
    
    handleSubCalls(gethTraceBlockData, 0, initial_subtrace)
}


func eth_getLogs(c *gin.Context) {
    var param ApiLogParam

    // Call BindJSON to bind the received JSON to
    // param.
    if err := c.BindJSON(&param); err != nil {
        return
    }

    // Construct API parameter.
    ApiParam := ApiLogParam {
        Method: "eth_getLogs",
        Params: param,
        Id : 1,
        Jsonrpc: "2.0"
    }

    json_data, err := json.Marshal(ApiParam)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post(QuickNodeURL, "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Fatal(err)
    }

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)

    c.IndentedJSON(http.StatusOK, res)
}

func eth_getTransactionByHash(c *gin.Context) {
    var param ApiTransactionParam

    // Call BindJSON to bind the received JSON to
    // param.
    if err := c.BindJSON(&param); err != nil {
        return
    }

    // Construct API parameter.
    ApiParam := ApiTransactionParam {
        Method: "eth_getTransactionByHash",
        Params: param,
        Id : 1,
        Jsonrpc: "2.0"
    }

    json_data, err := json.Marshal(ApiParam)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post(QuickNodeURL, "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Fatal(err)
    }

    var gethTraceData GethTransactionTraceData

    json.Unmarshal(resp.Body, &gethTraceData)

    convertTransactionTraceData(gethTraceData)

    res, _ := json.MarshalIndent(erigonTraceData, "", "\t")

    c.IndentedJSON(http.StatusOK, res)
}

func eth_getBlockByNumber(c *gin.Context) {
    var param ApiBlockParam

    // Call BindJSON to bind the received JSON to
    // param.
    if err := c.BindJSON(&param); err != nil {
        return
    }

    // Construct API parameter.
    ApiParam := ApiBlockParam {
        Method: "eth_getBlockByNumber",
        Params: param,
        Id : 1,
        Jsonrpc: "2.0"
    }

    json_data, err := json.Marshal(ApiParam)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post(QuickNodeURL, "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Fatal(err)
    }

    var gethTraceData GethBlockTraceData

    json.Unmarshal(resp.Body, &gethTraceData)

    convertBlockTraceData(gethTraceData)

    res, _ := json.MarshalIndent(erigonTraceData, "", "\t")

    c.IndentedJSON(http.StatusOK, res)
}

func eth_getTransactionReceipt(c *gin.Context) {
    var param ApiReceiptParam

    // Call BindJSON to bind the received JSON to
    // param.
    if err := c.BindJSON(&param); err != nil {
        return
    }

    // Construct API parameter.
    ApiParam := ApiReceiptParam {
        Method: "eth_getTransactionReceipt",
        Params: param,
        Id : 1,
        Jsonrpc: "2.0"
    }

    json_data, err := json.Marshal(ApiParam)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post(QuickNodeURL, "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Fatal(err)
    }

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)

    c.IndentedJSON(http.StatusOK, res)
}

func main() {
    router := gin.Default()
    router.POST("/eth_getLogs", eth_getLogs)
    router.POST("/eth_getTransactionByHash", eth_getTransactionByHash)
    router.POST("/eth_getBlockByNumber", eth_getBlockByNumber)
    router.POST("/eth_getTransactionReceipt", eth_getTransactionReceipt)

    router.Run("localhost:8080")
}
