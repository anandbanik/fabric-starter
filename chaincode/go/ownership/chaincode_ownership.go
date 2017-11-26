
package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("OwnershipChaincode")

type OwnershipChaincode struct {
}

func (t *OwnershipChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *OwnershipChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "move" {
		// Make payment of X units from A to B
		return t.move(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"unknown function name"}
}

func (t *OwnershipChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 3 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	}

	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *OwnershipChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		//return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
		return shim.Success(nil)
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(OwnershipChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
