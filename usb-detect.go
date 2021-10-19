package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func getUsbDev() {
	fmt.Println("Please plug in your usb device")
	fmt.Println("Press enter and plug in the usb device within 5 seconds")
	fmt.Scanln()
	// dummy cmd
	//	cmd := exec.Command("echo", "moni")

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
	fmt.Println(productNumber)

	// get name for usb device
	fmt.Println("Please enter a name for the usb device")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	usbDevice := scanner.Text()

	fmt.Println(usbDevice)

}

func register(registerCmd *flag.FlagSet) {
	registerCmd.Parse(os.Args[2:])
}

func addDevice(addCmd *flag.FlagSet, id *string) {
	addCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Print("ID is required")
		addCmd.PrintDefaults()
		os.Exit(1)
	}
}

func removeDevice(removeCmd *flag.FlagSet, id *string) {
	removeCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Print("ID is required")
		removeCmd.PrintDefaults()
		os.Exit(1)
	}
}
