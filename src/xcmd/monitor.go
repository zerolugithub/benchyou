/*
 * benchyou
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package xcmd

import (
	"fmt"
	"os"
	"sysbench"
	"text/tabwriter"
	"time"
	"xcommon"
	"xstat"
)

type Stats struct {
	SystemCS uint64
	IdleCPU  uint64
	RRQM_S   float64
	WRQM_S   float64
	R_S      float64
	W_S      float64
	RKB_S    float64
	WKB_S    float64
	AWAIT    float64
	UTIL     float64
}

type Monitor struct {
	conf    *xcommon.Conf
	workers []sysbench.Worker
	ticker  *time.Ticker
	vms     *xstat.VMS
	ios     *xstat.IOS
	stats   *Stats
	all     *Stats
	seconds uint64
}

func NewMonitor(conf *xcommon.Conf, workers []sysbench.Worker) *Monitor {
	return &Monitor{
		conf:    conf,
		workers: workers,
		ticker:  time.NewTicker(time.Second),
		vms:     xstat.NewVMS(conf),
		ios:     xstat.NewIOS(conf),
		stats:   &Stats{},
		all:     &Stats{},
	}
}

func (m *Monitor) Start() {
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	m.vms.Start()
	m.ios.Start()
	go func() {
		newm := &sysbench.Metric{}
		oldm := &sysbench.Metric{}
		for _ = range m.ticker.C {
			m.seconds++
			m.stats.SystemCS = m.vms.Stat.SystemCS
			m.stats.IdleCPU = m.vms.Stat.IdleCPU
			m.stats.RRQM_S = m.ios.Stat.RRQM_S
			m.stats.WRQM_S = m.ios.Stat.WRQM_S
			m.stats.R_S = m.ios.Stat.R_S
			m.stats.W_S = m.ios.Stat.W_S
			m.stats.RKB_S = m.ios.Stat.RKB_S
			m.stats.WKB_S = m.ios.Stat.WKB_S
			m.stats.AWAIT = m.ios.Stat.AWAIT
			m.stats.UTIL = m.ios.Stat.UTIL

			m.all.SystemCS += m.stats.SystemCS
			m.all.IdleCPU += m.stats.IdleCPU
			m.all.RRQM_S += m.stats.RRQM_S
			m.all.WRQM_S += m.stats.WRQM_S
			m.all.R_S += m.stats.R_S
			m.all.W_S += m.stats.W_S
			m.all.RKB_S += m.stats.RKB_S
			m.all.WKB_S += m.stats.WKB_S
			m.all.AWAIT += m.stats.AWAIT
			m.all.UTIL += m.stats.UTIL

			newm = sysbench.AllWorkersMetric(m.workers)
			wtps := float64(newm.WNums - oldm.WNums)
			rtps := float64(newm.QNums - oldm.QNums)
			tps := wtps + rtps

			fmt.Fprintln(w, "time   \t\t   thds  \t tps   \twtps  \trtps  \trio  \trio/op \twio  \twio/op  \trMB   \trKB/op  \twMB   \twKB/op \tcpu/op\t")
			line := fmt.Sprintf("[%ds]\t\t[r:%d,w:%d]\t%d\t%d\t%d\t%d\t%.2f\t%d\t%0.2f\t%2.2f\t%.2f\t%2.2f\t%.2f\t%.2f\t\n",
				m.seconds,
				m.conf.Read_threads,
				m.conf.Write_threads,
				int(tps),
				int(wtps),
				int(rtps),
				int(m.stats.R_S),
				m.stats.R_S/tps,
				int(m.stats.W_S),
				m.stats.W_S/tps,
				m.stats.RKB_S/1024,
				m.stats.RKB_S/tps,
				m.stats.WKB_S/1024,
				m.stats.WKB_S/tps,
				float64(m.stats.SystemCS)/tps,
			)
			fmt.Fprintln(w, line)

			w.Flush()
			*oldm = *newm
		}
	}()
}

func (m *Monitor) Stop() {
	m.ticker.Stop()
	sysbench.StopWorkers(m.workers)

	// avg results at the end
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	counts := float64(m.seconds)
	all := sysbench.AllWorkersMetric(m.workers)
	wtps := float64(all.WNums)
	rtps := float64(all.QNums)
	tps := wtps + rtps

	fmt.Fprintln(w, "-----------------------------------------------------------------------------------avg---------------------------------------------------------------------------------------------")
	fmt.Fprintln(w, "time   \t\t tps   \twtps  \trtps  \trio  \trio/op \twio  \twio/op  \trMB   \trKB/op  \twMB   \twKB/op \tcpu/op\t          w-rsp(ms)\t          r-rsp(ms)")
	line := fmt.Sprintf("[%ds]\t\t%d\t%d\t%d\t%d\t%.2f\t%d\t%0.2f\t%2.2f\t%.2f\t%2.2f\t%.2f\t%.2f\t[avg:%.2f,min:%.2f,max:%.2f]\t[avg:%.2f,min:%.2f,max:%.2f]\n",
		m.seconds,
		int(tps/counts),
		int(wtps/counts),
		int(rtps/counts),
		int(m.stats.R_S/counts),
		m.stats.R_S/tps,
		int(m.stats.W_S/counts),
		m.stats.W_S/tps/counts,
		m.stats.RKB_S/1024/counts,
		m.stats.RKB_S/tps/counts,
		m.stats.WKB_S/1024/counts,
		m.stats.WKB_S/tps/counts,
		float64(m.stats.SystemCS)/tps/counts,
		float64(all.WCosts)/1e6/wtps/counts,
		float64(all.WMin)/1e6,
		float64(all.WMax)/1e6,
		float64(all.QCosts)/1e6/rtps/counts,
		float64(all.QMin)/1e6,
		float64(all.QMax)/1e6,
	)
	fmt.Fprintln(w, line)
	w.Flush()
}