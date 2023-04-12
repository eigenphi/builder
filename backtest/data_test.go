package backtest

import (
	"fmt"
	"testing"
)

func TestBuilderPayloadAttributesFromFile(t *testing.T) {
	attr, err := BuilderPayloadAttributesFromFile("../test_txs/payload_attr.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(attr)
}

func TestPendingTxsFromFile(t *testing.T) {
	addrTxs, err := PendingTxsFromFile("../test_txs/txs")
	if err != nil {
		t.Error(err)
	}
	count := 0
	for addr, txs := range addrTxs {
		for _, tx := range txs {
			count += 1
			fmt.Println(addr.String(), tx.Hash().Hex(), tx)
		}
	}
	fmt.Println("total", count, "txs")
}
