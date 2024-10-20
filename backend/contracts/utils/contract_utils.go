package utils

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CallMethod(client *ethclient.Client, contractABI abi.ABI, contractAddress common.Address, method string, args ...interface{}) ([]interface{}, error) {
	m, exist := contractABI.Methods[method]
	if !exist {
		return nil, fmt.Errorf("方法 %s 不存在", method)
	}

	data, err := contractABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("打包方法 %s 参数失败: %w", method, err)
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用方法 %s 失败: %w", method, err)
	}

	return m.Outputs.Unpack(result)
}
