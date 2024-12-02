package devices

import (
	"OpenLinkHub/src/config"
	"OpenLinkHub/src/devices/cc"
	"OpenLinkHub/src/devices/ccxt"
	"OpenLinkHub/src/devices/cpro"
	"OpenLinkHub/src/devices/elite"
	"OpenLinkHub/src/devices/k100air"
	"OpenLinkHub/src/devices/k100airW"
	"OpenLinkHub/src/devices/k55core"
	"OpenLinkHub/src/devices/k65plus"
	"OpenLinkHub/src/devices/k65plusW"
	"OpenLinkHub/src/devices/k65pm"
	"OpenLinkHub/src/devices/k70core"
	"OpenLinkHub/src/devices/k70pro"
	"OpenLinkHub/src/devices/lncore"
	"OpenLinkHub/src/devices/lnpro"
	"OpenLinkHub/src/devices/lsh"
	"OpenLinkHub/src/devices/memory"
	"OpenLinkHub/src/devices/xc7"
	"OpenLinkHub/src/logger"
	"OpenLinkHub/src/metrics"
	"OpenLinkHub/src/rgb"
	"OpenLinkHub/src/smbus"
	"github.com/sstallion/go-hid"
	"slices"
	"strconv"
)

const (
	productTypeLinkHub  = 0
	productTypeCC       = 1
	productTypeCCXT     = 2
	productTypeElite    = 3
	productTypeLNCore   = 4
	productTypeLnPro    = 5
	productTypeCPro     = 6
	productTypeXC7      = 7
	productTypeMemory   = 8
	productTypeK65PM    = 101
	productTypeK70Core  = 102
	productTypeK55Core  = 103
	productTypeK70Pro   = 104
	productTypeK65Plus  = 105
	productTypeK65PlusW = 106
	productTypeK100Air  = 107
	productTypeK100AirW = 108
	productTypeST100    = 401
	productTypeMM700    = 402
	productTypeLT100    = 403
)

type AIOData struct {
	Rpm         int16
	Temperature float32
	Serial      string
}

type Device struct {
	ProductType uint8
	Product     string
	Serial      string
	Firmware    string
	Lsh         *lsh.Device      `json:"lsh,omitempty"`
	CC          *cc.Device       `json:"cc,omitempty"`
	CCXT        *ccxt.Device     `json:"ccxt,omitempty"`
	Elite       *elite.Device    `json:"elite,omitempty"`
	LnCore      *lncore.Device   `json:"lncore,omitempty"`
	LnPro       *lnpro.Device    `json:"lnpro,omitempty"`
	CPro        *cpro.Device     `json:"cPro,omitempty"`
	XC7         *xc7.Device      `json:"xc7,omitempty"`
	Memory      *memory.Device   `json:"memory,omitempty"`
	K65PM       *k65pm.Device    `json:"k65PM,omitempty"`
	K70Core     *k70core.Device  `json:"k70core,omitempty"`
	K55Core     *k55core.Device  `json:"k55core,omitempty"`
	K70Pro      *k70pro.Device   `json:"k70pro,omitempty"`
	K65Plus     *k65plus.Device  `json:"k65plus,omitempty"`
	K65PlusW    *k65plusW.Device `json:"k65plusW,omitempty"`
	K100Air     *k100air.Device  `json:"k100air,omitempty"`
	K100AirW    *k100airW.Device `json:"k100airW,omitempty"`
	GetDevice   interface{}
}

var (
	vendorId    uint16 = 6940 // Corsair
	interfaceId        = 0
	devices            = make(map[string]*Device, 0)
	products           = make(map[string]uint16)
	keyboards          = []uint16{7127, 7165, 7166, 7110, 7083, 7132, 11024, 11015}
)

// Stop will stop all active devices
func Stop() {
	for _, device := range devices {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					device.Lsh.Stop()
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					device.CC.Stop()
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					device.CCXT.Stop()
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					device.Elite.Stop()
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					device.LnCore.Stop()
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					device.LnPro.Stop()
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					device.CPro.Stop()
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					device.XC7.Stop()
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					device.Memory.Stop()
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					device.K65PM.Stop()
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					device.K70Core.Stop()
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					device.K55Core.Stop()
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					device.K70Pro.Stop()
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					device.K65Plus.Stop()
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					device.K65PlusW.Stop()
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					device.K100Air.Stop()
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					device.K100AirW.Stop()
				}
			}
		}
	}
	err := hid.Exit()
	if err != nil {
		logger.Log(logger.Fields{"error": err}).Error("Unable to exit HID interface")
	}
}

// UpdateKeyboardColor will process POST request from a client for keyboard color change
func UpdateKeyboardColor(deviceId string, keyId, keyOptions int, color rgb.Color) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateDeviceColor(keyOptions, color)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.UpdateDeviceColor(keyId, keyOptions, color)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.UpdateDeviceColor(keyOptions, color)
				}
			}
		}
	}
	return 0
}

// UpdateARGBDevice will process POST request from a client for ARGB 3-pin devices
func UpdateARGBDevice(deviceId string, portId, deviceType int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateARGBDevice(portId, deviceType)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateARGBDevice(portId, deviceType)
				}
			}
		}
	}
	return 0
}

// UpdateExternalHubDeviceType will update a device type connected to an external-LED hub
func UpdateExternalHubDeviceType(deviceId string, portId, deviceType int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateExternalHubDeviceType(deviceType)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.UpdateExternalHubDeviceType(deviceType)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.UpdateExternalHubDeviceType(portId, deviceType)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateExternalHubDeviceType(portId, deviceType)
				}
			}
		}
	}
	return 0
}

// UpdateExternalHubDeviceAmount will update a device amount connected to an external-LED hub
func UpdateExternalHubDeviceAmount(deviceId string, portId, deviceType int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateExternalHubDeviceAmount(deviceType)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.UpdateExternalHubDeviceAmount(deviceType)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.UpdateExternalHubDeviceAmount(portId, deviceType)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateExternalHubDeviceAmount(portId, deviceType)
				}
			}
		}
	}
	return 0
}

// UpdateDeviceMetrics will update device metrics
func UpdateDeviceMetrics() {
	// Default
	metrics.PopulateDefault()

	// Storage
	metrics.PopulateStorage()

	// Devices
	for _, device := range devices {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					device.Lsh.UpdateDeviceMetrics()
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					device.CC.UpdateDeviceMetrics()
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					device.Elite.UpdateDeviceMetrics()
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					device.CPro.UpdateDeviceMetrics()
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					device.CCXT.UpdateDeviceMetrics()
				}
			}
		}
	}
}

// SaveKeyboardProfile will save keyboard profile
func SaveKeyboardProfile(deviceId, profileName string, new bool) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.SaveKeyboardProfile(profileName, new)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.SaveKeyboardProfile(profileName, new)
				}
			}
		}
	}
	return 0
}

// ChangeKeyboardLayout will change keyboard layout
func ChangeKeyboardLayout(deviceId, layout string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.ChangeKeyboardLayout(layout)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.ChangeKeyboardLayout(layout)
				}
			}
		}
	}
	return 0
}

// ChangeKeyboardControlDial will change keyboard control dial function
func ChangeKeyboardControlDial(deviceId string, controlDial int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.UpdateControlDial(controlDial)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateControlDial(controlDial)
				}
			}
		}
	}
	return 0
}

// ChangeKeyboardSleepMode will change keyboard control dial function
func ChangeKeyboardSleepMode(deviceId string, sleepMode int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateSleepTimer(sleepMode)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.UpdateSleepTimer(sleepMode)
				}
			}
		}
	}
	return 0
}

// ChangeKeyboardProfile will change keyboard profile
func ChangeKeyboardProfile(deviceId, profileName string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.UpdateKeyboardProfile(profileName)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.UpdateKeyboardProfile(profileName)
				}
			}
		}
	}
	return 0
}

// DeleteKeyboardProfile will save keyboard profile
func DeleteKeyboardProfile(deviceId, profileName string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.DeleteKeyboardProfile(profileName)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.DeleteKeyboardProfile(profileName)
				}
			}
		}
	}
	return 0
}

// SaveUserProfile will save new device user profile
func SaveUserProfile(deviceId, profileName string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.SaveUserProfile(profileName)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.SaveUserProfile(profileName)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.SaveUserProfile(profileName)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.SaveUserProfile(profileName)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.SaveUserProfile(profileName)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.SaveUserProfile(profileName)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.SaveUserProfile(profileName)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.SaveUserProfile(profileName)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory.SaveUserProfile(profileName)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.SaveUserProfile(profileName)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.SaveUserProfile(profileName)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.SaveUserProfile(profileName)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.SaveUserProfile(profileName)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.SaveUserProfile(profileName)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.SaveUserProfile(profileName)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.SaveUserProfile(profileName)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.SaveUserProfile(profileName)
				}
			}
		}
	}
	return 0
}

// UpdateDevicePosition will change device position
func UpdateDevicePosition(deviceId string, position, direction int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateDevicePosition(position, direction)
				}
			}
		}
	}
	return 0
}

// ScheduleDeviceBrightness will change device brightness level based on scheduler
func ScheduleDeviceBrightness(mode uint8) {
	for _, device := range GetDevices() {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					device.Lsh.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					device.CCXT.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					device.CC.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					device.CPro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					device.LnPro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					device.LnCore.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					device.Elite.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					device.XC7.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					device.Memory.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					device.K65PM.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					device.K70Core.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					device.K70Pro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					device.K55Core.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					device.K65Plus.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					device.K65PlusW.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					device.K100Air.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					device.K100AirW.ChangeDeviceBrightness(mode)
				}
			}
		}
	}
}

// ChangeDeviceBrightness will change device brightness level
func ChangeDeviceBrightness(deviceId string, mode uint8) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.ChangeDeviceBrightness(mode)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.ChangeDeviceBrightness(mode)
				}
			}
		}
	}
	return 0
}

// ChangeUserProfile will change device user profile
func ChangeUserProfile(deviceId, profileName string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.ChangeDeviceProfile(profileName)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.ChangeDeviceProfile(profileName)
				}
			}
		}
	}
	return 0
}

// UpdateDeviceLcd will update device LCD
func UpdateDeviceLcd(deviceId string, channelId int, mode uint8) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateDeviceLcd(mode)
				}
			}
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateDeviceLcd(channelId, mode)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.UpdateDeviceLcd(mode)
				}
			}
		}
	}
	return 0
}

// ChangeDeviceLcd will change device LCD
func ChangeDeviceLcd(deviceId string, channelId int, lcdSerial string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.ChangeDeviceLcd(channelId, lcdSerial)
				}
			}
		}
	}
	return 0
}

// UpdateDeviceLcdRotation will update device LCD rotation
func UpdateDeviceLcdRotation(deviceId string, channelId int, rotation uint8) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateDeviceLcdRotation(rotation)
				}
			}
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateDeviceLcdRotation(channelId, rotation)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.UpdateDeviceLcdRotation(rotation)
				}
			}
		}
	}
	return 0
}

// UpdateDeviceLabel will set / update device label
func UpdateDeviceLabel(deviceId string, channelId int, label string, deviceType int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					if deviceType == 0 {
						return device.CC.UpdateDeviceLabel(channelId, label)
					} else {
						return device.CC.UpdateRGBDeviceLabel(channelId, label)
					}
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					if deviceType == 0 {
						return device.CCXT.UpdateDeviceLabel(channelId, label)
					} else {
						return device.CCXT.UpdateRGBDeviceLabel(channelId, label)
					}
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.UpdateDeviceLabel(label)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory.UpdateDeviceLabel(channelId, label)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.UpdateDeviceLabel(label)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.UpdateDeviceLabel(label)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.UpdateDeviceLabel(label)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.UpdateDeviceLabel(label)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.UpdateDeviceLabel(label)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateDeviceLabel(label)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.UpdateDeviceLabel(label)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.UpdateDeviceLabel(label)
				}
			}
		}
	}
	return 0
}

// UpdateSpeedProfile will update device speeds with a given serial number
func UpdateSpeedProfile(deviceId string, channelId int, profile string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateSpeedProfile(channelId, profile)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateSpeedProfile(channelId, profile)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateSpeedProfile(channelId, profile)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.UpdateSpeedProfile(channelId, profile)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateSpeedProfile(channelId, profile)
				}
			}
		}
	}
	return 0
}

// UpdateManualSpeed will update device speeds with a given serial number
func UpdateManualSpeed(deviceId string, channelId int, value uint16) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateDeviceSpeed(channelId, value)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateDeviceSpeed(channelId, value)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateDeviceSpeed(channelId, value)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.UpdateDeviceSpeed(channelId, value)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateDeviceSpeed(channelId, value)
				}
			}
		}
	}
	return 0
}

// UpdateRgbStrip will update device RGB strip
func UpdateRgbStrip(deviceId string, channelId int, stripId int) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateExternalAdapter(channelId, stripId)
				}
			}
		}
	}
	return 0
}

// UpdateRgbProfile will update device RGB profile
func UpdateRgbProfile(deviceId string, channelId int, profile string) uint8 {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7.UpdateRgbProfile(profile)
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory.UpdateRgbProfile(channelId, profile)
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM.UpdateRgbProfile(profile)
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core.UpdateRgbProfile(profile)
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro.UpdateRgbProfile(profile)
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core.UpdateRgbProfile(profile)
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus.UpdateRgbProfile(profile)
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW.UpdateRgbProfile(profile)
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air.UpdateRgbProfile(profile)
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW.UpdateRgbProfile(profile)
				}
			}
		}
	}
	return 0
}

// ResetSpeedProfiles will reset the speed profile on each available device
func ResetSpeedProfiles(profile string) {
	for _, device := range devices {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					device.Lsh.ResetSpeedProfiles(profile)
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					device.CC.ResetSpeedProfiles(profile)
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					device.CCXT.ResetSpeedProfiles(profile)
				}
			}
		}
	}
}

// GetDevices will return all available devices
func GetDevices() map[string]*Device {
	return devices
}

// GetTemperatureProbes will return a list of temperature probes
func GetTemperatureProbes() interface{} {
	var probes []interface{}
	for _, device := range devices {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					probes = append(probes, device.Lsh.GetTemperatureProbes())
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					probes = append(probes, device.CC.GetTemperatureProbes())
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					probes = append(probes, device.CCXT.GetTemperatureProbes())
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					probes = append(probes, device.CPro.GetTemperatureProbes())
				}
			}
		}
	}
	return probes
}

// GetDevice will return a device by device serial
func GetDevice(deviceId string) interface{} {
	if device, ok := devices[deviceId]; ok {
		switch device.ProductType {
		case productTypeLinkHub:
			{
				if device.Lsh != nil {
					return device.Lsh
				}
			}
		case productTypeCC:
			{
				if device.CC != nil {
					return device.CC
				}
			}
		case productTypeCCXT:
			{
				if device.CCXT != nil {
					return device.CCXT
				}
			}
		case productTypeElite:
			{
				if device.Elite != nil {
					return device.Elite
				}
			}
		case productTypeLNCore:
			{
				if device.LnCore != nil {
					return device.LnCore
				}
			}
		case productTypeLnPro:
			{
				if device.LnPro != nil {
					return device.LnPro
				}
			}
		case productTypeCPro:
			{
				if device.CPro != nil {
					return device.CPro
				}
			}
		case productTypeXC7:
			{
				if device.XC7 != nil {
					return device.XC7
				}
			}
		case productTypeMemory:
			{
				if device.Memory != nil {
					return device.Memory
				}
			}
		case productTypeK65PM:
			{
				if device.K65PM != nil {
					return device.K65PM
				}
			}
		case productTypeK70Core:
			{
				if device.K70Core != nil {
					return device.K70Core
				}
			}
		case productTypeK70Pro:
			{
				if device.K70Pro != nil {
					return device.K70Pro
				}
			}
		case productTypeK55Core:
			{
				if device.K55Core != nil {
					return device.K55Core
				}
			}
		case productTypeK65Plus:
			{
				if device.K65Plus != nil {
					return device.K65Plus
				}
			}
		case productTypeK65PlusW:
			{
				if device.K65PlusW != nil {
					return device.K65PlusW
				}
			}
		case productTypeK100Air:
			{
				if device.K100Air != nil {
					return device.K100Air
				}
			}
		case productTypeK100AirW:
			{
				if device.K100AirW != nil {
					return device.K100AirW
				}
			}
		}
	}
	return nil
}

// Init will initialize all compatible Corsair devices in your system
func Init() {
	// Initialize general HID interface
	if err := hid.Init(); err != nil {
		logger.Log(logger.Fields{"error": err}).Fatal("Unable to initialize HID interface")
	}

	enum := hid.EnumFunc(func(info *hid.DeviceInfo) error {
		keyboard := false
		if slices.Contains(keyboards, info.ProductID) {
			interfaceId = 1 // Keyboards
			keyboard = true
		} else {
			interfaceId = 0
		}
		if info.InterfaceNbr == interfaceId {
			if keyboard {
				products[info.Path] = info.ProductID
			} else {
				products[info.SerialNbr] = info.ProductID
			}
		}
		return nil
	})

	// Enumerate all Corsair devices
	err := hid.Enumerate(vendorId, hid.ProductIDAny, enum)
	if err != nil {
		logger.Log(logger.Fields{"error": err, "vendorId": vendorId}).Fatal("Unable to enumerate devices")
	}

	if config.GetConfig().Memory {
		sm, err := smbus.GetSmBus()
		if err == nil {
			if len(sm.Path) > 0 {
				products[sm.Path] = 0
			}
		} else {
			logger.Log(logger.Fields{"error": err}).Warn("No valid I2C devices found")
		}
	}

	// USB-HID
	for key, productId := range products {
		if slices.Contains(config.GetConfig().Exclude, productId) {
			logger.Log(logger.Fields{"productId": productId}).Warn("Product excluded via config.json")
			continue
		}

		switch productId {
		case 3135: // CORSAIR iCUE Link System Hub
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := lsh.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						Lsh:         dev,
						ProductType: productTypeLinkHub,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId, key)
			}
		case 3122, 3100: // CORSAIR iCUE COMMANDER Core
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := cc.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						CC:          dev,
						ProductType: productTypeCC,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId, key)
			}
		case 3114: // CORSAIR iCUE COMMANDER CORE XT
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := ccxt.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						CCXT:        dev,
						ProductType: productTypeCCXT,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId, key)
			}
		case 3104, 3105, 3106, 3125, 3126, 3127, 3136, 3137:
			// iCUE H100i ELITE RGB
			// iCUE H115i ELITE RGB
			// iCUE H150i ELITE RGB
			// iCUE H100i ELITE RGB White
			// iCUE H150i ELITE RGB White
			// iCUE H100i RGB PRO XT
			// iCUE H115i RGB PRO XT
			// iCUE H150i RGB PRO XT
			{
				go func(vendorId, productId uint16) {
					dev := elite.Init(vendorId, productId)
					if dev == nil {
						return
					}
					devices[strconv.Itoa(int(productId))] = &Device{
						Elite:       dev,
						ProductType: productTypeElite,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId)
			}
		case 3098: // CORSAIR Lighting Node CORE
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := lncore.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						LnCore:      dev,
						ProductType: productTypeLNCore,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 3083: // CORSAIR Lighting Node Pro
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := lnpro.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						LnPro:       dev,
						ProductType: productTypeLnPro,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 3088: // Corsair Commander Pro
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := cpro.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						CPro:        dev,
						ProductType: productTypeCPro,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId, key)
			}
		case 3138: // CORSAIR XC7 ELITE LCD CPU Water Block
			{
				go func(vendorId, productId uint16, serialId string) {
					dev := xc7.Init(vendorId, productId, serialId)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						XC7:         dev,
						ProductType: productTypeXC7,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
					devices[dev.Serial].GetDevice = GetDevice(dev.Serial)
				}(vendorId, productId, key)
			}
		case 7127: // K65 Pro Mini
			{
				go func(vendorId, productId uint16, key string) {
					dev := k65pm.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K65PM:       dev,
						ProductType: productTypeK65PM,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 7165: // K70 CORE RGB
			{
				go func(vendorId, productId uint16, key string) {
					dev := k70core.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K70Core:     dev,
						ProductType: productTypeK70Core,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 7166: // K55 CORE RGB
			{
				go func(vendorId, productId uint16, key string) {
					dev := k55core.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K55Core:     dev,
						ProductType: productTypeK55Core,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 7110: // K70 RGB PRO
			{
				go func(vendorId, productId uint16, key string) {
					dev := k70pro.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K70Pro:      dev,
						ProductType: productTypeK70Pro,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 11024: // K65 PLUS USB
			{
				go func(vendorId, productId uint16, key string) {
					dev := k65plus.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K65Plus:     dev,
						ProductType: productTypeK65Plus,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 11015: // K65 PLUS USB
			{
				go func(vendorId, productId uint16, key string) {
					dev := k65plusW.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K65PlusW:    dev,
						ProductType: productTypeK65PlusW,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 7083: // K100 AIR USB
			{
				go func(vendorId, productId uint16, key string) {
					dev := k100air.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K100Air:     dev,
						ProductType: productTypeK100Air,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 7132: // K100 AIR WIRELESS
			{
				go func(vendorId, productId uint16, key string) {
					dev := k100airW.Init(vendorId, productId, key)
					if dev == nil {
						return
					}
					devices[dev.Serial] = &Device{
						K100AirW:    dev,
						ProductType: productTypeK100AirW,
						Product:     dev.Product,
						Serial:      dev.Serial,
						Firmware:    dev.Firmware,
					}
				}(vendorId, productId, key)
			}
		case 0: // Memory
			{
				go func(serialId string) {
					dev := memory.Init(serialId, "Memory")
					if dev != nil {
						devices[dev.Serial] = &Device{
							Memory:      dev,
							ProductType: productTypeMemory,
							Product:     dev.Product,
							Serial:      dev.Serial,
							Firmware:    "0",
						}
					}
				}(key)
			}
		default:
			continue
		}
	}
}
