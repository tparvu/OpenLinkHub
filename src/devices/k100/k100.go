package k100

// Package: K100 RGB
// This is the primary package for K100 RGB
// All device actions are controlled from this package.
// Author: Nikola Jurkovic
// License: GPL-3.0 or later

import (
	"OpenLinkHub/src/common"
	"OpenLinkHub/src/config"
	"OpenLinkHub/src/inputmanager"
	"OpenLinkHub/src/keyboards"
	"OpenLinkHub/src/logger"
	"OpenLinkHub/src/rgb"
	"OpenLinkHub/src/temperatures"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/sstallion/go-hid"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// DeviceProfile struct contains all device profile
type DeviceProfile struct {
	Active          bool
	Path            string
	Product         string
	Serial          string
	LCDMode         uint8
	LCDRotation     uint8
	Brightness      uint8
	RGBProfile      string
	Label           string
	Layout          string
	Keyboards       map[string]*keyboards.Keyboard
	Profile         string
	BrightnessLevel uint16
	Profiles        []string
	ControlDial     int
}

type Device struct {
	Debug              bool
	dev                *hid.Device
	listener           *hid.Device
	Manufacturer       string `json:"manufacturer"`
	Product            string `json:"product"`
	Serial             string `json:"serial"`
	Firmware           string `json:"firmware"`
	activeRgb          *rgb.ActiveRGB
	UserProfiles       map[string]*DeviceProfile `json:"userProfiles"`
	Devices            map[int]string            `json:"devices"`
	DeviceProfile      *DeviceProfile
	OriginalProfile    *DeviceProfile
	Template           string
	VendorId           uint16
	ProductId          uint16
	Brightness         map[int]string
	LEDChannels        int
	CpuTemp            float32
	GpuTemp            float32
	Layouts            []string
	ControlDialOptions map[int]string
	Rgb                *rgb.RGB
}

var (
	pwd             = ""
	cmdSoftwareMode = []byte{0x01, 0x03, 0x00, 0x02}
	cmdHardwareMode = []byte{0x01, 0x03, 0x00, 0x01}
	cmdActivateLed  = []byte{0x0d, 0x01, 0x22}
	cmdBrightness   = map[int][]byte{
		0: {0x01, 0x02, 0x00},
		1: {0x01, 0x6d, 0x00},
	}
	cmdGetFirmware          = []byte{0x02, 0x13}
	dataTypeSetColor        = []byte{0x12, 0x00}
	dataTypeSubColor        = []byte{0x07, 0x00}
	cmdWriteColor           = []byte{0x06, 0x01}
	deviceRefreshInterval   = 1000
	deviceKeepAlive         = 20000
	timer                   = &time.Ticker{}
	timerKeepAlive          = &time.Ticker{}
	authRefreshChan         = make(chan bool)
	keepAliveChan           = make(chan bool)
	mutex                   sync.Mutex
	transferTimeout         = 500
	bufferSize              = 1024
	bufferSizeWrite         = bufferSize + 1
	headerSize              = 2
	headerWriteSize         = 4
	maxBufferSizePerRequest = 1021
	colorPacketLength       = 581
	keyboardKey             = "k100-default"
	defaultLayout           = "k100-default-US"
)

func Init(vendorId, productId uint16, key string) *Device {
	// Set global working directory
	pwd = config.GetConfig().ConfigPath

	dev, err := hid.OpenPath(key)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "vendorId": vendorId, "productId": productId}).Error("Unable to open HID device")
		return nil
	}

	// Init new struct with HID device
	d := &Device{
		dev:       dev,
		Template:  "k100.html",
		VendorId:  vendorId,
		ProductId: productId,
		Brightness: map[int]string{
			0: "RGB Profile",
			1: "33 %",
			2: "66 %",
			3: "100 %",
		},
		Product:     "K100 RGB",
		LEDChannels: 194,
		Layouts:     keyboards.GetLayouts(keyboardKey),
		ControlDialOptions: map[int]string{
			1: "Brightness",
		},
	}

	d.getDebugMode()       // Debug mode
	d.getManufacturer()    // Manufacturer
	d.getSerial()          // Serial
	d.loadRgb()            // Load RGB
	d.setSoftwareMode()    // Activate software mode
	d.initLeds()           // Init LED ports
	d.getDeviceFirmware()  // Firmware
	d.loadDeviceProfiles() // Load all device profiles
	d.saveDeviceProfile()  // Save profile
	d.setAutoRefresh()     // Set auto device refresh
	d.setKeepAlive()       // Keepalive
	d.setDeviceColor()     // Device color
	d.setBrightnessLevel() // Brightness
	d.controlListener()    // Control listener
	return d
}

// Stop will stop all device operations and switch a device back to hardware mode
func (d *Device) Stop() {
	logger.Log(logger.Fields{"serial": d.Serial}).Info("Stopping device...")
	if d.activeRgb != nil {
		d.activeRgb.Stop()
	}
	timer.Stop()
	authRefreshChan <- true

	timerKeepAlive.Stop()
	keepAliveChan <- true

	d.setHardwareMode()
	if d.dev != nil {
		err := d.dev.Close()
		if err != nil {
			logger.Log(logger.Fields{"error": err}).Error("Unable to close HID device")
		}
	}
}

// loadRgb will load RGB file if found, or create the default.
func (d *Device) loadRgb() {
	rgbDirectory := pwd + "/database/rgb/"
	rgbFilename := rgbDirectory + d.Serial + ".json"

	// Check if filename has .json extension
	if !common.IsValidExtension(rgbFilename, ".json") {
		return
	}

	if !common.FileExists(rgbFilename) {
		profile := rgb.GetRGB()
		profile.Device = d.Product
		rotator := profile.Profiles["rotator"]
		rotator.Brightness = 1
		profile.Profiles["rotator"] = rotator

		// Convert to JSON
		buffer, err := json.MarshalIndent(profile, "", "    ")
		if err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to encode RGB json")
			return
		}

		// Create profile filename
		file, err := os.Create(rgbFilename)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to create RGB json file")
			return
		}

		// Write JSON buffer to file
		_, err = file.Write(buffer)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to write to RGB json file")
			return
		}

		// Close file
		err = file.Close()
		if err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to close RGB json file")
			return
		}
	}

	file, err := os.Open(rgbFilename)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to load RGB")
		return
	}
	if err = json.NewDecoder(file).Decode(&d.Rgb); err != nil {
		logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": rgbFilename}).Warn("Unable to decode profile")
		return
	}
	err = file.Close()
	if err != nil {
		logger.Log(logger.Fields{"location": rgbFilename, "serial": d.Serial}).Warn("Failed to close file handle")
	}
}

// GetRgbProfile will return rgb.Profile struct
func (d *Device) GetRgbProfile(profile string) *rgb.Profile {
	if d.Rgb == nil {
		return nil
	}

	if val, ok := d.Rgb.Profiles[profile]; ok {
		return &val
	}
	return nil
}

// GetDeviceTemplate will return device template name
func (d *Device) GetDeviceTemplate() string {
	return d.Template
}

// getManufacturer will return device manufacturer
func (d *Device) getDebugMode() {
	d.Debug = config.GetConfig().Debug
}

// getManufacturer will return device manufacturer
func (d *Device) getManufacturer() {
	manufacturer, err := d.dev.GetMfrStr()
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to get manufacturer")
	}
	d.Manufacturer = manufacturer
}

// getProduct will return device name
func (d *Device) getProduct() {
	product, err := d.dev.GetProductStr()
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to get product")
	}
	d.Product = product
}

// getSerial will return device serial number
func (d *Device) getSerial() {
	serial, err := d.dev.GetSerialNbr()
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to get device serial number")
	}
	d.Serial = serial
}

// setHardwareMode will switch a device to hardware mode
func (d *Device) setHardwareMode() {
	_, err := d.transfer(cmdHardwareMode, nil)
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to change device mode")
	}
}

// setSoftwareMode will switch a device to software mode
func (d *Device) setSoftwareMode() {
	_, err := d.transfer(cmdSoftwareMode, nil)
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to change device mode")
	}
}

// getDeviceFirmware will return a device firmware version out as string
func (d *Device) getDeviceFirmware() {
	fw, err := d.transfer(
		cmdGetFirmware,
		nil,
	)
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to write to a device")
	}

	v1, v2, v3 := int(fw[3]), int(fw[4]), int(binary.LittleEndian.Uint16(fw[5:7]))
	d.Firmware = fmt.Sprintf("%d.%d.%d", v1, v2, v3)
}

// initLeds will initialize LED ports
func (d *Device) initLeds() {
	_, err := d.transfer(cmdActivateLed, nil)
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to change device mode")
	}
	// We need to wait around 500 ms for physical ports to re-initialize
	// After that we can grab any new connected / disconnected device values
	time.Sleep(time.Duration(transferTimeout) * time.Millisecond)
}

// saveDeviceProfile will save device profile for persistent configuration
func (d *Device) saveDeviceProfile() {
	profilePath := pwd + "/database/profiles/" + d.Serial + ".json"
	keyboardMap := make(map[string]*keyboards.Keyboard, 0)

	deviceProfile := &DeviceProfile{
		Product: d.Product,
		Serial:  d.Serial,
		Path:    profilePath,
	}

	// First save, assign saved profile to a device
	if d.DeviceProfile == nil {
		// RGB, Label
		deviceProfile.RGBProfile = "keyboard"
		deviceProfile.Label = "Keyboard"
		deviceProfile.Active = true
		keyboardMap["default"] = keyboards.GetKeyboard(defaultLayout)
		deviceProfile.Keyboards = keyboardMap
		deviceProfile.Profile = "default"
		deviceProfile.Profiles = []string{"default"}
		deviceProfile.BrightnessLevel = 1000
		deviceProfile.Brightness = 0
		deviceProfile.Layout = "US"
		deviceProfile.ControlDial = 1
	} else {
		if len(d.DeviceProfile.Layout) == 0 {
			deviceProfile.Layout = "US"
		} else {
			deviceProfile.Layout = d.DeviceProfile.Layout
		}
		deviceProfile.Active = d.DeviceProfile.Active
		deviceProfile.Brightness = d.DeviceProfile.Brightness
		deviceProfile.RGBProfile = d.DeviceProfile.RGBProfile
		deviceProfile.Label = d.DeviceProfile.Label
		deviceProfile.Profile = d.DeviceProfile.Profile
		deviceProfile.Profiles = d.DeviceProfile.Profiles
		deviceProfile.Keyboards = d.DeviceProfile.Keyboards
		deviceProfile.ControlDial = d.DeviceProfile.ControlDial
		deviceProfile.BrightnessLevel = d.DeviceProfile.BrightnessLevel
		if len(d.DeviceProfile.Path) < 1 {
			deviceProfile.Path = profilePath
			d.DeviceProfile.Path = profilePath
		} else {
			deviceProfile.Path = d.DeviceProfile.Path
		}
		deviceProfile.LCDMode = d.DeviceProfile.LCDMode
		deviceProfile.LCDRotation = d.DeviceProfile.LCDRotation
	}

	// Convert to JSON
	buffer, err := json.MarshalIndent(deviceProfile, "", "    ")
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Error("Unable to convert to json format")
		return
	}

	// Create profile filename
	file, fileErr := os.Create(deviceProfile.Path)
	if fileErr != nil {
		logger.Log(logger.Fields{"error": err, "location": deviceProfile.Path}).Error("Unable to create new device profile")
		return
	}

	// Write JSON buffer to file
	_, err = file.Write(buffer)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "location": deviceProfile.Path}).Error("Unable to write data")
		return
	}

	// Close file
	err = file.Close()
	if err != nil {
		logger.Log(logger.Fields{"error": err, "location": deviceProfile.Path}).Fatal("Unable to close file handle")
	}

	d.loadDeviceProfiles() // Reload
}

// loadDeviceProfiles will load custom user profiles
func (d *Device) loadDeviceProfiles() {
	profileList := make(map[string]*DeviceProfile, 0)
	userProfileDirectory := pwd + "/database/profiles/"

	files, err := os.ReadDir(userProfileDirectory)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "location": userProfileDirectory, "serial": d.Serial}).Fatal("Unable to read content of a folder")
	}

	for _, fi := range files {
		pf := &DeviceProfile{}
		if fi.IsDir() {
			continue // Exclude folders if any
		}

		// Define a full path of filename
		profileLocation := userProfileDirectory + fi.Name()

		// Check if filename has .json extension
		if !common.IsValidExtension(profileLocation, ".json") {
			continue
		}

		fileName := strings.Split(fi.Name(), ".")[0]
		if m, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", fileName); !m {
			continue
		}

		fileSerial := ""
		if strings.Contains(fileName, "-") {
			fileSerial = strings.Split(fileName, "-")[0]
		} else {
			fileSerial = fileName
		}

		if fileSerial != d.Serial {
			continue
		}

		file, err := os.Open(profileLocation)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": profileLocation}).Warn("Unable to load profile")
			continue
		}
		if err = json.NewDecoder(file).Decode(pf); err != nil {
			logger.Log(logger.Fields{"error": err, "serial": d.Serial, "location": profileLocation}).Warn("Unable to decode profile")
			continue
		}
		err = file.Close()
		if err != nil {
			logger.Log(logger.Fields{"location": profileLocation, "serial": d.Serial}).Warn("Failed to close file handle")
		}

		if pf.Serial == d.Serial {
			if fileName == d.Serial {
				profileList["default"] = pf
			} else {
				name := strings.Split(fileName, "-")[1]
				profileList[name] = pf
			}
			logger.Log(logger.Fields{"location": profileLocation, "serial": d.Serial}).Info("Loaded custom user profile")
		}
	}
	d.UserProfiles = profileList
	d.getDeviceProfile()
}

// getDeviceProfile will load persistent device configuration
func (d *Device) getDeviceProfile() {
	if len(d.UserProfiles) == 0 {
		logger.Log(logger.Fields{"serial": d.Serial}).Warn("No profile found for device. Probably initial start")
	} else {
		for _, pf := range d.UserProfiles {
			if pf.Active {
				d.DeviceProfile = pf
			}
		}
	}
}

// keepAlive will keep a device alive
func (d *Device) keepAlive() {
	_, err := d.transfer([]byte{0x12}, nil)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Unable to write to a device")
	}
}

// setAutoRefresh will refresh device data
func (d *Device) setKeepAlive() {
	timerKeepAlive = time.NewTicker(time.Duration(deviceKeepAlive) * time.Millisecond)
	keepAliveChan = make(chan bool)
	go func() {
		for {
			select {
			case <-timerKeepAlive.C:
				d.keepAlive()
			case <-keepAliveChan:
				timerKeepAlive.Stop()
				return
			}
		}
	}()
}

// setAutoRefresh will refresh device data
func (d *Device) setAutoRefresh() {
	timer = time.NewTicker(time.Duration(deviceRefreshInterval) * time.Millisecond)
	authRefreshChan = make(chan bool)
	go func() {
		for {
			select {
			case <-timer.C:
				d.setTemperatures()
			case <-authRefreshChan:
				timer.Stop()
				return
			}
		}
	}()
}

// setCpuTemperature will store current CPU temperature
func (d *Device) setTemperatures() {
	d.CpuTemp = temperatures.GetCpuTemperature()
	d.GpuTemp = temperatures.GetGpuTemperature()
}

// UpdateDeviceLabel will set / update device label
func (d *Device) UpdateDeviceLabel(_ int, label string) uint8 {
	mutex.Lock()
	defer mutex.Unlock()

	d.DeviceProfile.Label = label
	d.saveDeviceProfile()
	return 1
}

// UpdateRgbProfile will update device RGB profile
func (d *Device) UpdateRgbProfile(_ int, profile string) uint8 {
	if d.GetRgbProfile(profile) == nil {
		logger.Log(logger.Fields{"serial": d.Serial, "profile": profile}).Warn("Non-existing RGB profile")
		return 0
	}
	d.DeviceProfile.RGBProfile = profile // Set profile
	d.saveDeviceProfile()                // Save profile
	if d.activeRgb != nil {
		d.activeRgb.Exit <- true // Exit current RGB mode
		d.activeRgb = nil
	}
	d.setDeviceColor() // Restart RGB
	return 1

}

// ChangeDeviceBrightness will change device brightness
func (d *Device) ChangeDeviceBrightness(mode uint8) uint8 {
	d.DeviceProfile.Brightness = mode
	d.saveDeviceProfile()
	if d.activeRgb != nil {
		d.activeRgb.Exit <- true // Exit current RGB mode
		d.activeRgb = nil
	}
	d.setDeviceColor() // Restart RGB
	return 1
}

// ChangeDeviceProfile will change device profile
func (d *Device) ChangeDeviceProfile(profileName string) uint8 {
	if profile, ok := d.UserProfiles[profileName]; ok {
		currentProfile := d.DeviceProfile
		currentProfile.Active = false
		d.DeviceProfile = currentProfile
		d.saveDeviceProfile()

		// RGB reset
		if d.activeRgb != nil {
			d.activeRgb.Exit <- true // Exit current RGB mode
			d.activeRgb = nil
		}

		newProfile := profile
		newProfile.Active = true
		d.DeviceProfile = newProfile
		d.saveDeviceProfile()
		d.setDeviceColor()
		return 1
	}
	return 0
}

// ChangeKeyboardLayout will change keyboard layout
func (d *Device) ChangeKeyboardLayout(layout string) uint8 {
	layouts := keyboards.GetLayouts(keyboardKey)
	if len(layouts) < 1 {
		return 2
	}

	if slices.Contains(layouts, layout) {
		if d.DeviceProfile != nil {
			if _, ok := d.DeviceProfile.Keyboards["default"]; ok {
				layoutKey := fmt.Sprintf("%s-%s", keyboardKey, layout)
				keyboardLayout := keyboards.GetKeyboard(layoutKey)
				if keyboardLayout == nil {
					logger.Log(logger.Fields{"serial": d.Serial}).Warn("Trying to apply non-existing keyboard layout")
					return 2
				}

				d.DeviceProfile.Keyboards["default"] = keyboardLayout
				d.DeviceProfile.Layout = layout
				d.saveDeviceProfile()
				return 1
			}
		} else {
			logger.Log(logger.Fields{"serial": d.Serial}).Warn("DeviceProfile is null")
			return 0
		}
	} else {
		logger.Log(logger.Fields{"serial": d.Serial}).Warn("No such layout")
		return 2
	}
	return 0
}

// getCurrentKeyboard will return current active keyboard
func (d *Device) getCurrentKeyboard() *keyboards.Keyboard {
	if keyboard, ok := d.DeviceProfile.Keyboards[d.DeviceProfile.Profile]; ok {
		return keyboard
	}
	return nil
}

// SaveDeviceProfile will save a new keyboard profile
func (d *Device) SaveDeviceProfile(profileName string, new bool) uint8 {
	if new {
		if d.DeviceProfile == nil {
			return 0
		}

		if slices.Contains(d.DeviceProfile.Profiles, profileName) {
			return 2
		}

		if _, ok := d.DeviceProfile.Keyboards[profileName]; ok {
			return 2
		}

		d.DeviceProfile.Profiles = append(d.DeviceProfile.Profiles, profileName)
		d.DeviceProfile.Keyboards[profileName] = d.getCurrentKeyboard()
		d.saveDeviceProfile()
		return 1
	} else {
		d.saveDeviceProfile()
		return 1
	}
}

// UpdateKeyboardProfile will change keyboard profile
func (d *Device) UpdateKeyboardProfile(profileName string) uint8 {
	if d.DeviceProfile == nil {
		return 0
	}

	if !slices.Contains(d.DeviceProfile.Profiles, profileName) {
		return 2
	}

	if _, ok := d.DeviceProfile.Keyboards[profileName]; !ok {
		return 2
	}

	d.DeviceProfile.Profile = profileName
	d.saveDeviceProfile()
	// RGB reset
	if d.activeRgb != nil {
		d.activeRgb.Exit <- true // Exit current RGB mode
		d.activeRgb = nil
	}
	d.setDeviceColor()
	return 1
}

// UpdateControlDial will update control dial function
func (d *Device) UpdateControlDial(value int) uint8 {
	d.DeviceProfile.ControlDial = value
	d.saveDeviceProfile()
	return 1
}

// DeleteKeyboardProfile will delete keyboard profile
func (d *Device) DeleteKeyboardProfile(profileName string) uint8 {
	if d.DeviceProfile == nil {
		return 0
	}

	if profileName == "default" {
		return 3
	}

	if !slices.Contains(d.DeviceProfile.Profiles, profileName) {
		return 2
	}

	if _, ok := d.DeviceProfile.Keyboards[profileName]; !ok {
		return 2
	}

	index := common.IndexOfString(d.DeviceProfile.Profiles, profileName)
	if index < 0 {
		return 0
	}

	d.DeviceProfile.Profile = "default"
	d.DeviceProfile.Profiles = append(d.DeviceProfile.Profiles[:index], d.DeviceProfile.Profiles[index+1:]...)
	delete(d.DeviceProfile.Keyboards, profileName)

	d.saveDeviceProfile()
	// RGB reset
	if d.activeRgb != nil {
		d.activeRgb.Exit <- true // Exit current RGB mode
		d.activeRgb = nil
	}
	d.setDeviceColor()
	return 1
}

// SaveUserProfile will generate a new user profile configuration and save it to a file
func (d *Device) SaveUserProfile(profileName string) uint8 {
	if d.DeviceProfile != nil {
		profilePath := pwd + "/database/profiles/" + d.Serial + "-" + profileName + ".json"

		newProfile := d.DeviceProfile
		newProfile.Path = profilePath
		newProfile.Active = false

		buffer, err := json.Marshal(newProfile)
		if err != nil {
			logger.Log(logger.Fields{"error": err}).Error("Unable to convert to json format")
			return 0
		}

		// Create profile filename
		file, err := os.Create(profilePath)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "location": newProfile.Path}).Error("Unable to create new device profile")
			return 0
		}

		_, err = file.Write(buffer)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "location": newProfile.Path}).Error("Unable to write data")
			return 0
		}

		err = file.Close()
		if err != nil {
			logger.Log(logger.Fields{"error": err, "location": newProfile.Path}).Error("Unable to close file handle")
			return 0
		}
		d.loadDeviceProfiles()
		return 1
	}
	return 0
}

// UpdateDeviceColor will update device color based on selected input
func (d *Device) UpdateDeviceColor(keyId, keyOption int, color rgb.Color) uint8 {
	switch keyOption {
	case 0:
		{
			for rowIndex, row := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
				for keyIndex, key := range row.Keys {
					if keyIndex == keyId {
						key.Color = rgb.Color{
							Red:        color.Red,
							Green:      color.Green,
							Blue:       color.Blue,
							Brightness: 0,
						}
						d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row[rowIndex].Keys[keyIndex] = key
						if d.activeRgb != nil {
							d.activeRgb.Exit <- true // Exit current RGB mode
							d.activeRgb = nil
						}
						d.setDeviceColor() // Restart RGB
						return 1
					}
				}
			}
		}
	case 1:
		{
			rowId := -1
			for rowIndex, row := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
				for keyIndex := range row.Keys {
					if keyIndex == keyId {
						rowId = rowIndex
						break
					}
				}
			}

			if rowId < 0 {
				return 0
			}

			for keyIndex, key := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row[rowId].Keys {
				key.Color = rgb.Color{
					Red:        color.Red,
					Green:      color.Green,
					Blue:       color.Blue,
					Brightness: 0,
				}
				d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row[rowId].Keys[keyIndex] = key
			}
			if d.activeRgb != nil {
				d.activeRgb.Exit <- true // Exit current RGB mode
				d.activeRgb = nil
			}
			d.setDeviceColor() // Restart RGB
			return 1
		}
	case 2:
		{
			for rowIndex, row := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
				for keyIndex, key := range row.Keys {
					key.Color = rgb.Color{
						Red:        color.Red,
						Green:      color.Green,
						Blue:       color.Blue,
						Brightness: 0,
					}
					d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row[rowIndex].Keys[keyIndex] = key
				}
			}
			if d.activeRgb != nil {
				d.activeRgb.Exit <- true // Exit current RGB mode
				d.activeRgb = nil
			}
			d.setDeviceColor() // Restart RGB
			return 1
		}
	}
	return 0
}

// setBrightnessLevel will set global brightness level
func (d *Device) setBrightnessLevel() {
	if d.DeviceProfile != nil {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf[0:2], d.DeviceProfile.BrightnessLevel)
		for i := 0; i <= 1; i++ {
			_, err := d.transfer(cmdBrightness[i], buf)
			if err != nil {
				logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Warn("Unable to change brightness")
			}
		}
	}
}

// setDeviceColor will activate and set device RGB
func (d *Device) setDeviceColor() {
	// Reset
	var buf = make([]byte, colorPacketLength)

	for i := 0; i < colorPacketLength; i++ {
		buf[i] = 0xff
	}
	d.writeColor(buf)
	if d.DeviceProfile == nil {
		logger.Log(logger.Fields{"serial": d.Serial}).Error("Unable to set color. DeviceProfile is null!")
		return
	}

	if d.DeviceProfile.Keyboards == nil {
		logger.Log(logger.Fields{"serial": d.Serial}).Error("Unable to set color. DeviceProfile Keyboards is null!")
		return
	}

	if d.DeviceProfile.RGBProfile == "keyboard" {
		if _, ok := d.DeviceProfile.Keyboards[d.DeviceProfile.Profile]; ok {
			for _, rows := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
				for _, keys := range rows.Keys {
					for _, packetIndex := range keys.PacketIndex {
						buf[packetIndex] = byte(keys.Color.Red)
						buf[packetIndex+1] = byte(keys.Color.Green)
						buf[packetIndex+2] = byte(keys.Color.Blue)
					}
				}
			}
			d.writeColor(buf) // Write color once
			return
		} else {
			logger.Log(logger.Fields{"serial": d.Serial}).Error("Unable to set color. Unknown keyboard")
			return
		}
	}

	if d.DeviceProfile.RGBProfile == "static" {
		profile := d.GetRgbProfile("static")
		if d.DeviceProfile.Brightness != 0 {
			profile.StartColor.Brightness = rgb.GetBrightnessValue(d.DeviceProfile.Brightness)
		}

		profileColor := rgb.ModifyBrightness(profile.StartColor)
		for _, rows := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
			for _, keys := range rows.Keys {
				for _, packetIndex := range keys.PacketIndex {
					buf[packetIndex] = byte(profileColor.Red)
					buf[packetIndex+1] = byte(profileColor.Green)
					buf[packetIndex+2] = byte(profileColor.Blue)
				}
			}
		}
		d.writeColor(buf) // Write color once
		return
	}

	go func(lightChannels int) {
		lock := sync.Mutex{}
		startTime := time.Now()
		reverse := false
		counterColorpulse := 0
		counterFlickering := 0
		counterColorshift := 0
		counterCircleshift := 0
		counterCircle := 0
		counterColorwarp := 0
		counterSpinner := 0
		counterCpuTemp := 0
		counterGpuTemp := 0
		var temperatureKeys *rgb.Color
		colorwarpGeneratedReverse := false
		d.activeRgb = rgb.Exit()

		// Generate random colors
		d.activeRgb.RGBStartColor = rgb.GenerateRandomColor(1)
		d.activeRgb.RGBEndColor = rgb.GenerateRandomColor(1)

		hue := 1
		wavePosition := 0.0
		for {
			select {
			case <-d.activeRgb.Exit:
				return
			default:
				buff := make([]byte, 0)

				rgbCustomColor := true
				profile := d.GetRgbProfile(d.DeviceProfile.RGBProfile)
				if profile == nil {
					for i := 0; i < d.LEDChannels; i++ {
						buff = append(buff, []byte{0, 0, 0}...)
					}
					logger.Log(logger.Fields{"profile": d.DeviceProfile.RGBProfile, "serial": d.Serial}).Warn("No such RGB profile found")
					continue
				}
				rgbModeSpeed := common.FClamp(profile.Speed, 0.1, 10)
				// Check if we have custom colors
				if (rgb.Color{}) == profile.StartColor || (rgb.Color{}) == profile.EndColor {
					rgbCustomColor = false
				}

				r := rgb.New(
					d.LEDChannels,
					rgbModeSpeed,
					nil,
					nil,
					profile.Brightness,
					common.Clamp(profile.Smoothness, 1, 100),
					time.Duration(rgbModeSpeed)*time.Second,
					rgbCustomColor,
				)

				if rgbCustomColor {
					r.RGBStartColor = &profile.StartColor
					r.RGBEndColor = &profile.EndColor
				} else {
					r.RGBStartColor = d.activeRgb.RGBStartColor
					r.RGBEndColor = d.activeRgb.RGBEndColor
				}

				// Brightness
				if d.DeviceProfile.Brightness > 0 {
					r.RGBBrightness = rgb.GetBrightnessValue(d.DeviceProfile.Brightness)
					r.RGBStartColor.Brightness = r.RGBBrightness
					r.RGBEndColor.Brightness = r.RGBBrightness
				}

				switch d.DeviceProfile.RGBProfile {
				case "off":
					{
						for n := 0; n < d.LEDChannels; n++ {
							buff = append(buff, []byte{0, 0, 0}...)
						}
					}
				case "rainbow":
					{
						r.Rainbow(startTime)
						buff = append(buff, r.Output...)
					}
				case "watercolor":
					{
						r.Watercolor(startTime)
						buff = append(buff, r.Output...)
					}
				case "cpu-temperature":
					{
						lock.Lock()
						counterCpuTemp++
						if counterCpuTemp >= r.Smoothness {
							counterCpuTemp = 0
						}

						if temperatureKeys == nil {
							temperatureKeys = r.RGBStartColor
						}

						r.MinTemp = profile.MinTemp
						r.MaxTemp = profile.MaxTemp
						res := r.Temperature(float64(d.CpuTemp), counterCpuTemp, temperatureKeys)
						temperatureKeys = res
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "gpu-temperature":
					{
						lock.Lock()
						counterGpuTemp++
						if counterGpuTemp >= r.Smoothness {
							counterGpuTemp = 0
						}

						if temperatureKeys == nil {
							temperatureKeys = r.RGBStartColor
						}

						r.MinTemp = profile.MinTemp
						r.MaxTemp = profile.MaxTemp
						res := r.Temperature(float64(d.GpuTemp), counterGpuTemp, temperatureKeys)
						temperatureKeys = res
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "colorpulse":
					{
						lock.Lock()
						counterColorpulse++
						if counterColorpulse >= r.Smoothness {
							counterColorpulse = 0
						}

						r.Colorpulse(counterColorpulse)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "static":
					{
						r.Static()
						buff = append(buff, r.Output...)
					}
				case "rotator":
					{
						r.Rotator(hue)
						buff = append(buff, r.Output...)
					}
				case "wave":
					{
						r.Wave(wavePosition)
						buff = append(buff, r.Output...)
					}
				case "storm":
					{
						r.Storm()
						buff = append(buff, r.Output...)
					}
				case "flickering":
					{
						lock.Lock()
						if counterFlickering >= r.Smoothness {
							counterFlickering = 0
						} else {
							counterFlickering++
						}

						r.Flickering(counterFlickering)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "colorshift":
					{
						lock.Lock()
						if counterColorshift >= r.Smoothness && !reverse {
							counterColorshift = 0
							reverse = true
						} else if counterColorshift >= r.Smoothness && reverse {
							counterColorshift = 0
							reverse = false
						}

						r.Colorshift(counterColorshift, reverse)
						counterColorshift++
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "circleshift":
					{
						lock.Lock()
						if counterCircleshift >= lightChannels {
							counterCircleshift = 0
						} else {
							counterCircleshift++
						}

						r.Circle(counterCircleshift)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "circle":
					{
						lock.Lock()
						if counterCircle >= lightChannels {
							counterCircle = 0
						} else {
							counterCircle++
						}

						r.Circle(counterCircle)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "spinner":
					{
						lock.Lock()
						if counterSpinner >= lightChannels {
							counterSpinner = 0
						} else {
							counterSpinner++
						}
						r.Spinner(counterSpinner)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				case "colorwarp":
					{
						lock.Lock()
						if counterColorwarp >= r.Smoothness {
							if !colorwarpGeneratedReverse {
								colorwarpGeneratedReverse = true
								d.activeRgb.RGBStartColor = d.activeRgb.RGBEndColor
								d.activeRgb.RGBEndColor = rgb.GenerateRandomColor(r.RGBBrightness)
							}
							counterColorwarp = 0
						} else if counterColorwarp == 0 && colorwarpGeneratedReverse == true {
							colorwarpGeneratedReverse = false
						} else {
							counterColorwarp++
						}

						r.Colorwarp(counterColorwarp, d.activeRgb.RGBStartColor, d.activeRgb.RGBEndColor)
						lock.Unlock()
						buff = append(buff, r.Output...)
					}
				}

				for _, rows := range d.DeviceProfile.Keyboards[d.DeviceProfile.Profile].Row {
					for _, keys := range rows.Keys {
						for _, packetIndex := range keys.PacketIndex {
							buf[packetIndex] = buff[packetIndex]
							buf[packetIndex+1] = buff[packetIndex+1]
							buf[packetIndex+2] = buff[packetIndex+2]
						}
					}
				}

				// Send it
				d.writeColor(buf)
				time.Sleep(20 * time.Millisecond)
				hue++
				wavePosition += 0.2
			}
		}
	}(d.LEDChannels)
}

// writeColor will write data to the device with a specific endpoint.
// writeColor does not require endpoint closing and opening like normal Write requires.
// Endpoint is open only once. Once the endpoint is open, color can be sent continuously.
func (d *Device) writeColor(data []byte) {
	buffer := make([]byte, len(dataTypeSetColor)+len(data)+headerWriteSize)
	binary.LittleEndian.PutUint16(buffer[0:2], uint16(len(data)))
	copy(buffer[headerWriteSize:headerWriteSize+len(dataTypeSetColor)], dataTypeSetColor)
	copy(buffer[headerWriteSize+len(dataTypeSetColor):], data)

	// Split packet into chunks
	chunks := common.ProcessMultiChunkPacket(buffer, maxBufferSizePerRequest)
	for i, chunk := range chunks {
		if i == 0 {
			// Initial packet is using cmdWriteColor
			_, err := d.transfer(cmdWriteColor, chunk)
			if err != nil {
				logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Unable to write to color endpoint")
			}
		} else {
			// Chunks don't use cmdWriteColor, they use static dataTypeSubColor
			_, err := d.transfer(dataTypeSubColor, chunk)
			if err != nil {
				logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Unable to write to endpoint")
			}
		}
	}
}

// transfer will send data to a device and retrieve device output
func (d *Device) transfer(endpoint, buffer []byte) ([]byte, error) {
	// Packet control, mandatory for this device
	mutex.Lock()
	defer mutex.Unlock()

	// Create write buffer
	bufferW := make([]byte, bufferSizeWrite)
	bufferW[1] = 0x08
	endpointHeaderPosition := bufferW[headerSize : headerSize+len(endpoint)]
	copy(endpointHeaderPosition, endpoint)
	if len(buffer) > 0 {
		copy(bufferW[headerSize+len(endpoint):headerSize+len(endpoint)+len(buffer)], buffer)
	}

	// Create read buffer
	bufferR := make([]byte, bufferSize)

	// Send command to a device
	if _, err := d.dev.Write(bufferW); err != nil {
		logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Unable to write to a device")
		return nil, err
	}

	// Get data from a device
	if _, err := d.dev.Read(bufferR); err != nil {
		logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Unable to read data from device")
		return nil, err
	}

	return bufferR, nil
}

// controlListener will listen for events from the control buttons
func (d *Device) controlListener() {
	var brightness uint16 = 0

	if d.DeviceProfile.BrightnessLevel == 0 {
		brightness = 1000
	} else {
		brightness = d.DeviceProfile.BrightnessLevel
	}

	go func() {
		enum := hid.EnumFunc(func(info *hid.DeviceInfo) error {
			if info.InterfaceNbr == 2 {
				listener, err := hid.OpenPath(info.Path)
				if err != nil {
					return err
				}
				d.listener = listener
			}
			return nil
		})

		err := hid.Enumerate(d.VendorId, d.ProductId, enum)
		if err != nil {
			logger.Log(logger.Fields{"error": err, "vendorId": d.VendorId}).Error("Unable to enumerate devices")
		}

		// Listen loop
		data := make([]byte, bufferSize)
		for {
			// Read data from the HID device
			_, err = d.listener.Read(data)
			if err != nil {
				logger.Log(logger.Fields{"error": err, "serial": d.Serial}).Error("Error reading data")
				break
			}

			switch data[1] {
			case 0x05: // // Right horizontal spinning wheel
				{
					switch data[4] {
					case 0x01: // Up direction
						{
							switch d.DeviceProfile.ControlDial {
							case 1:
								if brightness >= 1000 {
									brightness = 1000
								} else {
									brightness += 100
								}

								if d.DeviceProfile != nil {
									d.DeviceProfile.BrightnessLevel = brightness
									d.saveDeviceProfile()
									d.setBrightnessLevel()
								}
								break
							}
						}
						break
					case 0xff: // Down direction
						{
							switch d.DeviceProfile.ControlDial {
							case 1:
								if brightness <= 0 {
									brightness = 0
								} else {
									brightness -= 100
								}

								if d.DeviceProfile != nil {
									d.DeviceProfile.BrightnessLevel = brightness
									d.saveDeviceProfile()
									d.setBrightnessLevel()
								}
								break
							}
						}
					}
				}
			case 0x02: // Media, logo keys, etc...
				{
					if data[19] == 0x02 { // Disable / enable lighting, iCUE button
						if brightness > 0 {
							brightness = 0
						} else {
							brightness = 1000
						}
						if d.DeviceProfile != nil {
							d.DeviceProfile.BrightnessLevel = brightness
							d.saveDeviceProfile()
							d.setBrightnessLevel()
						}
					} else if data[14] == 0x40 { // Mute
						inputmanager.InputControl(inputmanager.VolumeMute, d.Serial)
					} else if data[15] == 0x01 { // Volume down
						inputmanager.InputControl(inputmanager.VolumeDown, d.Serial)
					} else if data[14] == 0x80 { // Volume up
						inputmanager.InputControl(inputmanager.VolumeUp, d.Serial)
					} else if data[17] == 0x08 { // Media Stop
						inputmanager.InputControl(inputmanager.MediaStop, d.Serial)
					} else if data[17] == 0x40 { // Media Previous
						inputmanager.InputControl(inputmanager.MediaPrev, d.Serial)
					} else if data[17] == 0x10 { // Media Play / Pause
						inputmanager.InputControl(inputmanager.MediaPlayPause, d.Serial)
					} else if data[17] == 0x20 { // Media Next
						inputmanager.InputControl(inputmanager.MediaNext, d.Serial)
					}
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}