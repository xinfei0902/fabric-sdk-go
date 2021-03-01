package dconfig

import (
	"os"

	"../derrors"
	"github.com/spf13/cobra"
)

func AddFlags(one *cobra.Command) (err error) {
	flags := one.Flags()
	for key, v := range globalFlags {
		if one, ok := globalContainer[key]; true == ok && one.Command != nil {
			continue
		}
		if v.Value == nil {
			p := flags.StringP(v.Name, v.Short, "", v.Usage)
			value, _ := globalContainer[key]
			value.Command = p
			globalContainer[key] = value
			continue
		}
		switch v.Value.(type) {
		case int, *int:
			p := flags.IntP(v.Name, v.Short, 0, v.Usage)
			value, _ := globalContainer[key]
			value.Command = p
			globalContainer[key] = value
		case bool, *bool:
			p := flags.BoolP(v.Name, v.Short, false, v.Usage)
			value, _ := globalContainer[key]
			value.Command = p
			globalContainer[key] = value
		case string, *string:
			p := flags.StringP(v.Name, v.Short, "", v.Usage)
			value, _ := globalContainer[key]
			value.Command = p
			globalContainer[key] = value
		default:
			panic("add code here")
		}
	}
	return
}

func ReadFileAndHoldCommand(cmd *cobra.Command) (err error) {
	tmp := make(map[string]interface{})

	filename := ""

	if len(globalFileKey) > 0 {
		v, ok := globalContainer[globalFileKey]
		if !ok {
			return derrors.ErrorKeyNotContainValuef(globalFileKey)
		}
		filename = v.GetStringValue()
		if len(filename) == 0 {
			return derrors.ErrorKeyNotContainValuef(globalFileKey)
		}
		tmp, err = ReadConfigFile(filename)
		if err != nil {
			if false == os.IsNotExist(err) {
				return err
			}
		}
	}
	for key, value := range tmp {
		one, _ := globalContainer[key]
		one.File = value
		globalContainer[key] = one
	}

	for key, value := range globalContainer {
		globalValues[key] = value.GetInterface()
	}

	if len(filename) > 0 {
		Flush(filename)
	}

	return nil
}
