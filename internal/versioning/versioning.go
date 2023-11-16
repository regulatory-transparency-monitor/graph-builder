package versioning

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

type VersionManager struct {
	CurrentVersion string
}

func NewVersionManager(initialVersion string) *VersionManager {
	return &VersionManager{CurrentVersion: initialVersion}
}

func (vm *VersionManager) GetCurrentVersion() string {
	return vm.CurrentVersion
}

func (vm *VersionManager) IncrementVersion() {
	vm.CurrentVersion = incrementVersion(vm.CurrentVersion)
}

func incrementVersion(version string) string {
	splitVersion := strings.Split(version, ".")
	if len(splitVersion) != 3 {
		logger.Error("Invalid version format: %s", version)
		return "0.0.0" // default if version format is not as expected
	}

	major, err1 := strconv.Atoi(splitVersion[0])
	minor, err2 := strconv.Atoi(splitVersion[1])
	patch, err3 := strconv.Atoi(splitVersion[2])

	if err1 != nil || err2 != nil || err3 != nil {
		logger.Error("Failed to parse version components: ", logger.LogFields{"major": err1, "minor": err2, "patch": err3})
		return "0.0.0"
	}
	patch++

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
