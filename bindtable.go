package katti

import "strings"

type BindType uint8

const (
	Match BindType = iota
	MatchList
)

type BindValue struct {
	valueType BindType
	value     any
}

type BindTable struct {
	collection []map[string]BindValue
}

func (bm *BindTable) appendBindMap() {
	bm.collection = append(bm.collection, map[string]BindValue{})
}

func (bm *BindTable) activeMap() map[string]BindValue {
	activeLen := len(bm.collection)

	if activeLen == 0 {
		bm.collection = append(bm.collection, map[string]BindValue{})
		activeLen++
	}

	return bm.collection[activeLen-1]
}

func (bm *BindTable) Set(k string, v BindValue) {
	active := bm.activeMap()
	active[k] = v
}

func (bm *BindTable) get(k string) (BindValue, bool) {
	active := bm.activeMap()
	v, ok := active[k]
	return v, ok
}

func (bm *BindTable) Get(k string) []string {
	v, ok := bm.get(k)
	if !ok || v.valueType != Match {
		return nil
	}

	return v.value.([]string)
}

func (bm *BindTable) GetString(k string) string {
	return strings.Join(bm.Get(k), "")
}

func (bm *BindTable) GetList(k string) [][]string {
	v, ok := bm.get(k)
	if !ok || v.valueType != MatchList {
		return nil
	}

	return v.value.([][]string)
}
