package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
	"encoding/pem"
	"crypto/x509"
)

var logger = shim.NewLogger("PaymentChaincode")

type PaymentChaincode struct {
}

func (t *PaymentChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

func (t *PaymentChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "debit" {
		return t.debit(stub, args)
	} else if function == "add" {
		return t.add(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"unknown function name"}
}

func (t *PaymentChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}



func (t *PaymentChaincode) debit(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	} 

	debitAmt := args[0]
	producer := args[1]

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot find creator")
	}

	name, org := getCreator(creatorBytes)
	consumer := name + "@" + org


	consumervalbytes, err := stub.GetState(consumer)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if consumervalbytes == nil {
		return pb.Response{Status:403, Message:"consumer not found"}
	}

	consumerval, _ := strconv.Atoi(string(consumervalbytes))

	X, err := strconv.Atoi(debitAmt)
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}

	if X < consumerval {
		return pb.Response{Status:403, Message:"consumer does not have balance"}
	}
	consumerval = consumerval - X

	logger.Debug("consumerval = %d \n", consumerval)

	// Write the state back to the ledger
	err = stub.PutState(consumer, []byte(strconv.Itoa(consumerval)))
	if err != nil {
		return shim.Error(err.Error())
	}


	stub.SetEvent(producer,[]byte(producer))


	return shim.Success(nil)
}

func (t *PaymentChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	name, org := getCreator(creatorBytes)

	key := name + "@" + org

	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("cannot get state")
	}

	return shim.Success(bytes)
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(PaymentChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
