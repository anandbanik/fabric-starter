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
	if function == "move" {
		return t.move(stub, args)
	} else if function == "add" {
		return t.add(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}else if function == "credit" {
		return t.credit(stub, args)
	}

	return pb.Response{Status:403, Message:"unknown function name"}
}

func (t *PaymentChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}


func (t *PaymentChaincode) credit(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	
	var valueToAdd, total int
	
	if len(args) != 2 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	}

	funcCall := []byte("query")
	key := []byte(args[0])

	argTocc := [][]byte{funcCall,key}

	response := stub.InvokeChaincode("mycc",argTocc,"gateway-producer")

	payloadBytes := response.GetPayload()

	owner := string(payloadBytes)

	valueToAdd, err := strconv.Atoi(args[1])

	rs, err := stub.GetState(owner)
	if err != nil {
		return shim.Error("Cannot get owner balance")
	}

	strBalance:= string(rs)
	
	balance, err := strconv.Atoi(strBalance)
	if err != nil {
		return shim.Error("Cannot get owner balance")
	}

	total = balance + valueToAdd

	err = stub.PutState(owner, []byte(strconv.Itoa(total)))
	if err != nil {
		return shim.Error("cannot put state")
	}

	return shim.Success(nil)
}
func (t *PaymentChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	}

	A = args[0]
	B = args[1]

	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

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
