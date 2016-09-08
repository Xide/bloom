package bloom

import "fmt"

// ScalableFilter can handle situation where filter total number
// of elements is undetermined at instantiation
// Params:
// s: Growth rate, each new filter will be (size_prev_filter) * s
// P: false positive maximum probability
// m0: size of the initial filter
// r: tightning ratio, like 's' but for false positive precision
type ScalableFilter struct {
	s       float64
	p       float64
	m0      uint64
	r       float64
	filters []*Filter
}

// NewScalable Create a new ScalableFilter
func NewScalable(p float64, s float64, m0 uint64, r float64) *ScalableFilter {
	filts := make([]*Filter, 1)
	filts[0] = New(m0, hashCountForFP(p))
	return &ScalableFilter{
		s:       s,
		p:       p,
		m0:      m0,
		r:       r,
		filters: filts,
	}
}

// NewDefaultScalable create a new ScalableFilter with default arguments
// More details on arguments : http://gsd.di.uminho.pt/members/cbm/ps/dbloom.pdf
func NewDefaultScalable(p float64) *ScalableFilter {
	return NewScalable(p, 2.0, 1024, 0.8)
}

// Match : Check if s have an entry in the filter
// May return false positive
func (sbf *ScalableFilter) Match(s string) bool {
	for hid := len(sbf.filters) - 1; hid >= 0; hid-- {
		if sbf.filters[hid].Match(s) {
			return true
		}
	}
	return false
}

func (sbf *ScalableFilter) dumpsFilters() {
	fmt.Println("============Bloom_filter============")
	for i := 0; i < len(sbf.filters); i++ {
		fmt.Printf("[%3d] size: [%4d] k: [%3d] fr:[%.2f]\n",
			i, sbf.filters[i].Size, sbf.filters[i].k, sbf.filters[i].FillRatio())
	}
	fmt.Println("====================================")

}

// Feed : Add an entry in the scalable bloom filter
func (sbf *ScalableFilter) Feed(s string) *ScalableFilter {
	// fmt.Printf("[R]: %.5f | [E]: %.5f\n",
	// sbf.filters[0].fillRatio(), sbf.filters[0].estimateFillRatio())
	if sbf.filters[0].EstimateFillRatio() > 0.3 {
		sbf.p *= sbf.r
		sbf.filters = append(make([]*Filter, 1), sbf.filters...)
		sbf.filters[0] = New(
			uint64(float64(sbf.filters[1].Size)*sbf.s),
			hashCountForFP(sbf.p))
		// sbf.dumpsFilters()
	}
	sbf.filters[0].Feed(s)
	return sbf
}
