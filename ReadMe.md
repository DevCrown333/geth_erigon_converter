# Environment Setup (Ubuntu)

Open the terminal window and then type the following snap command to install the latest Go lang:
sudo snap install go --classic

This will install Go programming language compiler, linker, and stdlib. You will see confirmation as follows:

go 1.18 from Michael Hudson-Doyle (mwhudson) installed

# How to install proxy node

## Create a folder for your code
To begin, create a project for the code you’ll write.

Open a command prompt and change to your home directory.

On Linux or Mac:

$ cd
On Windows:

C:\> cd %HOMEPATH%
Using the command prompt, create a directory for your code called proxy_node.

$ mkdir proxy_node
$ cd proxy_node
Create a module in which you can manage dependencies.

Run the go mod init command, giving it the path of the module your code will be in.

$ go mod init example/proxy_node
go: creating new go.mod: module example/proxy_node
This command creates a go.mod file in which dependencies you add will be listed for tracking. For more about naming a module with a module path, see Managing dependencies.

## Run the code
Begin tracking the Gin module as a dependency.

At the command line, use go get to add the github.com/gin-gonic/gin module as a dependency for your module. Use a dot argument to mean “get dependencies for code in the current directory.”

$ go get .
go get: added github.com/gin-gonic/gin v1.7.2
Go resolved and downloaded this dependency to satisfy the import declaration you added in the previous step.

From the command line in the directory containing main.go, run the code. Use a dot argument to mean “run code in the current directory.”

$ go run .
Once the code is running, you have a running HTTP server to which you can send requests.

# Test

From a new command line window, use curl to make a request to your running web service.

$ curl http://localhost:8080/eth_getTransactionByHash \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"method":"eth_getTransactionByHash","params":["0xb1fac2cb5074a4eda8296faebe3b5a3c10b48947dd9a738b2fdf859be0e1fbaf"],"id":1,"jsonrpc":"2.0"}'

More sample curl commands are defined at curl-commands.txt