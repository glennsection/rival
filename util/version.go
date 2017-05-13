package util

import (
	"strings"
	"strconv"
)

// compare version numbers
// * expectedVersion       - what are we looking for
// * observedVersion       - what we received from the client
// * requiredCompatibility - how many components are required to match:
//     0 - no requirement
//     1 - only major version
//     2 - major and minor
//     3 - major, minor & build
//     etc.
func CompareVersion(expectedVersion string, observedVersion string, requiredCompatibility int) int {
	// split version into components
	expectedVersionComponents := strings.Split(expectedVersion, ".")
	expectedVersionSize := len(expectedVersionComponents)
	observedVersionComponents := strings.Split(observedVersion, ".")
	observedVersionSize := len(observedVersionComponents)

	// determine max components
	maxComponents := observedVersionSize
	if expectedVersionSize > observedVersionSize {
		maxComponents = expectedVersionSize
	}

	// loop through and compare components
	for i := 0; i < maxComponents && i < requiredCompatibility; i++ {
		var x, y string
		if expectedVersionSize > i {
			x = expectedVersionComponents[i]
		}
		if observedVersionSize > i {
			y = observedVersionComponents[i]
		}

		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			return -1
		} else if xi < yi {
			return 1
		}
	}
	return 0
}