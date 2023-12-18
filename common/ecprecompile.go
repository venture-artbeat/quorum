package common

func ECPrecompileHelloContractAddress() Address {
	return BytesToAddress([]byte{byte(0x8a)})
}
