package tuntap

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-client-electron/service/config"
	"github.com/pritunl/pritunl-client-electron/service/platform"
	"github.com/pritunl/pritunl-client-electron/service/utils"
)

var (
	curSize      = 0
	curTotalSize = 0
	taps         = []string{}
	tapsLock     = sync.Mutex{}
)

func getToolpath() string {
	pth := filepath.Join(utils.GetRootDir(), "..",
		"tuntap_win", "tapctl.exe")

	exists, _ := utils.ExistsFile(pth)
	if exists {
		return pth
	}

	return filepath.Join(utils.GetRootDir(), "tuntap", "tapctl.exe")
}

func Configure() (err error) {
	if runtime.GOOS != "windows" {
		return
	}

	metric := config.Config.InterfaceMetric

	metricStr := ""
	if metric == 0 {
		metricStr = "automatic"
	} else {
		metricStr = strconv.Itoa(metric)
	}

	size := Size()

	systemDir, err := platform.SystemDirectory()
	if err != nil {
		return
	}

	pth := path.Join(systemDir, "netsh.exe")

	for i := 0; i < size; i++ {
		_, _ = utils.ExecInputOutputCombindLogged(
			fmt.Sprintf(
				"interface ipv4 set interface \"Pritunl %d\" metric=%s",
				i+1,
				metricStr,
			),
			pth,
		)
		_, _ = utils.ExecInputOutputCombindLogged(
			fmt.Sprintf(
				"interface ipv6 set interface \"Pritunl %d\" metric=%s",
				i+1,
				metricStr,
			),
			pth,
		)
	}

	return
}

func Get() (adpaters []string, count int, err error) {
	toolpath := getToolpath()

	output, err := utils.ExecCombinedOutputLogged(
		nil,
		toolpath,
		"list",
	)
	if err != nil {
		return
	}

	adpaters = []string{}
	for _, line := range strings.Split(output, "\n") {
		lines := strings.Fields(line)
		if len(lines) < 2 {
			continue
		}

		count += 1

		if !strings.Contains(strings.ToLower(line), "pritunl") {
			continue
		}

		adpaters = append(adpaters, lines[0])
	}

	return
}

func Clean() (err error) {
	toolpath := getToolpath()

	adapters, totalCount, err := Get()
	if err != nil {
		return
	}

	for _, adapter := range adapters {
		_, _ = utils.ExecCombinedOutputLogged(
			[]string{
				"No devices",
			},
			toolpath,
			"delete",
			adapter,
		)
		totalCount -= 1
	}

	curSize = 0
	curTotalSize = totalCount

	return
}

func Resize(size int) (err error) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	toolpath := getToolpath()

	if size <= 3 {
		size = 3
	} else if size < 6 {
		size = 6
	} else {
		size = 9
	}

	add := size - curSize

	for i := 0; i < add; i++ {
		tapName := fmt.Sprintf("Pritunl %d", curSize+1)

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			toolpath,
			"create",
			"--name", tapName,
		)
		if err != nil {
			err = nil

			_, _ = utils.ExecCombinedOutputLogged(
				[]string{
					"No devices",
				},
				toolpath,
				"delete",
				tapName,
			)

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				toolpath,
				"create",
				"--name", tapName,
			)
			if err != nil {
				err = nil

				if size-curTotalSize > 0 {
					_, err = utils.ExecCombinedOutputLogged(
						nil,
						toolpath,
						"create",
					)
					if err != nil {
						_ = Clean()
						return
					}

					curTotalSize += 1
				}

				time.Sleep(200 * time.Millisecond)
				continue
			}
		}

		curSize += 1
		curTotalSize += 1
		taps = append(taps, tapName)

		time.Sleep(200 * time.Millisecond)
	}

	sort.Strings(taps)

	return
}

func Size() int {
	return curSize
}

func Acquire() (tap string) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	if len(taps) == 0 {
		tap = "null"
		return
	}

	tap, taps = taps[0], taps[1:]
	return
}

func Release(tap string) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	if tap == "null" {
		return
	}

	taps = append(taps, tap)
	sort.Strings(taps)
}
