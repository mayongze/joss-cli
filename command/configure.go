package command

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type Account struct {
	AccessKey string "json:\"access_key\""
	SecretKey string "json:\"secret_key\""
}

var(
	account Account
)

func (acc *Account) ToJson() (jsonStr string, err error) {
	jsonData, mErr := json.Marshal(acc)
	if mErr != nil {
		err = fmt.Errorf("Marshal account data error, %s", mErr)
		return
	}
	jsonStr = string(jsonData)
	return
}

func (acc *Account) String() string {
	return fmt.Sprintf("AccessKey: %s\nSecretKey: %s", acc.AccessKey, acc.SecretKey)
}

func NewAccountCommand() *cobra.Command {
	bc := &cobra.Command{
		Use:   "account --ak=***** --sk=*****",
		Short: "key operation command",
		Run: acountCommandFunc,
	}
	bc.Flags().StringVar(&account.AccessKey,"ak","","accessKey.")
	bc.Flags().StringVar(&account.SecretKey,"sk","","secretKey.")
	return bc
}

func acountCommandFunc (cmd *cobra.Command, args []string){
	//为空默认
	if account.AccessKey == "" || account.SecretKey== ""{
		account, err := GetAcount()
		if err != nil{
			ExitWithError(ExitError,err)
		}
		fmt.Fprintf(os.Stdout,"ak=%s\nsk=%s\n",account.AccessKey,account.SecretKey)
	}else{
		err := SetAccount(account.AccessKey,account.SecretKey)
		if err != nil{
			ExitWithError(ExitError,err)
		}
		fmt.Fprintf(os.Stdout,"OK")
	}
}

//配置文件>环境变量
func GetAcount() (account Account,err error){
	//获取失败从环境变量获取
	defer func() {
		if err != nil{
			account.AccessKey = os.Getenv("ACCESSKEY")
			account.SecretKey = os.Getenv("SECRETKEY")
			if account.AccessKey == "" || account.SecretKey == "" {
				err = fmt.Errorf("key not found")
			}
		}
	}()
	//获取用户路径
	curUser, err := user.Current()
	if  err != nil {
		return account,fmt.Errorf("Error: get current user error,%s \n",err)
	}
	storageDir := filepath.Join(curUser.HomeDir, ".jdcloud")
	accountFname := filepath.Join(storageDir, "account.json")
	accountFh, openErr := os.Open(accountFname)
	if openErr != nil {
		return account,fmt.Errorf("Open account file error, %s",openErr)
	}
	defer accountFh.Close()

	accountBytes, readErr := ioutil.ReadAll(accountFh)
	if readErr != nil {
		return account,fmt.Errorf("Read account file error, %s",readErr)
	}

	if umError := json.Unmarshal(accountBytes, &account); umError != nil {
		err = fmt.Errorf("Parse account file error, %s", umError)
		return
	}
	return account, nil
}


func SetAccount(accessKey,secretKey string) (err error) {

	if accessKey == "" || secretKey == "" {
		fmt.Errorf("accesskey secretkey cannot be empty.")
	}

	//获取用户路径
	curUser, err := user.Current()
	if  err != nil {
		return fmt.Errorf("Error: get current user error,%s \n",err)
	}
	storageDir := filepath.Join(curUser.HomeDir, ".jdcloud")
	if _, sErr := os.Stat(storageDir); sErr != nil {
		if mErr := os.MkdirAll(storageDir, 0755); mErr != nil {
			err = fmt.Errorf("Mkdir `%s` error, %s", storageDir, mErr)
			return
		}

	}
	accountFname := filepath.Join(storageDir, "account.json")
	accountFh, openErr := os.OpenFile(accountFname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if openErr != nil {
		err = fmt.Errorf("Open account file error, %s", openErr)
		return
	}
	defer accountFh.Close()

	account := Account{AccessKey:accessKey,SecretKey:secretKey}
	jsonStr, err := account.ToJson()
	if err != nil {
		return
	}
	_, wErr := accountFh.WriteString(jsonStr)
	if wErr != nil {
		err = fmt.Errorf("Write account info error, %s", wErr)
		return
	}
	return
}
