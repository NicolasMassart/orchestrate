package testdata

import (
	"encoding/json"

	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
)

func FakeRegisterContractRequest() *api.RegisterContractRequest {
	c := testdata.FakeContract()
	var abi interface{}
	_ = json.Unmarshal([]byte(c.RawABI), &abi)

	return &api.RegisterContractRequest{
		Name:             c.Name,
		Tag:              c.Tag,
		ABI:              abi,
		Bytecode:         c.Bytecode,
		DeployedBytecode: c.DeployedBytecode,
	}
}

func FakeSetContractCodeHashRequest() *api.SetContractCodeHashRequest {
	return &api.SetContractCodeHashRequest{
		CodeHash: testdata.FakeHash(),
	}
}

func FakeSearchContractRequest() *api.SearchContractRequest {
	return &api.SearchContractRequest{
		CodeHash: testdata.FakeHash(),
		Address:  testdata.FakeAddress(),
	}
}

func FakeContractResponse() *api.ContractResponse {
	c := testdata.FakeContract()
	return &api.ContractResponse{
		Name:             c.Name,
		Tag:              c.Tag,
		ABI:              c.RawABI,
		Bytecode:         c.Bytecode.String(),
		DeployedBytecode: c.DeployedBytecode.String(),
	}
}
