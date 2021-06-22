# edgex-device-tuya-go

[English](./README.md) | [中文版](./README_zh.md)

## Overview
Tuya Micro Service - device service for connecting Tuya devices to EdgeX.

## Usage

### 1. Prepare

**Please create a project in the [Tuya IoT platform cloud development](https://iot.tuya.com/cloud/) in advance and make sure that at least one device is associated with the project**

### 2. Run edgex

Download the latest edgex

```
git clone https://github.com/edgexfoundry/edgex-compose.git
```

Enter the `edgex-compose` directory that has just been cloned locally 

- run in **no-sec** mode

```
make pull no-secty
make run no-secty
```

- run in **sec** mode

```
make pull
make run
make get-token
```

*You can use the `make down` command to stop all containers*

For more information about docker-compose startup, please see: https://github.com/edgexfoundry/edgex-compose/blob/master/README.md

For detailed information about token generation in sec mode, please see: https://github.com/edgexfoundry/edgex-go/blob/master/SECURITY.md

### 3. Run device-tuya-go

1. Please make sure that edgex has run successfully

2. Set environment variables

   Set to **true** if running in **sec** mode

   ```
   EDGEX_SECURITY_SECRET_STORE=false
   ```

3. Get the latest driver code from github

   ```
   git clone https://github.com/Tuya-Community/edgex-device-tuya-go.git
   ```

4. Modify the configuration file before the driver runs

   The points that must be modified are:

   - Modify `Service.Host` to the **exact IP address** of the host running the driver, **cannot** be localhost, 127.0.0.1, 0.0.0.0, etc.

   - Modify all the content under `[TuyaConnectorInfo]`. These data are the project-related information created in the first step Tuya IoT cloud platform. The rules for filling in `Region` are as follows:

     | Region        | Value |
     | ------------- | ----- |
     | China         | CN    |
     | United States | US    |
     | European      | EU    |
     | India         | IN    |
     
   - If you run edgex in **sec** mode, you need to modify the file address of `TokenFile` in `[SecretStore]`

5. Run drive

   Go to the cmd directory of the project and run

   ```
   go run main.go --cp=consul.http://localhost:8500 --registry
   ```

   Or, if you are using a `linux` system, you can also start the driver through docker, which is invalid for mac (because docker does not support host network mode on mac)

   - Get the docker image

     Enter the project root directory

     ```
     make docker_device_tuya_go
     ```

   - Run docker

     **Note that the directory when mounting is set to your own directory, If it is sec mode, set EDGEX_SECURITY_SECRET_STORE="true"**

     ```
     docker run --name edgex-device-tuya -v /your/local/path/device-tuya-go/cmd/res:/res --network=host -e EDGEX_SECURITY_SECRET_STORE="false" -d edgexfoundry/device-tuya:0.0.0-dev
     ```

6. After running successfully

   After the operation is successful, the device service will be automatically registered into the core-metadata of edgex and the name is `device-tuya`, which can be viewed by the following command

   ```
   curl http://localhost:59881/api/v2/deviceservice/name/device-tuya
   ```

### 4. Add device

We add the devices added to the iot project in the first step to the core-metadata of edgex

1. Add device profile

   A demo file is prepared in the `cmd/res/` directory of the project. The example file is a socket. You can modify it into the configuration file of the corresponding device according to this file. The demo configuration file is as follows:

   ```yaml
   name: "Test.Device.TUYA.Profile"	# This name must be unique
   manufacturer: "Tuya"
   model: "socket"
   labels:
     - "test"
   description: "Test device profile"
   deviceResources:	# What is defined here is the function point of the device
     -
       name: switch_1	# You can set this name as same as the `Code` below 
       isHidden: true
       description: "switch_1"
       attributes:
         { Code: "switch_1" }	# You can get this code from Tuya IoT platform
       properties:
         valueType: "Bool"
         readWrite: "RW"
         defaultValue: "false"
     -
       name: countdown_1
       isHidden: true
       description: "countdown for switch_1"
       attributes:
         { Code: "countdown_1" }
       properties:
         valueType: "Uint32"
         readWrite: "RW"
         defaultValue: "1"
         minimum: "0"
         maximum: "86400"
   
   deviceCommands:	# Defined here are device commands
     -
       name: switch_1	# You can set this name as same as the resource name
       readWrite: "RW"
       isHidden: false
       resourceOperations:
         - { deviceResource: "switch_1" }	# Resources corresponding to this command
     -
       name: countdown_1
       readWrite: "RW"
       isHidden: false
       resourceOperations:
         - { deviceResource: "countdown_1" }
   
   ```

   Reference page for cloud platform

   ![image-20210622114727864](./image/image-20210622151011131.png)

   After preparing the device configuration file, use the following command to register the device profile file into edgex's core-metadata service, **pay attention to modifying the path of the configuration file**

   ```
   curl http://localhost:59881/api/v2/deviceprofile/uploadfile -X POST -F "file=@<Fill in the specific profile file path>"
   ```

   If the above command does not report an exception, it means that the device profile file has been successfully added. You can use the following command to view the profile file just added. Note that the `Test.Device.TUYA.Profile` is changed to the name item in your profile file Value.

   ```
   curl http://localhost:59881/api/v2/deviceprofile/name/Test.Device.TUYA.Profile
   ```

2. Add device

   Execute the command below to add a device, pay attention to the item `DeviceId`, change it to the device id added in the cloud platform in the first step, `serviceName` is `device-tuya`, `profileName` is the profile name added in the previous step, `name `Is the name of the device (custom, tuya-test-device in the example).

   ```
   curl http://localhost:59881/api/v2/device -X POST -H "Content-Type: application/json" -d \
   '[
       {
           "requestId":"",
           "apiVersion":"v2",
           "device":{
               "name":"tuya-test-device",
               "description":"tuya device is created for test purpose",
               "adminState":"UNLOCKED",
               "operatingState":"UP",
               "labels":[
                   "TUYA",
                   "test"
               ],
               "serviceName":"device-tuya",
               "profileName":"Test.Device.TUYA.Profile",
               "protocols":{
                   "tuya":{
                       "DeviceId":"06870016bcddc237998d"
                   }
               }
           }
       }
   ]'
   
   ```

### 5. Send command

The api for sending the command is `http://localhost:59882/api/v2/device/name/<device_name>/<command_name>`

Take the demo as an example, device_name="tuya-test-device", command_name="switch_1"

1. Send `GET` command

   ```
   curl http://localhost:59882/api/v2/device/name/tuya-test-device/switch_1
   ```

2. Send `SET` command

   The SET command needs to be sent through the PUT method. The data type carried is json, the key is the command name, and the value is the value to be set (key and value are both string types), such as the following demo

   ```
   curl http://localhost:59882/api/v2/device/name/tuya-test-device/switch_1 -X PUT \
   	-H "Content-Type: application/json" -d \
   	'{
       "switch_1": "true"
      }'
   ```

## Community
- Chat: https://edgexfoundry.slack.com
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo
- Tuya Developer: https://developer.tuya.com/en/

## License
[MIT](LICENSE)
