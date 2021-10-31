# prometheus-usb-detection
Prometheus alarm for (dis)connecting usb-devices on linux

# Installation

No installation needed, if you use the built binary [prometheus-usb-detection](./prometheus-usb-detection).  
If you want to build it yourself, clone the project and run `go build`.

# Usage

This app offers two subcommands `prometheus-usb-detection register` and `prometheus-usb-detection updateStatus`

## Registration
INFO: Since this app uses linux udev-rules that only allow changes by privileged users, the app must be run as sudo when registering. If you have concerns feel free to check the code and build the binary from source code.

The `sudo prometheus-usb-detection register` subcommands allows for registration of usb-devices. 
After command execution, follow the steps provided by the app's command line interface.  
You will be prompted for a device name. Afterwards you will need to plug in the usb-device to finish up it's registration. The devices are saved in the devices.json

## UpdateStatus
The subcommand `prometheus-usb-detection updateStatus` is not needed for the end-user.  
It is invoked by linux automatically, whenever there are events on registered usb-devices, updating the prometheus file. This subcommand can also be used for debugging.

## Prometheus 

To integrate any kind of metric into its ecosystem prometheus offers the [textfile-collector within its node-exporter package](https://github.com/prometheus/node_exporter#user-content-textfile-collector).   
Prometheus-usb-dection builds on top of it, updating a usb-device.prom file whenever a registered device is added or removed.   
Integrate the usb-device.prom file (or the folder where the file resides) in your textfile-collector and the metrics are exposed in the node-exporter.

# Configuration

This app comes with a config file where you can define the absolute path of your device.json file and more importantly the usb-device.prom file.  
If there are no changes or the changes are invalid it defaults to the folder of the prometheus-usb-detection binary.

# Issues
If you want to request new features or report bugs, please create an issue.