# Environment Setup (Ubuntu)

Open the terminal window and then type the following snap command to install the latest Go lang:

`sudo snap install go --classic`

This will install Go programming language compiler, linker, and stdlib. You will see confirmation as follows:

go 1.18 from Michael Hudson-Doyle (mwhudson) installed

# How to install proxy node

git clone git@github.com:Noves-Inc/geth-erigon-converter.git

cd geth-erigon-converter/proxy_server
go get .
go run .

# Test

From a new command line window, use curl to make a request to your running web service.

$ curl http://localhost:8085/eth_getTransactionByHash \
 --include \
 --header "Content-Type: application/json" \
 --request "POST" \
 --data '{"method":"eth_getTransactionByHash","params":["0xb1fac2cb5074a4eda8296faebe3b5a3c10b48947dd9a738b2fdf859be0e1fbaf"],"id":1,"jsonrpc":"2.0"}'

More sample curl commands are defined at curl-commands.txt
