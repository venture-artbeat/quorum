// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// ECPrecompiledContract is an interface for precompiled contracts to do with extended elliptic curve operations. The
// implementation requires a deterministic gas count based on the size of the input to the contract's `Run` method.
type ECPrecompiledContract interface {
	RequiredGas(input []byte) uint64            // Calculates the gas used in calling the precompiled contract
	Run(evm *EVM, input []byte) ([]byte, error) // Runs the precompiled contract
}

// ECPrecompiledContracts is the default set of pre-compiled contracts that implement extended support for elliptic
// curves.
var ECPrecompiledContracts = map[common.Address]ECPrecompiledContract{
	common.ECPrecompileHelloContractAddress(): &helloPrecompile{},
}

// RunECPrecompiledContract executes and evaluates the output of an elliptic curve precompiled contract.
//
// It returns:
//   - ret: the returned bytes
//   - remainingGas: The gas that remains from the `suppliedGas` after the precompiled contract's gas cost has been
//     deducted.
//   - err: Any error that may have occurred during execution.
func RunECPrecompiledContract(evm *EVM, precompile ECPrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	gasCost := precompile.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	suppliedGas -= gasCost

	output, err := precompile.Run(evm, input)
	return output, suppliedGas, err
}

// helloPrecompile is a basic test precompile that does nothing of interest.
type helloPrecompile struct{}

func (c *helloPrecompile) RequiredGas(_ []byte) uint64 {
	return 0
}

func (c *helloPrecompile) Run(evm *EVM, _ []byte) ([]byte, error) {
	log.Debug("Executing the hello precompile")
	log.Warn("Hello Precompile")

	return nil, nil
}
