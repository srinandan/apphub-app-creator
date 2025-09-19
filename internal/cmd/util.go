package cmd

import "github.com/spf13/pflag"

func GetStringParam(flag *pflag.Flag) (param string) {
	param = ""
	if flag != nil {
		param = flag.Value.String()
	}
	return param
}
