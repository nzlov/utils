package utils

// 数据区间
type Range struct {
	Name string
	// 最大值
	Max int64
}

type Ranges []Range

// 获取名称
func (r Ranges) Check(t int64) string {
	for _, v := range r {
		if t > v.Max {
			continue
		}
		return v.Name
	}
	return r[len(r)-1].Name
}

// 获取索引
func (r Ranges) Index(n string) int {
	for i, v := range r {
		if n == v.Name {
			return i
		}
	}
	return -1
}

func (r Ranges) RangesSort(i, j string) bool {
	return r.Index(i) < r.Index(j)
}

func (r Ranges) Names() []string {
	ss := []string{}
	for _, v := range r {
		ss = append(ss, v.Name)
	}
	return ss
}
