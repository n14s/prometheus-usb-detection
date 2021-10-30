package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type usbDevice struct {
	Id   string
	Name string
}

var devices = map[string]usbDevice{}
var pathRuleFile = "/etc/udev/rules.d/85-usb-device.rules"
var appDir = getAppDir()
var appFile = appDir + "/prometheus-usb-detection"
var devicesFile = appDir + "/devices.json"
var prometheusFile = appDir + "/usb-device.prom"

func register(registerCmd *flag.FlagSet) {

	// registerCmd.Parse(os.Args[2:])

	fmt.Println("==Register a usb device==")

	// interact with user
	deviceName := getDeviceName()
	if containsValue(devices, deviceName) {
		fmt.Println("Can't add device. Device has already been added")
	} else {
		productNumber := getProductNumber()

		//reg device
		newDevice := usbDevice{productNumber, deviceName}
		devices[newDevice.Id] = newDevice

		//add udev rule for device
		alreadyExists := addUdevRule(newDevice)

		// write to json if new
		if !alreadyExists {
			writeRegisteredDevices()
		}
	}

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

func addUdevRule(newDevice usbDevice) bool {
	alreadyExists := false
	//create file or open file
	f, err := os.OpenFile(pathRuleFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)

	// create rule string
	rule :=
		"ACTION==\"add\", SUBSYSTEM==\"usb\", ENV{PRODUCT}==\"" + newDevice.Id + "\", RUN+=\"" + appFile + " updateState -add " + newDevice.Id + "\"\n" +
			"ACTION==\"remove\", SUBSYSTEM==\"usb\", ENV{PRODUCT}==\"" + newDevice.Id + "\", RUN+=\"" + appFile + " updateState -remove " + newDevice.Id + "\""

	// check if string is already in file
	bytes, err := os.ReadFile(pathRuleFile)
	fileContent := string(bytes)
	check(err)

	if strings.Contains(fileContent, newDevice.Id) {
		fmt.Println("Can't add device. Device has already been added")
		alreadyExists = true
	} else {
		// insert string in file
		_, err = fmt.Fprintln(f, rule)
		if err != nil {
			log.Fatal(err)
			f.Close()
		}
		fmt.Println("Rule for device \"" + newDevice.Name + "\" added. (Id: \"" + newDevice.Id + "\")")
	}

	// close file
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
	reloadUdevRules()
	return alreadyExists
}
func reloadUdevRules() {
	cmd := exec.Command("udevadm", "control", "--reload-rules", "&&", "udevadm", "trigger")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
func removeUdevRule(oldDevice usbDevice) {
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

func updateState(addCmd *flag.FlagSet, addID *string, removeID *string) {
	if *addID == "" && *removeID == "" {
		fmt.Println("No ID passed.")
		addCmd.PrintDefaults()
		os.Exit(1)
	}

	if *addID != "" && *removeID != "" {
		fmt.Println("Cannot add and remove at the same time.")
		addCmd.PrintDefaults()
		os.Exit(1)
	}

	ok := false
	device := usbDevice{}
	isUp := 0

	if *addID != "" {
		device, ok = devices[*addID]
		isUp = 1
	} else {
		device, ok = devices[*removeID]
		isUp = 0
	}
	//if *removeID != ""
	if !ok {
		fmt.Println("Device is not registered.")
		addCmd.PrintDefaults()
		os.Exit(1)
	} else {
		if isUp == 1 {
			fmt.Println("The device", device.Name, "has been plugged in")
		} else {
			fmt.Println("The device", device.Name, "has been plugged out")
		}

		// create or open rule file
		f, err := os.OpenFile(prometheusFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		check(err)

		err = f.Close()
		check(err)

		input, err := ioutil.ReadFile(prometheusFile)
		check(err)

		lines := strings.Split(string(input), "\n")
		found := false

		for i, line := range lines {
			if strings.Contains(line, device.Name) {
				lines[i] = device.Name + " " + fmt.Sprint(isUp)
				found = true
				break
			}
		}

		if !found {
			if lines[0] == "" {
				lines[0] = device.Name + " " + fmt.Sprint(isUp)
			} else {
				lines = append(lines, device.Name+" "+fmt.Sprint(isUp))
			}
		}

		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(prometheusFile, []byte(output), 0644)
		check(err)
	}
}

func writeRegisteredDevices() {
	f, err := os.OpenFile(devicesFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	check(err)

	bytes, err := json.Marshal(devices)
	jsonString := string(bytes)
	check(err)

	f.WriteString(jsonString)

	err = f.Close()
	check(err)

}
func readRegisteredDevices() {
	// read rule File

	fmt.Println(devicesFile)
	if fileExists(devicesFile) {
		bytes, err := os.ReadFile(devicesFile)
		check(err)

		err = json.Unmarshal(bytes, &devices)
		check(err)
	}

}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func containsValue(m map[string]usbDevice, v string) bool {
	for _, x := range m {
		if x.Name == v {
			return true
		}
	}
	return false
}

func getAppDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
