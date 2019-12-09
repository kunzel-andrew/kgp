package main

import "sort"

func searchIndexForWord(word string) PairList {
	if titles, ok := indexCache[word]; ok {
		pl := make(PairList, len(titles))
		i := 0
		for k, v := range titles {
			pl[i] = Pair{k, v}
			i++
		}
		sort.Sort(sort.Reverse(pl))

		return pl
	}
	return nil
}

type Pair struct {
	Title string
	Count int
}
type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool {
	if p[i].Count == p[j].Count {
		return p[i].Title < p[j].Title
	} else {
		return p[i].Count < p[j].Count
	}
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
