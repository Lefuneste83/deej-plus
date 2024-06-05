package deej

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/thoas/go-funk"
)

type buttonMap struct {
	m    map[int][]string
	lock sync.Locker
}

func newButtonMap() *buttonMap {
	return &buttonMap{
		m:    make(map[int][]string),
		lock: &sync.Mutex{},
	}
}

func buttonrMapFromConfigs(userMapping map[string][]string, internalMapping map[string][]string) *buttonMap {
	resultMap := newButtonMap()

	// copy targets from user config, ignoring empty values
	for buttonIdxString, targets := range userMapping {
		buttonIdx, _ := strconv.Atoi(buttonIdxString)

		resultMap.set(buttonIdx, funk.FilterString(targets, func(s string) bool {
			return s != ""
		}))
	}

	// add targets from internal configs, ignoring duplicate or empty values
	for buttonIdxString, targets := range internalMapping {
		buttonIdx, _ := strconv.Atoi(buttonIdxString)

		existingTargets, ok := resultMap.get(buttonIdx)
		if !ok {
			existingTargets = []string{}
		}

		filteredTargets := funk.FilterString(targets, func(s string) bool {
			return (!funk.ContainsString(existingTargets, s)) && s != ""
		})

		existingTargets = append(existingTargets, filteredTargets...)
		resultMap.set(buttonIdx, existingTargets)
	}

	return resultMap
}

func (m *buttonMap) iterate(f func(int, []string)) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for key, value := range m.m {
		f(key, value)
	}
}

func (m *buttonMap) get(key int) ([]string, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	value, ok := m.m[key]
	return value, ok
}

func (m *buttonMap) set(key int, value []string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.m[key] = value
}

func (m *buttonMap) String() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	buttonCount := 0
	targetCount := 0

	for _, value := range m.m {
		buttonCount++
		targetCount += len(value)
	}

	return fmt.Sprintf("<%d buttons mapped to %d targets>", buttonCount, targetCount)
}
