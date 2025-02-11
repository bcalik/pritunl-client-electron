package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client-electron/service/errortypes"
	"github.com/pritunl/pritunl-client-electron/service/utils"
)

var (
	Config            = &ConfigData{}
	StaticRoot        = ""
	StaticTestingRoot = ""
)

type ConfigData struct {
	path              string `json:"-"`
	loaded            bool   `json:"-"`
	DisableDnsWatch   bool   `json:"disable_dns_watch"`
	DisableDnsRefresh bool   `json:"disable_dns_refresh"`
	DisableWakeWatch  bool   `json:"disable_wake_watch"`
	DisableNetClean   bool   `json:"disable_net_clean"`
	ForceLocalTpm     bool   `json:"force_local_tpm"`
	InterfaceMetric   int    `json:"interface_metric"`
	EnclavePrivateKey string `json:"enclave_private_key"`
}

func (c *ConfigData) Save() (err error) {
	if !c.loaded {
		err = &errortypes.WriteError{
			errors.New("config: Config file has not been loaded"),
		}
		return
	}

	pth := GetPath()

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File marshal error"),
		}
		return
	}

	err = utils.ExistsMkdir(filepath.Dir(pth), 0755)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(pth, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File write error"),
		}
		return
	}

	return
}

func Load() (err error) {
	data := &ConfigData{}

	pth := GetPath()

	_, err = os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			data.loaded = true
			Config = data
		} else {
			err = &errortypes.ReadError{
				errors.Wrap(err, "config: File stat error"),
			}
		}
		return
	}

	file, err := ioutil.ReadFile(pth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File read error"),
		}
		return
	}

	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File unmarshal error"),
		}
		return
	}

	data.loaded = true

	Config = data

	return
}

func Save() (err error) {
	err = Config.Save()
	if err != nil {
		return
	}

	return
}
