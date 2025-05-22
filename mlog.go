package mlog

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

type entry struct {
	time  time.Time
	label string
	msg   string
}

type panel struct {
	round [2048]entry
	count int
}

var (
	lock   sync.Mutex
	panels = map[string]*panel{}
	zone   = time.FixedZone("GMT+7", 7*3600)
	layout = "2006-01-02T15:04:05.000"
)

func Log(label, msg string, group ...string) {
	e := entry{time.Now(), label, msg}
	lock.Lock()
	defer lock.Unlock()

	p, ok := panels[""]
	if !ok {
		p = &panel{}
		panels[""] = p
	}
	p.add(e)

	for _, name := range slices.Compact(group) {
		if name == "" {
			continue
		}
		p, ok = panels[name]
		if !ok {
			p = &panel{}
			panels[name] = p
		}
		p.add(e)
	}
}

func (p *panel) add(e entry) {
	p.round[p.count%len(p.round)] = e
	p.count++
}

func (p *panel) Content() string {
	var slen int
	for i, n := p.count-1, 0; i >= 0 && n < len(p.round); {
		e := &p.round[i%len(p.round)]
		if slen < len(e.label) {
			slen = len(e.label)
		}
		i--
		n++
	}
	if slen > 0 {
		slen++
	}
	var b strings.Builder
	for i, n := p.count-1, 0; i >= 0 && n < len(p.round); {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		e := &p.round[i%len(p.round)]
		b.WriteString(e.time.In(zone).Format(layout))
		b.WriteByte(' ')
		b.WriteString(e.label)
		for j := len(e.label); j < slen; j++ {
			b.WriteByte(' ')
		}
		b.WriteString(e.msg)
		i--
		n++
	}
	return b.String()
}

func (p *panel) PerSecond(f func(string) bool) float64 {
	t := time.Now().Add(-3 * time.Second)
	var count int
	for i, n := p.count-1, 0; i >= 0 && n < len(p.round); {
		e := &p.round[i%len(p.round)]
		if t.After(e.time) {
			break
		}
		if f(e.msg) {
			count++
		}
		i--
		n++
	}
	count = count * 100 / 3
	return float64(count) / 100
}

func Groups() []string {
	lock.Lock()
	defer lock.Unlock()

	names := make([]string, 0, len(panels))
	for name := range panels {
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Output(group string) string {
	lock.Lock()
	defer lock.Unlock()

	if p, ok := panels[group]; ok {
		return p.Content()
	}
	return ""
}

func CountPerSecond(group string, contains string) float64 {
	return CountPerSecondFunc(group, func(s string) bool {
		return strings.Contains(s, contains)
	})
}

func CountPerSecondFunc(group string, f func(string) bool) float64 {
	lock.Lock()
	defer lock.Unlock()

	if p, ok := panels[group]; ok {
		return p.PerSecond(f)
	}
	return 0
}
