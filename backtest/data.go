package backtest

import (
	"bufio"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"io"
	"os"
	"strings"
)

func BuilderPayloadAttributesFromFile(
	filePath string) (*types.BuilderPayloadAttributes, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	fileContent, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var attr types.BuilderPayloadAttributes
	err = json.Unmarshal(fileContent, &attr)
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func PendingTxsFromFile(
	filePath string) (map[common.Address]types.Transactions, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(f)

	result := make(map[common.Address]types.Transactions)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		splitLine := strings.SplitN(line, " ", 2)
		if len(splitLine) != 2 {
			continue
		}
		fromString, txString := splitLine[0], splitLine[1]
		from := common.HexToAddress(fromString)
		tx := new(types.Transaction)
		err = tx.UnmarshalJSON([]byte(txString))
		if err != nil {
			return nil, err
		}
		result[from] = append(result[from], tx)
	}

	return result, nil
}
