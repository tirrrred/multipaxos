package bankhandler

import (
	"dat520/lab5/bank"
	"dat520/lab5/multipaxos"
	"fmt"
)

//BankHandler struct
type BankHandler struct {
	adu                int
	bufferDecidedValue []multipaxos.DecidedValue
	bankAccounts       map[int]bank.Account
	responseChanOut    chan<- multipaxos.Response
	proposer           *multipaxos.Proposer
}

//NewBankHandler inits a new bankHandler
func NewBankHandler(responseChan chan<- multipaxos.Response, proposer *multipaxos.Proposer) *BankHandler {
	return &BankHandler{
		adu:                -1,
		bufferDecidedValue: []multipaxos.DecidedValue{},
		bankAccounts:       map[int]bank.Account{},
		responseChanOut:    responseChan,
		proposer:           proposer,
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
func (bh *BankHandler) HandleDecidedValue(dVal multipaxos.DecidedValue) {
	fmt.Printf("Main: handleDecidedValue %v", dVal)
	if int(dVal.SlotID) <= bh.adu {
		fmt.Printf("Main: handleDecidedValue - SlotID (%d) is smaller than 'All Decided Upto' (adu = %d)", int(dVal.SlotID), bh.adu)
		return
	}
	if int(dVal.SlotID) > bh.adu+1 {
		bh.bufferDecidedValue = append(bh.bufferDecidedValue, dVal)
		return
	}
	if dVal.Value.Noop == false {
		accountID := dVal.Value.AccountNum
		if _, ok := bh.bankAccounts[accountID]; ok != true {
			bh.bankAccounts[accountID] = bank.Account{
				Number:  accountID,
				Balance: 0,
			}
		}
		bankAccount := bh.bankAccounts[accountID]
		transRes := bankAccount.Process(dVal.Value.Txn)
		response := multipaxos.Response{
			ClientID:  dVal.Value.ClientID,
			ClientSeq: dVal.Value.ClientSeq,
			TxnRes:    transRes,
		}
		bh.responseChanOut <- response
	}
	bh.adu++
	bh.proposer.IncrementAllDecidedUpTo()
	for i, buffdecidedVal := range bh.bufferDecidedValue {
		if int(buffdecidedVal.SlotID) == bh.adu+1 {
			bh.bufferDecidedValue = append(bh.bufferDecidedValue[:i], bh.bufferDecidedValue[i+1:]...)
			bh.HandleDecidedValue(buffdecidedVal)
		}
	}
}

//TestBankHandler testing bankhandler for some dVals
func (bh *BankHandler) TestBankHandler(dVals []multipaxos.DecidedValue) {

}
