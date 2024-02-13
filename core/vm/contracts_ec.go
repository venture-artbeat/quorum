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
	"bytes"
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	fr_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	plonk_bw6761 "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var (
	errBW6InvalidFieldElementLength = errors.New("invalid field element length")
	errBW6InvalidInputLength        = errors.New("invalid input length")
)

var (
	sizeOfFieldElement = 96
	sizeOfAffinePoint  = 2 * sizeOfFieldElement
	sizeOfEVMWordBytes = 32
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
	common.ECPrecompileBW6761G1AddContractAddress():            &bw6761G1AddPrecompile{},
	common.ECPrecompileBW6761G1ScalarMulContractAddress():      &bw6761G1ScalarMulPrecompile{},
	common.ECPrecompileBW6761G2AddContractAddress():            &bw6761G2AddPrecompile{},
	common.ECPrecompileBW6761G2ScalarMulContractAddress():      &bw6761G2ScalarMulPrecompile{},
	common.ECPrecompileBW6761PairingCheckContractAddress():     &bw6761PairingCheckPrecompile{},
	common.ECPrecompileBW6761PlonkProofVerifyContractAddress(): &bw6761PlonkProofVerifyPrecompile{},
}

// RunECPrecompiledContract executes and evaluates the output of an elliptic curve precompiled contract.
//
// It returns:
//   - ret: the returned bytes
//   - remainingGas: The gas that remains from the `suppliedGas` after the precompiled contract's gas cost has been
//     deducted.
//   - err: Any error that may have occurred during execution.
func RunECPrecompiledContract(evm *EVM, precompile ECPrecompiledContract, input []byte, suppliedGas uint64) (
	ret []byte, remainingGas uint64, err error,
) {
	gasCost := precompile.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	suppliedGas -= gasCost

	output, err := precompile.Run(evm, input)
	return output, suppliedGas, err
}

// bw6761PlonkProofVerifyPrecompile implements the verification of a given PlonK proof. It is assumed that the proof
// was generated using the BW6-761 curve.
//
// The input is assumed to encode, in the following order:
// - a PlonK proof of a circuit,
// - a verifying key for the proof,
// - a public witness suitable for verifying the proof.
//
// All the above components must be generated for the same circuit definition, over the BW6-761 finite field.
//
// Gnark version used by the proving service MUST BE the same as the one used by Quorum. The binary representation
// of the serialized objects is Gnark's internal implementation detail. No assumptions can be made about the byte-level
// structure of the input buffer. The only way of making sure the buffer is structured correctly is to maintain the
// order of the objects and keep Gnark versions synchronized among the interfacing projects.
//
// Example of serializing the objects into a single input buffer:
//
//	  // Assuming proof, verifyingKey and publicWitness were generated properly
//		var buf bytes.Buffer
//		proof.WriteTo(&buf)
//		verifyingKey.WriteTo(&buf)
//		publicWitness.WriteTo(&buf)
//	  // at this point buf contains all the objects and can be turned into []byte with buf.Bytes()
type bw6761PlonkProofVerifyPrecompile struct{}

func (*bw6761PlonkProofVerifyPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761PlonkProofVerifyGas
}

func (*bw6761PlonkProofVerifyPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	inputBytes := bytes.NewBuffer(input)

	proof := plonk.NewProof(ecc.BW6_761)
	_, err := proof.ReadFrom(inputBytes)
	if err != nil {
		log.Error("Proof verification: Reading proof", "error", err)
		return []byte{}, err
	}

	vk := plonk.NewVerifyingKey(ecc.BW6_761)
	_, err = vk.ReadFrom(inputBytes)
	if err != nil {
		log.Error("Proof verification: Reading vk", "error", err)
		return []byte{}, err
	}

	w, err := witness.New(ecc.BW6_761.ScalarField())
	if err != nil {
		log.Error("Proof verification: Creating witness", "error", err)
		return []byte{}, err
	}

	_, err = w.ReadFrom(inputBytes)
	if err != nil {
		log.Error("Proof verification: Reading witness", "error", err)
		return []byte{}, err
	}

	// Empty buffer is returned for API compatibility
	err = plonk_bw6761.Verify(
		proof.(*plonk_bw6761.Proof), vk.(*plonk_bw6761.VerifyingKey), w.Vector().(fr_bw6761.Vector),
	)
	if err != nil {
		log.Error("Proof verification: Verifying proof", "error", err)
		return []byte{}, err
	}

	log.Debug("Proof verification successful")

	return []byte{}, nil
}

// bw6761G1AddPrecompile implements the addition of G1 affine points where each coordinate is a 3-word field element.
//
// The input is assumed to encode all numbers in normal form using big-endian byte order. Operand encoding is as
// follows:
//
// - 96 bytes of x_1 coordinate for the first point
// - 96 bytes of y_1 coordinate for the first point
// - 96 bytes of x_2 coordinate for the second point
// - 96 bytes of y_2 coordinate for the second point
//
// It returns a G1 affine point consisting of 96 bytes for the x coordinate, and 96 bytes for the y coordinate. Both
// elements are encoded in big-endian byte ordering and normal form.
type bw6761G1AddPrecompile struct{}

func (*bw6761G1AddPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761G1AddGas
}

func (*bw6761G1AddPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	// Input contains four 96-byte numbers in normal form and using big-endian byte encoding.
	if len(input) != 2*sizeOfAffinePoint {
		return nil, errBW6InvalidInputLength
	}

	// After decoding, point1 is a G1 point with its elements in Montgomery form
	point1, err1 := decodeG1Point(input[:sizeOfAffinePoint])
	if err1 != nil {
		return nil, err1
	}

	// After decoding, point2 is a G1 point with its elements in Montgomery form
	point2, err2 := decodeG1Point(input[sizeOfAffinePoint:])
	if err2 != nil {
		return nil, err2
	}

	// Perform the actual math over the Jacobian as it's faster. FromAffine expects the G1Affine point to have elements
	// in Montgomery form.
	point1Jac := bw6761.G1Jac{}
	point1Jac.FromAffine(point1)
	point2Jac := bw6761.G1Jac{}
	point2Jac.FromAffine(point2)

	// Expects all elements in Montgomery form, which they are
	sumPointJac := point1Jac.AddAssign(&point2Jac)

	// This conversion also expects elements in Montgomery form, which they are
	sumPoint := bw6761.G1Affine{}
	sumPoint.FromJacobian(sumPointJac)

	// Encode back into bytes, which expects the field elements of the point in Montgomery form, but the result is in
	// normal form using big-endian byte ordering.
	enc := encodeG1Point(&sumPoint)

	return enc, nil
}

// bw6761G2AddPrecompile implements the addition of G2 affine points where each coordinate is a 3-word field element.
//
// Operand encoding is as follows:
//
// - 96 bytes of x_1 coordinate for the first point
// - 96 bytes of y_1 coordinate for the first point
// - 96 bytes of x_2 coordinate for the second point
// - 96 bytes of y_2 coordinate for the second point
//
// It returns a G2 affine point consisting of 3 words for the x coordinate, and 3 words for the y coordinate.
type bw6761G2AddPrecompile struct{}

func (*bw6761G2AddPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761G2AddGas
}

func (*bw6761G2AddPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	// Input contains four 96-byte numbers in normal form and using big-endian byte encoding.
	if len(input) != 2*sizeOfAffinePoint {
		return nil, errBW6InvalidInputLength
	}

	// After decoding, point1 is a G2 point with its elements in Montgomery form
	point1, err1 := decodeG2Point(input[:sizeOfAffinePoint])
	if err1 != nil {
		return nil, err1
	}

	// After decoding, point2 is a G2 point with its elements in Montgomery form
	point2, err2 := decodeG2Point(input[sizeOfAffinePoint:])
	if err2 != nil {
		return nil, err2
	}

	// Perform the actual math over the Jacobian as it's faster. FromAffine expects the G1Affine point to have elements
	// in Montgomery form.
	point1Jac := bw6761.G2Jac{}
	point1Jac.FromAffine(point1)
	point2Jac := bw6761.G2Jac{}
	point2Jac.FromAffine(point2)

	// Expects all elements in Montgomery form, which they are
	sumPointJac := point1Jac.AddAssign(&point2Jac)

	// This conversion also expects elements in Montgomery form, which they are
	sumPoint := bw6761.G2Affine{}
	sumPoint.FromJacobian(sumPointJac)

	// Encode back into bytes, which expects the field elements of the point in Montgomery form, but the result is in
	// normal form using big-endian byte ordering.
	enc := encodeG2Point(&sumPoint)

	return enc, nil
}

// bw6761G1ScalarMulPrecompile implements multiplication of a G1 affine point where each coordinate is a 3-word field
// element by a scalar value. The scalar value is expected to be an EVM word using big-endian encoding.
//
// Operand encoding is as follows:
//
// - 96 bytes of x coordinate for the point
// - 96 bytes of y coordinate for the point
// - 32 bytes of the scalar to multiply the point by.
//
// It returns a G1 affine point consisting of 3 words for the x coordinate, and 3 words for the y coordinate.
type bw6761G1ScalarMulPrecompile struct{}

func (*bw6761G1ScalarMulPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761G1MulGas
}

func (*bw6761G1ScalarMulPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	// Input contains two 96-byte numbers in normal form and using big-endian byte encoding, followed by a single 32
	// byte number in normal form using big-endian encoding
	if len(input) != sizeOfAffinePoint+sizeOfEVMWordBytes {
		return nil, errBW6InvalidInputLength
	}

	// After decoding, point is a G1 point with its elements in Montgomery form
	point, err1 := decodeG1Point(input[:sizeOfAffinePoint])
	if err1 != nil {
		return nil, err1
	}

	// After decoding, scalar is a big int in normal form
	scalar := big.NewInt(0).SetBytes(input[sizeOfAffinePoint:])

	// Now we can perform the multiplication, leaving point again as a G1 point in Montgomery form
	point.ScalarMultiplication(point, scalar)

	// And we can return the result
	return encodeG1Point(point), nil
}

// bw6761G2ScalarMulPrecompile implements multiplication of a G2 affine point where each coordinate is a 3-word field
// element by a scalar value. The scalar value is expected to be a word.
//
// Operand encoding is as follows:
//
// - 96 bytes of x coordinate for the point
// - 96 bytes of y coordinate for the point
// - 32 bytes of the scalar to multiply the point by.
//
// It returns a G2 affine point consisting of 3 words for the x coordinate, and 3 words for the y coordinate.
type bw6761G2ScalarMulPrecompile struct{}

func (*bw6761G2ScalarMulPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761G2MulGas
}

func (*bw6761G2ScalarMulPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	// Input contains two 96-byte numbers in normal form and using big-endian byte encoding, followed by a single 32
	// byte number in normal form using big-endian encoding
	if len(input) != sizeOfAffinePoint+sizeOfEVMWordBytes {
		return nil, errBW6InvalidInputLength
	}

	// After decoding, point is a G1 point with its elements in Montgomery form
	point, err1 := decodeG2Point(input[:sizeOfAffinePoint])
	if err1 != nil {
		return nil, err1
	}

	// After decoding, scalar is a big int in normal form
	scalar := big.NewInt(0).SetBytes(input[sizeOfAffinePoint:])

	// Now we can perform the multiplication, leaving point again as a G1 point in Montgomery form
	point.ScalarMultiplication(point, scalar)

	// And we can return the result
	return encodeG2Point(point), nil
}

// bw6761PairingCheckPrecompile implements the pairing check operation on two G1 affine points where each coordinate is
// a 3-word field element.
//
// Operand encoding is as follows:
//
// - 96 bytes of x_1 coordinate for the first point
// - 96 bytes of y_1 coordinate for the first point
// - 96 bytes of x_2 coordinate for the second point
// - 96 bytes of y_2 coordinate for the second point
//
// The result is 1 if the pairing is correct, and 0 otherwise, encoded in 1 byte.
type bw6761PairingCheckPrecompile struct{}

func (*bw6761PairingCheckPrecompile) RequiredGas(input []byte) uint64 {
	return params.Bw6761PairingGas
}

func (*bw6761PairingCheckPrecompile) Run(evm *EVM, input []byte) ([]byte, error) {
	// Input contains four 96-byte numbers in normal form and using big-endian byte encoding
	if len(input) != sizeOfAffinePoint*2 {
		return nil, errBW6InvalidInputLength
	}

	// After decoding, point1 is a G1 point with its elements in Montgomery form
	point1, err1 := decodeG1Point(input[:sizeOfAffinePoint])
	if err1 != nil {
		return nil, err1
	}

	// After decoding, point2 is a G2 point with its elements in Montgomery form
	point2, err2 := decodeG2Point(input[sizeOfAffinePoint:])
	if err2 != nil {
		return nil, err2
	}

	// Our data is in the right format, so we can now just compute the pairing
	result, err := bw6761.PairingCheck([]bw6761.G1Affine{*point1}, []bw6761.G2Affine{*point2})
	if err != nil {
		return nil, err
	}

	var wordResult byte
	if result {
		wordResult = 1
	} else {
		wordResult = 0
	}

	return []byte{wordResult}, nil
}

// decodeG1Point decodes a BW6 G1 Affine point from a byte array.
//
// It assumes that it is passed 192 bytes with the first 96 as the X coordinate and the second 96 as the Y coordinate.
// It assumes that these coordinates are in non-Montgomery form and are using big-endian encoding.
//
// It returns the G1 affine point from the bytes, with both elements in Montgomery form.
func decodeG1Point(input []byte) (*bw6761.G1Affine, error) {
	if len(input) != sizeOfAffinePoint {
		return nil, errBW6InvalidInputLength
	}

	var err error
	point := bw6761.G1Affine{}

	// Input is in normal big-endian, output is in Montgomery
	err = decodeFieldElementInto(input[:sizeOfFieldElement], &point.X)
	if err != nil {
		return nil, err
	}

	// Input is in normal big-endian, output is in Montgomery
	err = decodeFieldElementInto(input[sizeOfFieldElement:], &point.Y)
	if err != nil {
		return nil, err
	}

	return &point, nil
}

// decodeG2Point decodes a BW6 G2 Affine point from a byte array.
//
// It assumes that it is passed 192 bytes with the first 96 as the X coordinate and the second 96 as the Y coordinate.
// It assumes that these coordinates are in non-Montgomery form and are using big-endian encoding.
//
// It returns the G2 affine point from the bytes, with both elements in Montgomery form.
func decodeG2Point(input []byte) (*bw6761.G2Affine, error) {
	if len(input) != sizeOfAffinePoint {
		return nil, errBW6InvalidInputLength
	}

	var err error
	point := bw6761.G2Affine{}

	// Input is in normal big-endian, output is in Montgomery
	err = decodeFieldElementInto(input[:sizeOfFieldElement], &point.X)
	if err != nil {
		return nil, err
	}

	// Input is in normal big-endian, output is in Montgomery
	err = decodeFieldElementInto(input[sizeOfFieldElement:], &point.Y)
	if err != nil {
		return nil, err
	}

	return &point, nil
}

// encodeG1Point encodes a BW6 G1 affine point into a byte array.
//
// It assumes that the length of the byte array is 192 bytes. The X coordinate is in the first 96 bytes, and the Y
// coordinate is in the second 96 bytes. It assumes that the X and Y field elements in the input are in Montgomery form.
//
// The numbers encoded in the returned byte array are provided in normal form and using big-endian byte encoding.
func encodeG1Point(input *bw6761.G1Affine) []byte {
	output := make([]byte, sizeOfAffinePoint)

	// Bytes gets it in non-Montgomery, big-endian form, so we're good here.
	xBytes := input.X.Bytes()
	yBytes := input.Y.Bytes()
	copy(output[:sizeOfFieldElement], xBytes[:])
	copy(output[sizeOfFieldElement:], yBytes[:])

	return output
}

// encodeG2Point encodes a BW6 G2 affine point into a byte array.
//
// It assumes that the length of the byte array is 192 bytes. The X coordinate is in the first 96 bytes, and the Y
// coordinate is in the second 96 bytes. It assumes that the X and Y field elements in the input are in Montgomery form.
//
// The numbers encoded in the returned byte array are provided in normal form and using big-endian byte encoding.
func encodeG2Point(input *bw6761.G2Affine) []byte {
	output := make([]byte, sizeOfAffinePoint)

	// Bytes gets it in non-Montgomery, big-endian form, so we're good here.
	xBytes := input.X.Bytes()
	yBytes := input.Y.Bytes()
	copy(output[:sizeOfFieldElement], xBytes[:])
	copy(output[sizeOfFieldElement:], yBytes[:])

	return output
}

// decodeFieldElementInto decodes the input as a single field element, assuming that the input it is in big-endian
// encoding.
//
// The target `element` for the decoding contains the field element in Montgomery form when this call returns.
func decodeFieldElementInto(input []byte, element *fp.Element) error {
	// Input is big-endian and not Montgomery form
	if len(input) != sizeOfFieldElement {
		return errBW6InvalidFieldElementLength
	}

	// SetBytes interprets the input as the bytes of a BE unsigned integer, and sets element to the Montgomery form of
	// that integer.
	element.SetBytes(input)

	return nil
}
