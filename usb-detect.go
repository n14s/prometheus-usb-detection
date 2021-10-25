package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type usbDevice struct {
	Id   string
	Name string
}

var devices = map[string]usbDevice{}

func register(registerCmd *flag.FlagSet) {

	// registerCmd.Parse(os.Args[2:])

	fmt.Println("==Register a usb device==")

	// interact with user
	deviceName := getDeviceName()
	fmt.Println(deviceName)
	productNumber := getProductNumber()
	fmt.Println(productNumber)

	//reg device
	newDevice := usbDevice{productNumber, deviceName}
	devices[newDevice.Id] = newDevice

	//add udev rule for device
	addUdevRule(newDevice)
}

func getDeviceName() string {
	// get name for usb device
	fmt.Println("Please enter a name for the usb device")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	deviceName := scanner.Text()

	return deviceName
}

func getProductNumber() string {
	fmt.Println("Please plug in the usb device within 5 seconds\n Time starts after pressing enter.")
	fmt.Scanln()

	cmd := exec.Command("timeout", "--preserve-status", "5", "udevadm", "monitor", "--property")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	// searching product number
	// Before first subsystem=usb; the last product=; the text between = and \n
	i := strings.Index(string(out), "SUBSYSTEM=usb")
	li := strings.LastIndex(string(out)[:i], "PRODUCT=")
	nl := strings.Index(string(out[li:i]), "\n")
	productNumber := string(out)[li+len("PRODUCT=") : li+nl]

	return productNumber
}

func addUdevRule(newDevice usbDevice) {
	//create file or open file
	pathRuleFile := "/etc/udev/rules.d/85-usb-device.rules"
	f, err := os.OpenFile(pathRuleFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	// create rule string
	rule :=
		"ACTION==\"add\", SUBSYSTEM==\"usb\", ENV{PRODUCT}==\"" + newDevice.Id + "\", RUN+=\"prometheus-usb-detection add \"" + newDevice.Id + "\"\"\n" +
			"ACTION==\"remove\", SUBSYSTEM==\"usb\", ENV{PRODUCT}==\"" + newDevice.Id + "\",  RUN+=\"prometheus-usb-detection add " + newDevice.Id + "\""

	// check if string is already in file
	bytes, err := os.ReadFile(pathRuleFile)
	fileContent := string(bytes)
	if err != nil {
		log.Fatal(err)
		return
	}

	if strings.Contains(fileContent, newDevice.Id) {
		fmt.Println("Can't add device. Device has already been added")
	} else {
		// insert string in file
		_, err = fmt.Fprintln(f, rule)
		if err != nil {
			log.Fatal(err)
			f.Close()
			return
		}
		fmt.Println("Rule for device \"" + newDevice.Name + "\" added. (Id: \"" + newDevice.Id + "\")")
	}

	// close file
	err = f.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	reloadUdevRules()
}
func reloadUdevRules() {
	cmd := exec.Command("udevadm", "control", "--reload-rules", "&&", "udevadm", "trigger")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
func removeUdevRule(oldDevice usbDevice) {
	pathRuleFile := "/etc/udev/rules.d/85-usb-device.rules"
	if !fileExists(pathRuleFile) {
		fmt.Println("Can't remove usb device. Rule file does not exist.")
	} else {

		// read rule File
		bytes, err := ioutil.ReadFile(pathRuleFile)
		fileContent := string(bytes)
		if err != nil {
			log.Fatal(err)
			return
		}

		// delete lines with newDeviceId in it
		re := regexp.MustCompile("(?m)^.*" + oldDevice.Id + ".*$[\r\n]+")
		resultingFileContent := re.ReplaceAllString(fileContent, "")

		// change File to new String
		ioutil.WriteFile(pathRuleFile, []byte(resultingFileContent), 0)
		if err != nil {
			log.Fatal(err)
			return
		}

		reloadUdevRules()

		fmt.Println("Rule for device \"" + oldDevice.Name + "\" removed. (Id: \"" + oldDevice.Id + "\")")
	}
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func addDevice(addCmd *flag.FlagSet, id *string) {
	addCmd.Parse(os.Args[2:])
	fmt.Println(os.Args[1])
	fmt.Println(os.Args[2])

	if *id == "" {
		fmt.Println("ID is required")
		addCmd.PrintDefaults()
		os.Exit(1)
	}

	addedDevice := devices[*id]
	fmt.Println("The device ", addedDevice.Name, "has been plugged in")
	cmd := exec.Command("touch", "/tmp/ttt/ttt")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

func removeDevice(removeCmd *flag.FlagSet, id *string) {
	removeCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Print("ID is required")
		removeCmd.PrintDefaults()
		os.Exit(1)
	}

	removedDevice := devices[*id]
	fmt.Println("The device ", removedDevice.Name, "has been plugged out")

}
