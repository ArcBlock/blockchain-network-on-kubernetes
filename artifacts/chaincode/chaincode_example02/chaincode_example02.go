/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.
package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type GameChainCode struct {
}

func (cc *GameChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	var sum string // Sum of asset holdings across accounts. Initially 0
	var sumVal int // Sum of holdings
	var err error
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	// Initialize the chaincode
	sum = args[0]
	sumVal, err = strconv.Atoi(sum)
	if err != nil {
		return shim.Error("Expecting integer value for sum")
	}
	fmt.Printf("sumVal = %d\n", sumVal)
	if sumVal < 100 {
		return shim.Error("init value must bigger than 100")
	}
	// Write the state to the ledger
	err = stub.PutState("total", []byte(strconv.Itoa(sumVal)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *GameChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "query" {
		return cc.query(stub, args)
	} else if function == "lottery" {
		return cc.lottery(stub, args)
	} else if function == "create_user" {
		return cc.create_user(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting query uid,lottery uid token,create_user uid")
}

func (cc *GameChainCode) create_user(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("params error")
	}
	uId := args[0]
	balance, err := cc.getBalance(stub, uId)
	if balance < 0 || err != nil {
		stub.PutState(uId, []byte("1000"))
		return shim.Success([]byte("1000"))
	} else {
		return shim.Error("uid has been registed")
	}
}

func (cc *GameChainCode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 0 {
		t, err := cc.getTotal(stub)
		if err != nil {
			return shim.Error("get total error!")
		}
		if t < 100 {
			return shim.Success([]byte("false"))
		} else {
			return shim.Success([]byte("true"))
		}
	} else {
		t, err := cc.getBalance(stub, args[0])
		if err != nil {
			return shim.Error("get balance error!")
		} else {
			return shim.Success([]byte(strconv.Itoa(t)))
		}
	}

}

func (cc *GameChainCode) lottery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("params error")
	}
	t, err := cc.getTotal(stub)
	if err != nil {
		return shim.Error("get total error!")
	}
	uId := args[0]
	uWager, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("wager error!")
	}
	err = cc.plus(stub, uId, -uWager)
	if err != nil {
		return shim.Error("balance not enough")
	}

	if (uWager - t) == 0 {
		//hit
		err = cc.plus(stub, uId, uWager+t)
		return shim.Success([]byte("you win! game over!"))
	} else if ((t - uWager) > t>>2) && ((t - uWager) < t>>2+t>>3) {
		err = cc.plus(stub, uId, uWager>>1)
		return shim.Success([]byte("congratulation!!!!"))
	} else {
		err = cc.plus(stub, "total", uWager)
		return shim.Success([]byte("Good Luck Next Time!"))
	}
	return shim.Error("bet error")

}

//plus or minus value of key
func (cc *GameChainCode) plus(stub shim.ChaincodeStubInterface, k string, v int) error {
	Avalbytes, err := stub.GetState(k)
	if err != nil {
		return err
	}
	TotalV, err := strconv.Atoi(string(Avalbytes))
	if err != nil {
		return err
	}
	TotalV += v
	if TotalV < 0 {
		return errors.New("balance not enough")
	}
	stub.PutState(k, []byte(strconv.Itoa(TotalV)))
	return nil
}

func (cc *GameChainCode) getBalance(stub shim.ChaincodeStubInterface, k string) (int, error) {
	Avalbytes, err := stub.GetState(k)
	if err != nil {
		return -1, err
	}
	TotalV, err := strconv.Atoi(string(Avalbytes))
	if err != nil {
		return -1, err
	}
	return TotalV, nil
}

func (cc *GameChainCode) getTotal(stub shim.ChaincodeStubInterface) (int, error) {
	return cc.getBalance(stub, "total")
}

func main() {
	err := shim.Start(new(GameChainCode))
	if err != nil {
		fmt.Printf("Error starting game chaincode: %s", err)
	}
}
