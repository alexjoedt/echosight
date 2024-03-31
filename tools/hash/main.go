package main

import (
	"encoding/base64"
	"fmt"
	"os"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/spf13/pflag"
)

func main() {
	password := pflag.StringP("password", "p", "", "password to hash")
	convertToBase64 := pflag.Bool("base64", false, "converts the hash to base64")
	pflag.Parse()

	if *password == "" {
		fmt.Println("no password provided")
		pflag.Usage()
		os.Exit(1)
	}

	pass := echosight.Password{}
	err := pass.Set(*password)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	if *convertToBase64 {
		fmt.Println(base64.RawStdEncoding.EncodeToString(pass.Hash))
	} else {
		fmt.Println(string(pass.Hash))
	}
}
