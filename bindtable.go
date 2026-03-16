package katti

type BindTable struct {
	Collection []map[string][][]string
}

func (bm *BindTable) appendBindMap() {
	bm.Collection = append(bm.Collection, map[string][][]string{})
}

func (bm *BindTable) popBindMap() {
	bm.Collection = bm.Collection[len(bm.Collection):]
}

func (bm *BindTable) Set(k string, v []string) {
	activeLen := len(bm.Collection)

	if activeLen == 0 {
		bm.Collection = append(bm.Collection, map[string][][]string{})
		activeLen++
	}

	active := bm.Collection[activeLen-1]
	active[k] = append(active[k], v)
}

func (bm *BindTable) Get(k string) [][]string {
	for _, col := range bm.Collection {
		if v, ok := col[k]; ok {
			return v
		}
	}

	return nil
}
