// https://github.com/f-secure-foundry/armory-ums
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/imx6/usb"
	"github.com/f-secure-foundry/tamago/imx6/usdhc"

	"github.com/f-secure-foundry/tamago/usbarmory/mark-two"
)

var cards []*usdhc.USDHC

func init() {
	log.SetFlags(0)

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v\n", err))
	}
}

func detect(card *usdhc.USDHC) (err error) {
	err = card.Detect()

	if err != nil {
		return
	}

	info := card.Info()
	capacity := int64(info.BlockSize) * int64(info.Blocks)
	giga := capacity / (1000 * 1000 * 1000)
	gibi := capacity / (1024 * 1024 * 1024)

	log.Printf("imx6_usdhc: %d GB/%d GiB card detected %+v", giga, gibi, info)

	cards = append(cards, card)

	return
}

func main() {
	err := detect(usbarmory.SD)

	if err != nil {
		usbarmory.LED("white", false)
	} else {
		usbarmory.LED("white", true)
	}

	err = detect(usbarmory.MMC)

	if err != nil {
		usbarmory.LED("blue", false)
	} else {
		usbarmory.LED("blue", true)
	}

	device := &usb.Device{
		Setup: setup,
	}
	configureDevice(device)

	iface := buildMassStorageInterface()
	device.Configurations[0].AddInterface(iface)

	dma.Init(dmaStart, dmaSize)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
