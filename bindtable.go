package katti

import "strings"

type BindTable struct {
	Collection []map[string][]string
}

func (bm *BindTable) appendBindMap() {
	bm.Collection = append(bm.Collection, map[string][]string{})
}

func (bm *BindTable) popBindMap() {
	bm.Collection = bm.Collection[len(bm.Collection):]
}

func (bm *BindTable) activeMap() map[string][]string {
	activeLen := len(bm.Collection)

	if activeLen == 0 {
		bm.Collection = append(bm.Collection, map[string][]string{})
		activeLen++
	}

	return bm.Collection[activeLen-1]
}

func (bm *BindTable) Set(k string, v []string) {
	active := bm.activeMap()
	active[k] = append(active[k], v...)
}

func (bm *BindTable) Get(k string) []string {
	active := bm.activeMap()
	return active[k]
}

func (bm *BindTable) GetString(k string) string {
	return strings.Join(bm.Get(k), "")
}
