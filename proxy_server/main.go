package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	// "os"
	"bytes"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const QuickNodeURL = "https://black-dark-meadow.quiknode.pro/1a15b493e4c3bc9e6aab158ea100180b05a1944a/"

type GethBlockTraceData struct {
	Jsonrpc string
	Id      string
	Result  []BlockData
}

type GethTransactionTraceData struct {
	Jsonrpc string
	Id      string
	Result  Block
}

type BlockData struct {
	Result Block
}

type Block struct {
	From    string
	Gas     string
	GasUsed string
	To      string
	Input   string
	Output  string
	Calls   []Block
	Value   string
	Type    string
}

type ErigonTraceData struct {
	Jsonrpc string            `json:"jsonrpc"`
	Id      int               `json:"id"`
	Result  []ErigonBlockData `json:"result"`
}

type ErigonBlockData struct {
	Action              ErigonAction      `json:"action"`
	BlockHash           string            `json:"blockHash"`
	BlockNumber         int               `json:"blockNumber"`
	Result              ErigonBlockResult `json:"result"`
	Subtraces           int               `json:"subtraces"`
	TraceAddress        []int             `json:"traceAddress"`
	TransactionHash     string            `json:"transactionHash"`
	TransactionPosition int               `json:"transactionPosition"`
	Type                string            `json:"type"`
}

type ErigonBlockResult struct {
	GasUsed string `json:"gasUsed"`
	Output  string `json:"output"`
}

type ErigonAction struct {
	From     string `json:"from"`
	CallType string `json:"callType"`
	Gas      string `json:"gas"`
	Input    string `json:"input"`
	To       string `json:"to"`
	Value    string `json:"value"`
}

type EthGetLogParam struct {
	FromBlock string `json:"fromBlock,omitempty"`
	ToBlock   string `json:"toBlock,omitempty"`
	Address   string `json:"address,omitempty"`
	Topics    string `json:"topics,omitempty"`
	BlockHash string `json:"blockhash,omitempty"`
}

type EthGetTransactionParam struct {
	Hash string `json:"hash"`
}

type EthGetBlockByNumberParam struct {
	BlockNumber string `json:"blockNumber"`
	DetailFlag  bool   `json:"detailFlag,omitempty"`
}

type EthGetBlockByHashParam struct {
	BlockHash  string `json:"hash"`
	DetailFlag bool   `json:"detailFlag,omitempty"`
}

type EthGetTransactionReceiptParam struct {
	Hash string `json:"hash"`
}

type ApiLogParam struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type ApiTransactionParam struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type ApiBlockParam struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type ApiReceiptParam struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

func organizeData(gethTraceBlockData Block, index int, callStack []int) ErigonBlockData {
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
	oneBlock.BlockNumber = 0
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
		traceAddress := append(callStack, j)
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
	// ApiParam := ApiLogParam {
	//     Method: "eth_getLogs",
	//     Params: []EthGetLogParam{param},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func eth_getTransactionReceipt(c *gin.Context) {
	var param ApiReceiptParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	// Construct API parameter.
	// ApiParam := ApiReceiptParam {
	//     Method: "eth_getTransactionReceipt",
	//     Params: []string{param.Hash},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func eth_getTransactionByHash(c *gin.Context) {
	var param ApiTransactionParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	// Construct API parameter.
	// ApiParam := ApiReceiptParam {
	//     Method: "eth_getTransactionReceipt",
	//     Params: []string{param.Hash},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func eth_getBlockByNumber(c *gin.Context) {
	var param ApiBlockParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	// Construct API parameter.
	// ApiParam := ApiReceiptParam {
	//     Method: "eth_getTransactionReceipt",
	//     Params: []string{param.Hash},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func trace_transaction(c *gin.Context) {
	var param ApiTransactionParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	param.Method = "debug_traceTransaction"
	param.Params = append(param.Params, map[string]string{
		"tracer": "callTracer",
	})
	// Construct API parameter.
	// ApiParam := ApiTransactionParam {
	//     Method: "eth_getTransactionByHash",
	//     Params: []string{param.Hash},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var gethTraceData GethTransactionTraceData

	json.Unmarshal(body, &gethTraceData)

	convertTransactionTraceData(gethTraceData)

	c.IndentedJSON(http.StatusOK, erigonTraceData)
}

func trace_block(c *gin.Context) {
	var param ApiBlockParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	param.Method = "debug_traceBlockByNumber"
	param.Params = append(param.Params, map[string]string{
		"tracer": "callTracer",
	})
	// Construct API parameter.
	// ApiParam := ApiBlockParam {
	//     Method: "eth_getBlockByNumber",
	//     Params: []interface{}{
	//         "0xc5043f", false,
	//     },
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var gethTraceData GethBlockTraceData

	json.Unmarshal(body, &gethTraceData)

	convertBlockTraceData(gethTraceData)

	c.IndentedJSON(http.StatusOK, erigonTraceData)
}

func eth_getBlockByHash(c *gin.Context) {
	var param ApiBlockParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	// Construct API parameter.
	// ApiParam := ApiReceiptParam {
	//     Method: "eth_getTransactionReceipt",
	//     Params: []string{param.Hash},
	//     Id : 1,
	//     Jsonrpc: "2.0",
	// }

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func web3_clientVersion(c *gin.Context) {
	var param ApiBlockParam

	// Call BindJSON to bind the received JSON to
	// param.
	if err := c.BindJSON(&param); err != nil {
		return
	}

	json_data, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(QuickNodeURL, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, string(body))
}

func main() {
	router := gin.Default()
	fmt.Println("Proxy server started!")
	router.POST("/eth_getLogs", eth_getLogs)
	router.POST("/eth_getTransactionByHash", eth_getTransactionByHash)
	router.POST("/eth_getBlockByNumber", eth_getBlockByNumber)
	router.POST("/eth_getBlockByHash", eth_getBlockByHash)
	router.POST("/eth_getTransactionReceipt", eth_getTransactionReceipt)
	router.POST("/trace_block", trace_block)
	router.POST("/trace_transaction", trace_transaction)
	router.POST("/web3_clientVersion", web3_clientVersion)

	router.Run("localhost:8085")
}
