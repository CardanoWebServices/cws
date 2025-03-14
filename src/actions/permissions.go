package actions

import (
	"errors"
	"github.com/fxamacker/cbor/v2"
	"ows/ledger"
)

type AddUser struct {
	BaseAction
	Key ledger.PubKey `cbor:"0,keyasint"`
}

func NewAddUser(key ledger.PubKey) *AddUser {
	return &AddUser{BaseAction{}, key}
}

func (c *AddUser) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	return errors.New("AddUser.Apply() not yet implemented")
}

func (c *AddUser) GetName() string {
	return "AddUser"
}

func (c *AddUser) GetCategory() string {
	return "permissions"
}

var _AddUserRegistered = ledger.RegisterAction("permissions", "AddUser", func(attr []byte) (ledger.Action, error) {
	var c AddUser
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})
