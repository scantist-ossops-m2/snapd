// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2022 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/secboot/keymgr"
	"github.com/snapcore/snapd/secboot/keys"
)

var osStdin io.Reader = os.Stdin

type commonMultiDeviceMixin struct {
	Device        []string `long:"device" description:"encrypted device (can be more than one)" required:"yes"`
	Authorization []string `long:"authorization" description:"authorization source (for each device, either 'keyring' or 'file:<key-file>')" required:"yes"`
}

type cmdAddRecoveryKey struct {
	commonMultiDeviceMixin
	KeyFile string `long:"key-file" description:"path for generated recovery key file" required:"yes"`
}

type cmdRemoveRecoveryKey struct {
	commonMultiDeviceMixin
	KeyFile []string `long:"key-file" description:"path to recovery key file(s) to be removed" required:"yes"`
}

type cmdChangeEncryptionKey struct {
	Device string `long:"device" description:"encrypted device" required:"yes"`
}

type options struct {
	CmdAddRecoveryKey      cmdAddRecoveryKey      `command:"add-recovery-key"`
	CmdRemoveRecoveryKey   cmdRemoveRecoveryKey   `command:"remove-recovery-key"`
	CmdChangeEncryptionKey cmdChangeEncryptionKey `command:"change-encryption-key"`
}

var (
	keymgrAddRecoveryKeyToLUKSDevice              = keymgr.AddRecoveryKeyToLUKSDevice
	keymgrAddRecoveryKeyToLUKSDeviceUsingKey      = keymgr.AddRecoveryKeyToLUKSDeviceUsingKey
	keymgrRemoveRecoveryKeyFromLUKSDevice         = keymgr.RemoveRecoveryKeyFromLUKSDevice
	keymgrRemoveRecoveryKeyFromLUKSDeviceUsingKey = keymgr.RemoveRecoveryKeyFromLUKSDeviceUsingKey
	keymgrChangeLUKSDeviceEncryptionKey           = keymgr.ChangeLUKSDeviceEncryptionKey
)

func (c *cmdAddRecoveryKey) Execute(args []string) error {
	recoveryKey, err := keys.NewRecoveryKey()
	if err != nil {
		return fmt.Errorf("cannot create recovery key: %v", err)
	}
	// TODO make this idempotent, possible solution is:
	// 1. write the key file if none is present
	// 2. if the key file was present, read it back
	// 3. add the key
	// 4. if adding failed with keyslot already in used and the file was
	// present assume it's correct
	if len(c.Authorization) != len(c.Device) {
		return fmt.Errorf("cannot add recovery key: mismatch in the number of devices and authorizations")
	}
	for i, dev := range c.Device {
		authz := c.Authorization[i]
		if authz == "keyring" {
			if err := keymgrAddRecoveryKeyToLUKSDevice(recoveryKey, dev); err != nil {
				return fmt.Errorf("cannot add recovery key to LUKS device: %v", err)
			}
		} else if strings.HasPrefix(authz, "file:") {
			authzKey, err := ioutil.ReadFile(authz[len("file:"):])
			if err != nil {
				return fmt.Errorf("cannot load authorization key: %v", err)
			}
			if err := keymgrAddRecoveryKeyToLUKSDeviceUsingKey(recoveryKey, authzKey, dev); err != nil {
				return fmt.Errorf("cannot add recovery key to LUKS device using authorization key: %v", err)
			}
		}
	}
	if err := ioutil.WriteFile(c.KeyFile, recoveryKey[:], 0600); err != nil {
		return fmt.Errorf("cannot write recovery key to file: %v", err)
	}
	return nil
}

func (c *cmdRemoveRecoveryKey) Execute(args []string) error {
	if len(c.Authorization) != len(c.Device) {
		return fmt.Errorf("cannot remove recovery key: mismatch in the number of devices and authorizations")
	}
	for i, dev := range c.Device {
		authz := c.Authorization[i]
		if authz == "keyring" {
			if err := keymgrRemoveRecoveryKeyFromLUKSDevice(dev); err != nil {
				return fmt.Errorf("cannot remove recovery key from LUKS device: %v", err)
			}
		} else if strings.HasPrefix(authz, "file:") {
			authzKey, err := ioutil.ReadFile(authz[len("file:"):])
			if err != nil {
				return fmt.Errorf("cannot load authorization key: %v", err)
			}
			if err := keymgrRemoveRecoveryKeyFromLUKSDeviceUsingKey(authzKey, dev); err != nil {
				return fmt.Errorf("cannot remove recovery key from device using authorization key: %v", err)
			}
		}
	}
	for _, kf := range c.KeyFile {
		if err := os.Remove(kf); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot remove recovery key file: %v", err)
		}
	}
	return nil
}

type newKey struct {
	Key []byte `json:"key"`
}

func (c *cmdChangeEncryptionKey) Execute(args []string) error {
	var newEncryptionKeyData newKey
	dec := json.NewDecoder(osStdin)
	if err := dec.Decode(&newEncryptionKeyData); err != nil {
		return fmt.Errorf("cannot obtain new encryption key: %v", err)
	}
	if err := keymgrChangeLUKSDeviceEncryptionKey(newEncryptionKeyData.Key, c.Device); err != nil {
		return fmt.Errorf("cannot change LUKS device encryption key: %v", err)
	}
	return nil
}

func run(osArgs1 []string) error {
	var opts options
	p := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	if _, err := p.ParseArgs(osArgs1); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
