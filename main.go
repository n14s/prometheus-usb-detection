package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	someEnvVar string
}

func main() {
	fmt.Println("--- USB-DETECTION ---")

	//test
	/*
		testdev := usbDevice{"5235", "dings"}
		addUdevRule(testdev)
		removeUdevRule(testdev)
	*/

	//read envvars
	var config Config
	config.someEnvVar = os.Getenv("SOME_ENV")

	// add subcommands
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addID := addCmd.String("add", "", "Tell Prometheus which device has been added")

	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	removeID := removeCmd.String("remove", "", "Tell prometheus which device has been removed")
	// parse here of when opening command?
	//flag.Parse()

	// if args < 1 default to register asd aasd
	if len(os.Args) <= 1 {
		register(registerCmd)
	} else {
		switch os.Args[1] {
		case "register":
			register(registerCmd)
		case "add":
			addDevice(addCmd, addID)
		case "remove":
			removeDevice(removeCmd, removeID)
		default:
			register(registerCmd)
		}
	}

}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
