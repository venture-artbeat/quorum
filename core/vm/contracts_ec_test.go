package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestHelloPrecompile_Run ensures that it is possible to run the basic precompile directly.
func TestHelloPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	helloTx := types.NewTransaction(0, common.ECPrecompileHelloContractAddress(), nil, 0, nil, []byte{})
	require.EqualValues(t, *helloTx.To(), common.ECPrecompileHelloContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    helloTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Run the code
	retData, err := (&helloPrecompile{}).Run(evm, []byte{})
	require.Nil(t, retData)
	require.Nil(t, err)
}
