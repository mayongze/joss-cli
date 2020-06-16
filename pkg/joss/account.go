package joss

import "errors"

type Account struct {
	AccessKey string
	SecretKey string
}

func GetAccount() (account Account, err error) {
	ak, sk := AccessKey(), SecretKey()
	if ak != "" && sk != "" {
		return Account{
			AccessKey: ak,
			SecretKey: sk,
		}, nil
	}
	err = errors.New("empty")
	return
}
