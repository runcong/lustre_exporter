// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"
)

var (
	// HealthStatusEnabled specifies whether to collect Health metrics
	HealthStatusEnabled string

	// // OstEnabled specifies whether to collect OST metrics
	// OstEnabled string
	// // MdtEnabled specifies whether to collect MDT metrics
	// MdtEnabled string
	// // MgsEnabled specifies whether to collect MGS metrics
	// MgsEnabled string
	// // MdsEnabled specifies whether to collect MDS metrics
	// MdsEnabled string
	// // ClientEnabled specifies whether to collect Client metrics
	// ClientEnabled string
	// // GenericEnabled specifies whether to collect Generic metrics
	// GenericEnabled string
)

// type lustreJobsMetric struct {
// 	jobID string
// 	lustreStatsMetric
// }

// type lustreBRWMetric struct {
// 	size      string
// 	operation string
// 	value     string
// }

// type multistatParsingStruct struct {
// 	index   int
// 	pattern string
// }

func init() {
	Factories["sysfs"] = newLustreSysFsSource
}

type LustreSysFsSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *LustreSysFsSource) generateHealthStatusTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"": {
			{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "health", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateOSTMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*-OST*": {
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", gaugeMetric, false, core},
			{"grant_precreate", "grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", gaugeMetric, false, extended},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", gaugeMetric, false, extended},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction", gaugeMetric, false, extended},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync", gaugeMetric, false, extended},
			{"sync_journal", "sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", gaugeMetric, false, extended},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count", "Number of locks", gaugeMetric, false, extended},
			{"lock_timeouts", "lock_timeout", "Number of lock timeouts", counterMetric, false, extended},
			{"contended_locks", "lock_contended", "Number of contended locks", gaugeMetric, false, extended},
			{"contention_seconds", "lock_contention_seconds", "Time in seconds during which locks were contended", gaugeMetric, false, extended},

			{"pool/granted", "lock_granted", "Number of granted locks", gaugeMetric, false, extended},
			{"pool/grant_plan", "lock_grant_plan", "Number of planned lock grants per second", gaugeMetric, false, extended},
			{"pool/grant_rate", "lock_grant_rate", "Lock grant rate", gaugeMetric, false, extended},
		},
		"osd-*/*-OST*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"brw_stats", "pages_per_bulk_rw_total", pagesPerBlockRWHelp, counterMetric, false, extended},
			{"brw_stats", "discontiguous_pages_total", discontiguousPagesHelp, counterMetric, false, extended},
			{"brw_stats", "disk_io", diskIOsInFlightHelp, gaugeMetric, false, core},
			{"brw_stats", "io_time_milliseconds_total", ioTimeHelp, counterMetric, false, core},
			{"brw_stats", "disk_io_total", diskIOSizeHelp, counterMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "ost", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateMDTMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"osd-*/*-MDT*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
		"mdt/*": {
			{mdStats, "stats_total", statsHelp, counterMetric, true, core},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", counterMetric, false, core},
			{"job_stats", "job_stats_total", jobStatsHelp, counterMetric, true, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "mdt", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateMGSMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"mgs/MGS/osd/": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "mgs", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateMDSMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "mds", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateClientMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"llite/*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"checksum_pages", "checksum_pages_enabled", "Returns '1' if data checksumming is enabled for the client", gaugeMetric, false, extended},
			{"default_easize", "default_ea_size_bytes", "Default Extended Attribute (EA) size in bytes", gaugeMetric, false, extended},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
			{"lazystatfs", "lazystatfs_enabled", "Returns '1' if lazystatfs (a non-blocking alternative to statfs) is enabled for the client", gaugeMetric, false, extended},
			{"max_easize", "maximum_ea_size_bytes", "Maximum Extended Attribute (EA) size in bytes", gaugeMetric, false, extended},
			{"max_read_ahead_mb", "maximum_read_ahead_megabytes", "Maximum number of megabytes to read ahead", gaugeMetric, false, extended},
			{"max_read_ahead_per_file_mb", "maximum_read_ahead_per_file_megabytes", "Maximum number of megabytes per file to read ahead", gaugeMetric, false, extended},
			{"max_read_ahead_whole_mb", "maximum_read_ahead_whole_megabytes", "Maximum file size in megabytes for a file to be read in its entirety", gaugeMetric, false, extended},
			{"statahead_agl", "statahead_agl_enabled", "Returns '1' if the Asynchronous Glimpse Lock (AGL) for statahead is enabled", gaugeMetric, false, extended},
			{"statahead_max", "statahead_maximum", "Maximum window size for statahead", gaugeMetric, false, extended},
			{"stats", "read_samples_total", readSamplesHelp, counterMetric, false, core},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, gaugeMetric, false, extended},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, gaugeMetric, false, extended},
			{"stats", "read_bytes_total", readTotalHelp, counterMetric, false, core},
			{"stats", "write_samples_total", writeSamplesHelp, counterMetric, false, core},
			{"stats", "write_minimum_size_bytes", writeMinimumHelp, gaugeMetric, false, extended},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, gaugeMetric, false, extended},
			{"stats", "write_bytes_total", writeTotalHelp, counterMetric, false, core},
			{"stats", "stats_total", statsHelp, counterMetric, true, core},
			{"xattr_cache", "xattr_cache_enabled", "Returns '1' if extended attribute cache is enabled", gaugeMetric, false, extended},
		},
		"mdc/*": {
			{"rpc_stats", "rpcs_in_flight", rpcsInFlightHelp, gaugeMetric, true, core},
		},
		"osc/*": {
			{"rpc_stats", "pages_per_rpc_total", pagesPerRPCHelp, counterMetric, false, core},
			{"rpc_stats", "rpcs_in_flight", rpcsInFlightHelp, gaugeMetric, true, core},
			{"rpc_stats", "rpcs_offset", offsetHelp, gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "client", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateGenericMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"sptlrpc": {
			{"encrypt_page_pools", "physical_pages", physicalPagesHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "pages_per_pool", pagesPerPoolHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "maximum_pages", maxPagesHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "maximum_pools", maxPoolsHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "pages_in_pools", totalPagesHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "free_pages", totalFreeHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "maximum_pages_reached_total", maxPagesReachedHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "grows_total", growsHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "grows_failure_total", growsFailureHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "shrinks_total", shrinksHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "cache_access_total", cacheAccessHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "cache_miss_total", cacheMissingHelp, counterMetric, false, extended},
			{"encrypt_page_pools", "free_page_low", lowFreeMarkHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "maximum_waitqueue_depth", maxWaitQueueDepthHelp, gaugeMetric, false, extended},
			{"encrypt_page_pools", "out_of_memory_request_total", outOfMemHelp, counterMetric, false, extended},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "generic", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func newLustreSysFsSource() LustreSource {
	var l LustreSysFsSource
	l.basePath = filepath.Join(SysLocation, "fs/lustre")
	if HealthStatusEnabled != disabled {
		l.generateHealthStatusTemplates(HealthStatusEnabled)
	}
	if OstEnabled != disabled {
		l.generateOSTMetricTemplates(OstEnabled)
	}
	//control which node metrics you pull via flags

	if MdtEnabled != disabled {
		l.generateMDTMetricTemplates(MdtEnabled)
	}
	if MgsEnabled != disabled {
		l.generateMGSMetricTemplates(MgsEnabled)
	}
	if MdsEnabled != disabled {
		l.generateMDSMetricTemplates(MdsEnabled)
	}
	if ClientEnabled != disabled {
		l.generateClientMetricTemplates(ClientEnabled)
	}
	if GenericEnabled != disabled {
		l.generateGenericMetricTemplates(GenericEnabled)
	}
	return &l
}

func (s *LustreSysFsSource) Update(ch chan<- prometheus.Metric) (err error) {
	var directoryDepth int
	var metricType string

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			switch metric.filename {
			case "health_check":
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				err = s.parseFile(metric.source, single, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", extraLabel}, []string{nodeType, nodeName, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}
	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = single
			switch metric.filename {
			case "brw_stats", "rpc_stats":
				err = s.parseBRWStats(metric.source, "stats", path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target", "operation", "size"}, []string{nodeType, nodeName, brwOperation, brwSize}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", "operation", "size", extraLabel}, []string{nodeType, nodeName, brwOperation, brwSize, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = s.parseJobStats(metric.source, "job_stats", path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, jobid string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target", "jobid"}, []string{nodeType, nodeName, jobid}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", "jobid", extraLabel}, []string{nodeType, nodeName, jobid, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			default:
				if metric.filename == stats {
					metricType = stats
				} else if metric.filename == mdStats {
					metricType = mdStats
				} else if metric.filename == encryptPagePools {
					metricType = encryptPagePools
				}
				err = s.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", extraLabel}, []string{nodeType, nodeName, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// func getStatsOperationMetrics(statsFile string, promName string, helpText string) (metricList []lustreStatsMetric, err error) {
// 	operationSlice := []multistatParsingStruct{
// 		{pattern: "open", index: 1},
// 		{pattern: "close", index: 1},
// 		{pattern: "getattr", index: 1},
// 		{pattern: "setattr", index: 1},
// 		{pattern: "getxattr", index: 1},
// 		{pattern: "setxattr", index: 1},
// 		{pattern: "statfs", index: 1},
// 		{pattern: "seek", index: 1},
// 		{pattern: "readdir", index: 1},
// 		{pattern: "truncate", index: 1},
// 		{pattern: "alloc_inode", index: 1},
// 		{pattern: "removexattr", index: 1},
// 		{pattern: "unlink", index: 1},
// 		{pattern: "inode_permission", index: 1},
// 		{pattern: "create", index: 1},
// 		{pattern: "get_info", index: 1},
// 		{pattern: "set_info_async", index: 1},
// 		{pattern: "connect", index: 1},
// 		{pattern: "ping", index: 1},
// 	}
// 	for _, operation := range operationSlice {
// 		opStat := regexCaptureString(operation.pattern+" .*", statsFile)
// 		if len(opStat) < 1 {
// 			continue
// 		}
// 		r, err := regexp.Compile(" +")
// 		if err != nil {
// 			continue
// 		}
// 		bytesSplit := r.Split(opStat, -1)
// 		result, err := strconv.ParseFloat(bytesSplit[operation.index], 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		metricList = append(metricList, *newLustreStatsMetric(promName, helpText, result, "operation", operation.pattern))
// 	}
// 	return metricList, nil
// }

// func getStatsIOMetrics(statsFile string, promName string, helpText string) (metricList []lustreStatsMetric, err error) {
// 	// bytesSplit is in the following format:
// 	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
// 	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
// 	bytesMap := map[string]multistatParsingStruct{
// 		readSamplesHelp:       {pattern: "read_bytes .*", index: 1},
// 		readMinimumHelp:       {pattern: "read_bytes .*", index: 4},
// 		readMaximumHelp:       {pattern: "read_bytes .*", index: 5},
// 		readTotalHelp:         {pattern: "read_bytes .*", index: 6},
// 		writeSamplesHelp:      {pattern: "write_bytes .*", index: 1},
// 		writeMinimumHelp:      {pattern: "write_bytes .*", index: 4},
// 		writeMaximumHelp:      {pattern: "write_bytes .*", index: 5},
// 		writeTotalHelp:        {pattern: "write_bytes .*", index: 6},
// 		physicalPagesHelp:     {pattern: "physical pages: .*", index: 2},
// 		pagesPerPoolHelp:      {pattern: "pages per pool: .*", index: 3},
// 		maxPagesHelp:          {pattern: "max pages: .*", index: 2},
// 		maxPoolsHelp:          {pattern: "max pools: .*", index: 2},
// 		totalPagesHelp:        {pattern: "total pages: .*", index: 2},
// 		totalFreeHelp:         {pattern: "total free: .*", index: 2},
// 		maxPagesReachedHelp:   {pattern: "max pages reached: .*", index: 3},
// 		growsHelp:             {pattern: "grows: .*", index: 1},
// 		growsFailureHelp:      {pattern: "grows failure: .*", index: 2},
// 		shrinksHelp:           {pattern: "shrinks: .*", index: 1},
// 		cacheAccessHelp:       {pattern: "cache access: .*", index: 2},
// 		cacheMissingHelp:      {pattern: "cache missing: .*", index: 2},
// 		lowFreeMarkHelp:       {pattern: "low free mark: .*", index: 3},
// 		maxWaitQueueDepthHelp: {pattern: "max waitqueue depth: .*", index: 3},
// 		outOfMemHelp:          {pattern: "out of mem: .*", index: 3},
// 	}
// 	pattern := bytesMap[helpText].pattern
// 	bytesString := regexCaptureString(pattern, statsFile)
// 	if len(bytesString) < 1 {
// 		return nil, nil
// 	}
// 	r, err := regexp.Compile(" +")
// 	if err != nil {
// 		return nil, err
// 	}
// 	bytesSplit := r.Split(bytesString, -1)
// 	result, err := strconv.ParseFloat(bytesSplit[bytesMap[helpText].index], 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	metricList = append(metricList, *newLustreStatsMetric(promName, helpText, result, "", ""))

// 	return metricList, nil
// }

// func splitBRWStats(statBlock string) (metricList []lustreBRWMetric, err error) {
// 	if len(statBlock) == 0 || statBlock == "" {
// 		return nil, nil
// 	}

// 	// Skip the first line of text as it doesn't contain any metrics
// 	for _, line := range strings.Split(statBlock, "\n")[1:] {
// 		if len(line) > 1 {
// 			fields := strings.Fields(line)
// 			// Lines are in the following format:
// 			// [size] [# read RPCs] [relative read size (%)] [cumulative read size (%)] | [# write RPCs] [relative write size (%)] [cumulative write size (%)]
// 			// [0]    [1]           [2]                      [3]                       [4] [5]           [6]                       [7]
// 			if len(fields) >= 6 {
// 				size, readRPCs, writeRPCs := fields[0], fields[1], fields[5]
// 				size = strings.Replace(size, ":", "", -1)
// 				metricList = append(metricList, lustreBRWMetric{size: size, operation: "read", value: readRPCs})
// 				metricList = append(metricList, lustreBRWMetric{size: size, operation: "write", value: writeRPCs})
// 			} else if len(fields) >= 1 {
// 				size, rpcs := fields[0], fields[1]
// 				size = strings.Replace(size, ":", "", -1)
// 				metricList = append(metricList, lustreBRWMetric{size: size, operation: "read", value: rpcs})
// 			} else {
// 				continue
// 			}
// 		}
// 	}
// 	return metricList, nil
// }

// func parseStatsFile(helpText string, promName string, path string, hasMultipleVals bool) (metricList []lustreStatsMetric, err error) {
// 	statsFileBytes, err := ioutil.ReadFile(filepath.Clean(path))
// 	if err != nil {
// 		return nil, err
// 	}
// 	statsFile := string(statsFileBytes[:])
// 	var statsList []lustreStatsMetric
// 	if hasMultipleVals {
// 		statsList, err = getStatsOperationMetrics(statsFile, promName, helpText)
// 	} else {
// 		statsList, err = getStatsIOMetrics(statsFile, promName, helpText)
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	if statsList != nil {
// 		metricList = append(metricList, statsList...)
// 	}

// 	return metricList, nil
// }

// func getJobStatsIOMetrics(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
// 	// opMap matches the given helpText value with the placement of the numeric fields within each metric line.
// 	// For example, the number of samples is the first number in the line and has a helpText of readSamplesHelp,
// 	// hence the 'index' value of 0. 'pattern' is the regex capture pattern for the desired line.
// 	opMap := map[string]multistatParsingStruct{
// 		readSamplesHelp:  {index: 0, pattern: "read_bytes"},
// 		readMinimumHelp:  {index: 1, pattern: "read_bytes"},
// 		readMaximumHelp:  {index: 2, pattern: "read_bytes"},
// 		readTotalHelp:    {index: 3, pattern: "read_bytes"},
// 		writeSamplesHelp: {index: 0, pattern: "write_bytes"},
// 		writeMinimumHelp: {index: 1, pattern: "write_bytes"},
// 		writeMaximumHelp: {index: 2, pattern: "write_bytes"},
// 		writeTotalHelp:   {index: 3, pattern: "write_bytes"},
// 	}
// 	// If the metric isn't located in the map, don't try to parse a value for it.
// 	if _, exists := opMap[helpText]; !exists {
// 		return nil, nil
// 	}
// 	pattern := opMap[helpText].pattern
// 	opStat := regexCaptureString(pattern+": .*", jobBlock)
// 	opNumbers := regexCaptureNumbers(opStat)
// 	if len(opNumbers) < 1 {
// 		return nil, nil
// 	}
// 	result, err := strconv.ParseFloat(strings.TrimSpace(opNumbers[opMap[helpText].index]), 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if result == 0 {
// 		return nil, nil
// 	}
// 	metricList = append(metricList,
// 		lustreJobsMetric{jobID: jobID,
// 			lustreStatsMetric: *newLustreStatsMetric(promName, helpText, result, "", ""),
// 		})

// 	return metricList, err
// }

// func getJobNum(jobBlock string) (jobID string, err error) {
// 	jobID = regexCaptureJobid(jobBlock)
// 	if jobID == "" {
// 		return "", errors.New("No valid jobid found in block: " + jobBlock)
// 	}
// 	return strings.TrimSpace(jobID), nil
// }

// func getJobStatsOperationMetrics(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
// 	operationSlice := []multistatParsingStruct{
// 		{index: 0, pattern: "open"},
// 		{index: 0, pattern: "close"},
// 		{index: 0, pattern: "mknod"},
// 		{index: 0, pattern: "link"},
// 		{index: 0, pattern: "unlink"},
// 		{index: 0, pattern: "mkdir"},
// 		{index: 0, pattern: "rmdir"},
// 		{index: 0, pattern: "rename"},
// 		{index: 0, pattern: "getattr"},
// 		{index: 0, pattern: "setattr"},
// 		{index: 0, pattern: "getxattr"},
// 		{index: 0, pattern: "setxattr"},
// 		{index: 0, pattern: "statfs"},
// 		{index: 0, pattern: "sync"},
// 		{index: 0, pattern: "samedir_rename"},
// 		{index: 0, pattern: "crossdir_rename"},
// 		{index: 0, pattern: "punch"},
// 		{index: 0, pattern: "destroy"},
// 		{index: 0, pattern: "create"},
// 		{index: 0, pattern: "get_info"},
// 		{index: 0, pattern: "set_info"},
// 		{index: 0, pattern: "quotactl"},
// 	}
// 	for _, operation := range operationSlice {
// 		opStat := regexCaptureString(operation.pattern+": .*", jobBlock)
// 		opNumbers := regexCaptureStrings("[0-9]*\\.[0-9]+|[0-9]+", opStat)
// 		if len(opNumbers) < 1 {
// 			continue
// 		}
// 		var result float64
// 		result, err = strconv.ParseFloat(strings.TrimSpace(opNumbers[operation.index]), 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if result == 0 {
// 			continue
// 		}
// 		metricList = append(metricList,
// 			lustreJobsMetric{jobID: jobID,
// 				lustreStatsMetric: *newLustreStatsMetric(promName, helpText, result, "operation", operation.pattern),
// 			})
// 	}
// 	return metricList, err
// }

// func parseJobStatsText(jobStats string, promName string, helpText string, hasMultipleVals bool) (metricList []lustreJobsMetric, err error) {
// 	jobs := regexCaptureJobStats(jobStats)
// 	if len(jobs) < 1 {
// 		return nil, nil
// 	}
// 	var jobList []lustreJobsMetric
// 	for _, job := range jobs {
// 		jobID, err := getJobNum(job)
// 		if err != nil {
// 			log.Error(err)
// 			continue
// 		}
// 		if hasMultipleVals {
// 			jobList, err = getJobStatsOperationMetrics(job, jobID, promName, helpText)
// 		} else {
// 			jobList, err = getJobStatsIOMetrics(job, jobID, promName, helpText)
// 		}
// 		if err != nil {
// 			return nil, err
// 		}
// 		if jobList != nil {
// 			metricList = append(metricList, jobList...)
// 		}
// 	}
// 	return metricList, nil
// }

func (s *LustreSysFsSource) parseJobStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	jobStatsBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}

	jobStatsFile := string(jobStatsBytes[:])

	metricList, err := parseJobStatsText(jobStatsFile, promName, helpText, hasMultipleVals)
	if err != nil {
		return err
	}

	for _, item := range metricList {
		handler(nodeType, item.jobID, nodeName, item.lustreStatsMetric.title, item.lustreStatsMetric.help, item.lustreStatsMetric.value, item.lustreStatsMetric.extraLabel, item.lustreStatsMetric.extraLabelValue)
	}
	return nil
}

func (s *LustreSysFsSource) parseBRWStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	metricBlocks := map[string]string{
		pagesPerBlockRWHelp:    "pages per bulk r/w",
		discontiguousPagesHelp: "discontiguous pages",
		diskIOsInFlightHelp:    "disk I/Os in flight",
		ioTimeHelp:             "I/O time",
		diskIOSizeHelp:         "disk I/O size",
		pagesPerRPCHelp:        "pages per rpc",
		rpcsInFlightHelp:       "rpcs in flight",
		offsetHelp:             "offset",
	}
	statsFileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	statsFile := string(statsFileBytes[:])
	block := regexCaptureString("(?ms:^"+metricBlocks[helpText]+".*?(\n\n|\\z))", statsFile)
	metricList, err := splitBRWStats(block)
	if err != nil {
		return err
	}
	extraLabel := ""
	extraLabelValue := ""
	if hasMultipleVals {
		extraLabel = "type"
		pathElements := strings.Split(path, "/")
		extraLabelValue = pathElements[len(pathElements)-3]
	}
	for _, item := range metricList {
		value, err := strconv.ParseFloat(item.value, 64)
		if err != nil {
			return err
		}
		handler(nodeType, item.operation, convertToBytes(item.size), nodeName, promName, helpText, value, extraLabel, extraLabelValue)
	}
	return nil
}

func (s *LustreSysFsSource) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	fileString := string(fileBytes[:])
	switch filename {
	case "health_check":
		if strings.TrimSpace(fileString) == "healthy" {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckHealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckUnhealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
	}
	return nil
}

func (s *LustreSysFsSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue, "", "")
	case stats, mdStats, encryptPagePools:
		metricList, err := parseStatsFile(helpText, promName, path, hasMultipleVals)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, metric.help, metric.value, metric.extraLabel, metric.extraLabelValue)
		}
	}
	return nil
}
