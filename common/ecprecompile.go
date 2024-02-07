// Package common
//
// We need these as constants, but in a type that doesn't support easy constant initialization, so we have them here
// as functions.

package common

func ECPrecompileBW6761G1AddContractAddress() Address {
	return BytesToAddress([]byte{byte(0x80)})
}

func ECPrecompileBW6761G1ScalarMulContractAddress() Address {
	return BytesToAddress([]byte{byte(0x81)})
}

func ECPrecompileBW6761G2AddContractAddress() Address {
	return BytesToAddress([]byte{byte(0x82)})
}

func ECPrecompileBW6761G2ScalarMulContractAddress() Address {
	return BytesToAddress([]byte{byte(0x83)})
}

func ECPrecompileBW6761PairingCheckContractAddress() Address {
	return BytesToAddress([]byte{byte(0x84)})
}

func ECPrecompileBW6761PlonkProofVerifyContractAddress() Address {
	return BytesToAddress([]byte{byte(0x85)})
}
