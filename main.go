package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	someEnvVar string
}

var devices = map[string]usbDevice{}

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

	// fill map with udevrule devices
	readRegisteredDevices()

	// add subcommands
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)

	updateStateCmd := flag.NewFlagSet("updateState", flag.ExitOnError)
	addID := updateStateCmd.String("add", "", "Tell Prometheus which device has been added")
	removeID := updateStateCmd.String("remove", "", "Tell prometheus which device has been removed")
	// parse here of when opening command?
	//flag.Parse()

	// if args < 1 default to register asd aasd
	if len(os.Args) <= 1 {
		register(registerCmd)
	} else {
		switch os.Args[1] {
		case "register":
			register(registerCmd)
		case "updateState":
			updateStateCmd.Parse(os.Args[2:])
			updateState(updateStateCmd, addID, removeID)
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
