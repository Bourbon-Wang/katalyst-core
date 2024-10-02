/*
Copyright 2022 The Katalyst Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package monitor

import (
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubewharf/katalyst-core/pkg/agent/qrm-plugins/mb/controller/util"
	"github.com/kubewharf/katalyst-core/pkg/agent/qrm-plugins/mb/task"
)

// MBQoSGroup keeps MB of qos control group at level of CCD
type MBQoSGroup struct {
	//nodes []int

	// CCDs is the set of CCDs that this group has tasks allocated to
	CCDs sets.Int

	// CCDMB MUST be in line with CCDs
	CCDMB map[int]int
}

func newMBQoSGroup(ccdMB map[int]int) *MBQoSGroup {
	result := &MBQoSGroup{
		CCDs:  make(sets.Int),
		CCDMB: ccdMB,
	}

	for ccd, _ := range ccdMB {
		result.CCDs.Insert(ccd)
	}

	return result
}

func SumMB(groups map[task.QoSGroup]*MBQoSGroup) int {
	sum := 0

	for _, group := range groups {
		sum += util.SumCCDMB(group.CCDMB)
	}
	return sum
}
