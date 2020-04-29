package bankhandler

import (
	"dat520/lab5/bank"
	"dat520/lab5/multipaxos"
	"fmt"
)

type BankHandler struct {
	adu                int
	bufferDecidedValue []multipaxos.DecidedValue
	bankAccounts       map[int]bank.Account
	responseChanOut    chan<- multipaxos.Response
}

func NewBankHandler(responseChan chan<-multipaxos.Response) *BankHandler {
	return &BankHandler{
		adu: -1,
		bufferDecidedValue: []multipaxos.DecidedValue,
		bankAccounts: map[int]bank.Account{},
		responseChanOut: responseChan,
	}
}

func (bh *BankHandler) Start() {

}

func (bh *BankHandler) DeliverWhat() {
	//Some Chan to receive bank info?
}

func (bh *BankHandler) DeliverMoreWhat() {
	//Some Chan to send bank info?
}

//HandleDecidedValue from learner
func (bh *BankHandler) HandleDecidedValue(dVal multipaxos.DecidedValue, adu int) {
	fmt.Printf("Main: handleDecidedValue %v", dVal)
	if int(dVal.SlotID) <= bh.adu {
		fmt.Printf("Main: handleDecidedValue - SlotID (%d) is smaller than 'All Decided Upto' (adu = %d)", int(dVal.SlotID), adu)
		return
	}
	if int(dVal.SlotID) > bh.adu+1 {
		bufferDecidedValue := append(bufferDecidedValue, dVal)
		return
	}
	if dVal.Value.Noop == false {
		accountID := dVal.Value.AccountNum
		if _, ok := bh.bankAccounts[accountID]; ok {

		}
	}
}
