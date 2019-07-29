package merger

import (
	"sort"
)

// CPMerge was described in paper
// "Simple and Efficient Algorithm for Approximate Dictionary Matching"
// inspired by https://github.com/chokkan/simstring
func CPMerge() ListMerger {
	return &cpMerge{}
}

type cpMerge struct{}

// Merge returns list of candidates, that appears at least `threshold` times.
func (cp *cpMerge) Merge(rid Rid, threshold int) ([]MergeCandidate, error) {
	lenRid := len(rid)

	if threshold > lenRid {
		return []MergeCandidate{}, nil
	}

	minQueries := lenRid - threshold + 1
	candidates := make([]MergeCandidate, 0, lenRid)
	tmp := make([]MergeCandidate, 0, lenRid)
	j, endMergeCandidate := 0, 0

	sort.Sort(rid)

	for _, list := range rid[:minQueries] {
		isValid := true
		current, err := list.Get()

		if err != nil {
			if err == ErrIteratorIsNotDereferencable {
				isValid = false
			} else {
				return nil, err
			}
		}

		tmp = tmp[:0]
		j, endMergeCandidate = 0, len(candidates)

		for j < endMergeCandidate || isValid {
			if j >= endMergeCandidate || (isValid && candidates[j].Position > current) {
				tmp = append(tmp, MergeCandidate{current, 1})

				if list.HasNext() {
					current, err = list.Next()

					if err != nil {
						return nil, err
					}
				} else {
					isValid = false
				}
			} else if !isValid || (j < endMergeCandidate && candidates[j].Position < current) {
				tmp = append(tmp, candidates[j])
				j++
			} else {
				candidates[j].Overlap++
				tmp = append(tmp, candidates[j])
				j++

				if list.HasNext() {
					current, err = list.Next()

					if err != nil {
						return nil, err
					}
				} else {
					isValid = false
				}
			}
		}

		candidates, tmp = tmp, candidates
	}

	if len(candidates) == 0 {
		return candidates, nil
	}

	for i := minQueries; i < lenRid; i++ {
		tmp = tmp[:0]

		for _, c := range candidates {
			current, err := rid[i].LowerBound(c.Position)

			if err != nil && err != ErrIteratorIsNotDereferencable {
				return nil, err
			}

			if err != ErrIteratorIsNotDereferencable {
				if current == c.Position {
					c.Overlap++
				}
			}

			if c.Overlap+(lenRid-i-1) >= threshold {
				tmp = append(tmp, c)
			}
		}

		candidates, tmp = tmp, candidates

		if len(candidates) == 0 {
			break
		}
	}

	return candidates, nil
}