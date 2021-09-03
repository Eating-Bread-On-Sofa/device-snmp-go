// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var once sync.Once
var driver *Driver

type Driver struct {
	lc      logger.LoggingClient
	asyncCh chan<- *dsModels.AsyncValues
}

func NewProtocolDriver() dsModels.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Info(fmt.Sprintf("Driver.DisconnectDevice: device-snmp-go driver is disconnecting to %s", deviceName))
	return nil
}

func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	d.lc = lc
	d.asyncCh = asyncCh
	return nil
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	if len(reqs) != 1 {
		err = fmt.Errorf("Driver.HandleReadCommands; too many command requests; only one supported")
		return
	}

	d.lc.Debug(fmt.Sprintf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))

	res = make([]*dsModels.CommandValue, 1)
	now := time.Now().UnixNano()
	var SNMP = protocols["snmp"]
	addr := SNMP["Address"]
	port := SNMP["Port"]
	apiUrl := "http://localhost:8093/snmpread"

	param := url.Values{}
	param.Set("post", addr)
	param.Set("port", port)

	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		fmt.Printf("parse url requestUrl failed, err:%v\n", err)
	}
	u.RawQuery = param.Encode() // URL encode

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	var dt map[string]interface{}

	if err := json.Unmarshal(b, &dt); err != nil {
		fmt.Printf("json.Unmarsha jsonStr1 failed, err:%v\n", err)
		return nil, err
	}

	var v float64

	if reqs[0].DeviceResourceName == "TemperatureDeg" {
		v, _ = strconv.ParseFloat(fmt.Sprint(dt["temp"]), 64)
	} else {
		v, _ = strconv.ParseFloat(fmt.Sprint(dt["temp"]), 64)
	}

	cv, _ := dsModels.NewFloat64Value(reqs[0].DeviceResourceName, now, v)
	res[0] = cv

	return res, nil
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	d.lc.Debug(fmt.Sprintf("Driver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[0].DeviceResourceName, params))
	return nil
}

func (d *Driver) Stop(force bool) error {
	d.lc.Info("Driver.Stop: device-snmp-go driver is stopping...")
	return nil
}

func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}
