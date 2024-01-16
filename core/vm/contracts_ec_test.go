package vm

import (
	"testing"

	"math/big"

	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBw6761G1AddPrecompile_RequiredGas(t *testing.T) {
	input := make([]byte, 0)
	requiredGas := (&bw6761G1AddPrecompile{}).RequiredGas(input)
	require.Equal(t, requiredGas, params.Bw6761G1AddGas)
}

func TestBw6761G1AddPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	g1AddTx := types.NewTransaction(0, common.ECPrecompileBW6761G1AddContractAddress(), nil, 0, nil, []byte{})
	require.Equal(t, *g1AddTx.To(), common.ECPrecompileBW6761G1AddContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    g1AddTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Set up the inputs. They need to be in Montgomery form as Bytes converts back into normal form.
	var x1, y1, x2, y2 fp.Element
	_, _ = x1.SetRandom()
	_, _ = y1.SetRandom()
	_, _ = x2.SetRandom()
	_, _ = y2.SetRandom()
	x1Bytes := x1.Bytes()
	y1Bytes := y1.Bytes()
	x2Bytes := x2.Bytes()
	y2Bytes := y2.Bytes()

	point1Bytes := make([]byte, sizeOfAffinePoint)
	copy(point1Bytes[:sizeOfFieldElement], x1Bytes[:])
	copy(point1Bytes[sizeOfFieldElement:], y1Bytes[:])
	point1, err1 := decodeG1Point(point1Bytes)
	require.NoErrorf(t, err1, "")
	point2Bytes := make([]byte, sizeOfAffinePoint)
	copy(point2Bytes[:sizeOfFieldElement], x2Bytes[:])
	copy(point2Bytes[sizeOfFieldElement:], y2Bytes[:])
	point2, err2 := decodeG1Point(point2Bytes)
	require.NoErrorf(t, err2, "")
	inputBytes := make([]byte, 2*sizeOfAffinePoint)
	copy(inputBytes[:sizeOfFieldElement], x1Bytes[:])
	copy(inputBytes[sizeOfFieldElement:sizeOfFieldElement*2], y1Bytes[:])
	copy(inputBytes[sizeOfFieldElement*2:sizeOfFieldElement*3], x2Bytes[:])
	copy(inputBytes[sizeOfFieldElement*3:], y2Bytes[:])

	// Create the expected output. They also need to be in Montgomery form as Bytes converts back into normal form.
	point1Jac := bw6761.G1Jac{}
	point1Jac.FromAffine(point1)
	point2Jac := bw6761.G1Jac{}
	point2Jac.FromAffine(point2)
	resultJac := point1Jac.AddAssign(&point2Jac)
	resultAffine := bw6761.G1Affine{}
	resultAffine.FromJacobian(resultJac)
	resultXBytes := resultAffine.X.Bytes()
	resultYBytes := resultAffine.Y.Bytes()
	expectedBytes := make([]byte, sizeOfAffinePoint)
	copy(expectedBytes[:sizeOfFieldElement], resultXBytes[:])
	copy(expectedBytes[sizeOfFieldElement:], resultYBytes[:])

	// Run the precompile
	retData, err := (&bw6761G1AddPrecompile{}).Run(evm, inputBytes)
	require.Nil(t, err)

	// Check the returned value is correct
	require.Equal(t, len(retData), len(expectedBytes))
	require.Equal(t, retData, expectedBytes)
}

func TestBw6761G2AddPrecompile_RequiredGas(t *testing.T) {
	input := make([]byte, 0)
	requiredGas := (&bw6761G2AddPrecompile{}).RequiredGas(input)
	require.Equal(t, requiredGas, params.Bw6761G2AddGas)
}

func TestBw6761G2AddPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	g1AddTx := types.NewTransaction(0, common.ECPrecompileBW6761G2AddContractAddress(), nil, 0, nil, []byte{})
	require.Equal(t, *g1AddTx.To(), common.ECPrecompileBW6761G2AddContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    g1AddTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Set up the inputs. They need to be in Montgomery form as Bytes converts back into normal form.
	var x1, y1, x2, y2 fp.Element
	_, _ = x1.SetRandom()
	_, _ = y1.SetRandom()
	_, _ = x2.SetRandom()
	_, _ = y2.SetRandom()
	x1Bytes := x1.Bytes()
	y1Bytes := y1.Bytes()
	x2Bytes := x2.Bytes()
	y2Bytes := y2.Bytes()

	point1Bytes := make([]byte, sizeOfAffinePoint)
	copy(point1Bytes[:sizeOfFieldElement], x1Bytes[:])
	copy(point1Bytes[sizeOfFieldElement:], y1Bytes[:])
	point1, err1 := decodeG2Point(point1Bytes)
	require.NoErrorf(t, err1, "")
	point2Bytes := make([]byte, sizeOfAffinePoint)
	copy(point2Bytes[:sizeOfFieldElement], x2Bytes[:])
	copy(point2Bytes[sizeOfFieldElement:], y2Bytes[:])
	point2, err2 := decodeG2Point(point2Bytes)
	require.NoErrorf(t, err2, "")
	inputBytes := make([]byte, 2*sizeOfAffinePoint)
	copy(inputBytes[:sizeOfFieldElement], x1Bytes[:])
	copy(inputBytes[sizeOfFieldElement:sizeOfFieldElement*2], y1Bytes[:])
	copy(inputBytes[sizeOfFieldElement*2:sizeOfFieldElement*3], x2Bytes[:])
	copy(inputBytes[sizeOfFieldElement*3:], y2Bytes[:])

	// Create the expected output. They also need to be in Montgomery form as Bytes converts back into normal form.
	point1Jac := bw6761.G2Jac{}
	point1Jac.FromAffine(point1)
	point2Jac := bw6761.G2Jac{}
	point2Jac.FromAffine(point2)
	resultJac := point1Jac.AddAssign(&point2Jac)
	resultAffine := bw6761.G2Affine{}
	resultAffine.FromJacobian(resultJac)
	resultXBytes := resultAffine.X.Bytes()
	resultYBytes := resultAffine.Y.Bytes()
	expectedBytes := make([]byte, sizeOfAffinePoint)
	copy(expectedBytes[:sizeOfFieldElement], resultXBytes[:])
	copy(expectedBytes[sizeOfFieldElement:], resultYBytes[:])

	// Run the precompile
	retData, err := (&bw6761G2AddPrecompile{}).Run(evm, inputBytes)
	require.Nil(t, err)

	// Check the returned value is correct
	require.Equal(t, len(retData), len(expectedBytes))
	require.Equal(t, retData, expectedBytes)
}

func TestBw6761G1ScalarMulPrecompile_RequiredGas(t *testing.T) {
	input := make([]byte, 0)
	requiredGas := (&bw6761G1ScalarMulPrecompile{}).RequiredGas(input)
	require.Equal(t, requiredGas, params.Bw6761G1MulGas)
}

func TestBw6761G1ScalarMulPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	g1AddTx := types.NewTransaction(0, common.ECPrecompileBW6761G1ScalarMulContractAddress(), nil, 0, nil, []byte{})
	require.Equal(t, *g1AddTx.To(), common.ECPrecompileBW6761G1ScalarMulContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    g1AddTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Set up the inputs. They need to be in Montgomery form as Bytes converts back into normal form.
	var x, y fp.Element
	_, _ = x.SetRandom()
	_, _ = y.SetRandom()
	xBytes := x.Bytes()
	yBytes := y.Bytes()

	pointBytes := make([]byte, sizeOfAffinePoint)
	copy(pointBytes[:sizeOfFieldElement], xBytes[:])
	copy(pointBytes[sizeOfFieldElement:], yBytes[:])
	point, err1 := decodeG1Point(pointBytes)
	require.NoErrorf(t, err1, "")

	scalar := new(big.Int)
	scalar.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(scalar, big.NewInt(1))

	inputBytes := make([]byte, sizeOfAffinePoint+sizeOfEVMWordBytes)
	copy(inputBytes[:sizeOfFieldElement], xBytes[:])
	copy(inputBytes[sizeOfFieldElement:sizeOfFieldElement*2], yBytes[:])
	copy(inputBytes[sizeOfFieldElement*2:], scalar.Bytes())

	// Create the expected output. They also need to be in Montgomery form as Bytes converts back into normal form.
	expectedResult := point.ScalarMultiplication(point, scalar)
	expectedXBytes := expectedResult.X.Bytes()
	expectedYBytes := expectedResult.Y.Bytes()
	expectedBytes := make([]byte, sizeOfAffinePoint)
	copy(expectedBytes[:sizeOfFieldElement], expectedXBytes[:])
	copy(expectedBytes[sizeOfFieldElement:], expectedYBytes[:])

	// Run the precompile
	retData, err := (&bw6761G1ScalarMulPrecompile{}).Run(evm, inputBytes)
	require.Nil(t, err)

	// Check the returned value is correct
	require.Equal(t, len(retData), len(expectedBytes))
	require.Equal(t, retData, expectedBytes)
}

func TestBw6761G2ScalarMulPrecompile_RequiredGas(t *testing.T) {
	input := make([]byte, 0)
	requiredGas := (&bw6761G2ScalarMulPrecompile{}).RequiredGas(input)
	require.Equal(t, requiredGas, params.Bw6761G2MulGas)
}

func TestBw6761G2ScalarMulPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	g1AddTx := types.NewTransaction(0, common.ECPrecompileBW6761G2ScalarMulContractAddress(), nil, 0, nil, []byte{})
	require.Equal(t, *g1AddTx.To(), common.ECPrecompileBW6761G2ScalarMulContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    g1AddTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Set up the inputs. They need to be in Montgomery form as Bytes converts back into normal form.
	var x, y fp.Element
	_, _ = x.SetRandom()
	_, _ = y.SetRandom()
	xBytes := x.Bytes()
	yBytes := y.Bytes()

	pointBytes := make([]byte, sizeOfAffinePoint)
	copy(pointBytes[:sizeOfFieldElement], xBytes[:])
	copy(pointBytes[sizeOfFieldElement:], yBytes[:])
	point, err1 := decodeG2Point(pointBytes)
	require.NoErrorf(t, err1, "")

	scalar := new(big.Int)
	scalar.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(scalar, big.NewInt(1))

	inputBytes := make([]byte, sizeOfAffinePoint+sizeOfEVMWordBytes)
	copy(inputBytes[:sizeOfFieldElement], xBytes[:])
	copy(inputBytes[sizeOfFieldElement:sizeOfFieldElement*2], yBytes[:])
	copy(inputBytes[sizeOfFieldElement*2:], scalar.Bytes())

	// Create the expected output. They also need to be in Montgomery form as Bytes converts back into normal form.
	expectedResult := point.ScalarMultiplication(point, scalar)
	expectedXBytes := expectedResult.X.Bytes()
	expectedYBytes := expectedResult.Y.Bytes()
	expectedBytes := make([]byte, sizeOfAffinePoint)
	copy(expectedBytes[:sizeOfFieldElement], expectedXBytes[:])
	copy(expectedBytes[sizeOfFieldElement:], expectedYBytes[:])

	// Run the precompile
	retData, err := (&bw6761G2ScalarMulPrecompile{}).Run(evm, inputBytes)
	require.Nil(t, err)

	// Check the returned value is correct
	require.Equal(t, len(retData), len(expectedBytes))
	require.Equal(t, retData, expectedBytes)
}

func TestBw6761PairingCheckPrecompile_RequiredGas(t *testing.T) {
	input := make([]byte, 0)
	requiredGas := (&bw6761PairingCheckPrecompile{}).RequiredGas(input)
	require.Equal(t, requiredGas, params.Bw6761PairingGas)
}

func TestBw6761PairingCheckPrecompile_Run(t *testing.T) {
	// Set up the test controller and make sure it runs cleanup
	controller := gomock.NewController(t)
	defer controller.Finish()

	// Build our test transaction
	g1AddTx := types.NewTransaction(0, common.ECPrecompileBW6761PairingCheckContractAddress(), nil, 0, nil, []byte{})
	require.Equal(t, *g1AddTx.To(), common.ECPrecompileBW6761PairingCheckContractAddress())

	// Set up the EVM state
	publicState := NewMockStateDB(controller)
	depth := 1
	evm := &EVM{
		depth:        depth,
		currentTx:    g1AddTx,
		publicState:  publicState,
		privateState: publicState,
	}

	// Set up the inputs. They need to be in Montgomery form as Bytes converts back into normal form.
	var g1Gen bw6761.G1Jac
	g1Gen.X.SetString("6238772257594679368032145693622812838779005809760824733138787810501188623461307351759238099287535516224314149266511977132140828635950940021790489507611754366317801811090811367945064510304504157188661901055903167026722666149426237")
	g1Gen.Y.SetString("2101735126520897423911504562215834951148127555913367997162789335052900271653517958562461315794228241561913734371411178226936527683203879553093934185950470971848972085321797958124416462268292467002957525517188485984766314758624099")
	g1Gen.Z.SetOne()

	var g2Gen bw6761.G2Jac
	g2Gen.X.SetString("6445332910596979336035888152774071626898886139774101364933948236926875073754470830732273879639675437155036544153105017729592600560631678554299562762294743927912429096636156401171909259073181112518725201388196280039960074422214428")
	g2Gen.Y.SetString("562923658089539719386922163444547387757586534741080263946953401595155211934630598999300396317104182598044793758153214972605680357108252243146746187917218885078195819486220416605630144001533548163105316661692978285266378674355041")
	g2Gen.Z.SetOne()

	g1GenAff := bw6761.G1Affine{}
	g1GenAff.FromJacobian(&g1Gen)

	g2GenAff := bw6761.G2Affine{}
	g2GenAff.FromJacobian(&g2Gen)

	point1Bytes := encodeG1Point(&g1GenAff)
	point2Bytes := encodeG2Point(&g2GenAff)

	inputBytes := make([]byte, 2*sizeOfAffinePoint)
	copy(inputBytes[:sizeOfAffinePoint], point1Bytes[:])
	copy(inputBytes[sizeOfAffinePoint:], point2Bytes[:])

	// Create the expected output. They also need to be in Montgomery form as Bytes converts back into normal form.
	expectedResult, err := bw6761.PairingCheck([]bw6761.G1Affine{g1GenAff}, []bw6761.G2Affine{g2GenAff})
	require.Nil(t, err)
	var expectedResultWord byte
	if expectedResult {
		expectedResultWord = 1
	} else {
		expectedResultWord = 0
	}
	expectedBytes := []byte{expectedResultWord}

	// Run the precompile
	retData, err := (&bw6761PairingCheckPrecompile{}).Run(evm, inputBytes)
	require.Nil(t, err)

	// Check the returned value is correct
	require.Equal(t, len(retData), len(expectedBytes))
	require.Equal(t, retData, expectedBytes)
}

func TestDecodeFieldElementInto(t *testing.T) {
	// Create our input
	var input fp.Element
	_, err1 := input.SetRandom()
	require.NoErrorf(t, err1, "generating input failed")
	inputBytes := input.Bytes()

	// Test decoding
	var element fp.Element
	err2 := decodeFieldElementInto(inputBytes[:], &element)
	require.NoErrorf(t, err2, "decoding failed")
	require.Equal(t, element, input)
}

// TestDecodeEncodeG1Point ensures that we can decode and encode a G1 point correctly.
func TestDecodeEncodeG1Point(t *testing.T) {
	// Create our input
	input := make([]byte, sizeOfAffinePoint)
	var inputX, inputY fp.Element
	_, err1 := inputX.SetRandom()
	require.NoErrorf(t, err1, "generating input failed")
	_, err2 := inputY.SetRandom()
	require.NoErrorf(t, err2, "generating input failed")
	xBytes := inputX.Bytes()
	yBytes := inputY.Bytes()
	copy(input[:sizeOfFieldElement], xBytes[:])
	copy(input[sizeOfFieldElement:], yBytes[:])

	// Test the decoding functionality
	point, decodeErr := decodeG1Point(input)
	require.NoErrorf(t, decodeErr, "decoding the point failed")
	require.Equal(t, point.X, inputX)
	require.Equal(t, point.Y, inputY)

	// Test the encoding functionality
	bytes := encodeG1Point(point)
	require.Equal(t, bytes, input)
}

// TestDecodeEncodeG2Point ensures that we can decode and encode a G2 point correctly.
func TestDecodeEncodeG2Point(t *testing.T) {
	// Create our input
	input := make([]byte, sizeOfAffinePoint)
	var inputX, inputY fp.Element
	_, err1 := inputX.SetRandom()
	require.NoErrorf(t, err1, "generating input failed")
	_, err2 := inputY.SetRandom()
	require.NoErrorf(t, err2, "generating input failed")
	xBytes := inputX.Bytes()
	yBytes := inputY.Bytes()
	copy(input[:sizeOfFieldElement], xBytes[:])
	copy(input[sizeOfFieldElement:], yBytes[:])

	// Test the decoding functionality
	point, decodeErr := decodeG2Point(input)
	require.NoErrorf(t, decodeErr, "decoding the point failed")
	require.Equal(t, point.X, inputX)
	require.Equal(t, point.Y, inputY)

	// Test the encoding functionality
	bytes := encodeG2Point(point)
	require.Equal(t, bytes, input)
}
