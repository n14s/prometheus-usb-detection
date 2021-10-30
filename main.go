package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	PathDeviceFile     string
	PathPrometheusFile string
}

var devices = map[string]usbDevice{}
var pathRuleFile = "/etc/udev/rules.d/85-usb-device.rules"
var appDir = getAppDir()
var pathAppFile = appDir + "/prometheus-usb-detection"
var pathDevicesFile = appDir + "/devices.json"
var pathPrometheusFile = appDir + "/usb-device.prom"

func main() {
	fmt.Println("--- USB-DETECTION ---")

	// initialize default Config Values
	myConfig := Config{pathDevicesFile, pathPrometheusFile}

	// read config from file
	if fileExists("./config.json") {
		myConfig = readConfig(myConfig)
	}

	//read envvars

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

func readConfig(myConfig Config) Config {

	bytes, err := os.ReadFile("./config.json")
	check(err)

	defaultConfig := "{\"PathDeviceFile\":\"./devices.json\",\"PathPrometheusFile\":\"./usb-device.prom\"}"
	err = json.Unmarshal(bytes, &myConfig)
	if err != nil || string(bytes) == defaultConfig {
		fmt.Println("true")
		myConfig = Config{pathDevicesFile, pathPrometheusFile}
	}
	return myConfig
}
