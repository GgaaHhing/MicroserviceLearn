package internal

import (
	"fmt"
	"testing"
)

func TestReg(t *testing.T) {
	err := Reg(ViperConf.AccountWebConfig.SrvName, ViperConf.AccountWebConfig.Host,
		ViperConf.AccountWebConfig.SrvName, ViperConf.AccountWebConfig.Port,
		ViperConf.AccountWebConfig.Tags)
	fmt.Println(err)
}
