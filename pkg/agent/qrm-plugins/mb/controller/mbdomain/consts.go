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

package mbdomain

const (
	DomainTotalMB         = 120_000          //120 GBps in one mb sharing domain
	ReservedPerNuma       = 25_000           // 25 GBps reserved per node for dedicated pod
	MaxMBDedicatedPerNuma = 60_000           // if a socket pod assigned to one numa node, its max mb is 60 GB
	LoungeMB              = 6_000            // lounge zone MB earmarked to dedicated qos is 6 GBps
	MaxMBPerCCD           = 2048 / 8 * 1_000 // AMD max MB value in schemata file
)
