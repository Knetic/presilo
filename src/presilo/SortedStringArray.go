package presilo

import "sort"

type SortableStringArray []string

func (this *SortableStringArray) Sort() {
	sort.Sort(this)
}

// Len is part of sort.Interface.
func (this SortableStringArray) Len() int {
	return len(this)
}

// Swap is part of sort.Interface.
func (this SortableStringArray) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (this SortableStringArray) Less(i, j int) bool {
	return this[i] < this[j]
}
