package command

import (
	"testing"
)

func TestGetAcount(t *testing.T) {
	account, err := GetAcount()
	if err != nil {
		t.Errorf("GetAccount error,%v \n", err)
		return
	}
	t.Logf("ak=%s,sk=%s \n", account.AccessKey, account.SecretKey)
}

func TestSetAccount(t *testing.T) {
	account := Account{
		AccessKey: "443F8B1324FA36B5FABC2F81F08C19A1",
		SecretKey: "6B9148CD6A52F219593EFD1DC48387F8",
	}
	err := SetAccount(account.AccessKey, account.SecretKey)
	if err != nil {
		t.Errorf("SetAccount error, %v \n", err)
		return
	}
	account, err = GetAcount()
	if err != nil {
		t.Errorf("setAccount getAcount error.")
		return
	}
	t.Logf(account.String())
}
