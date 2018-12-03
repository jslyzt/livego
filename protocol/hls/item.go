package hls

// TSItem 节点
type TSItem struct {
	Name     string
	SeqNum   int
	Duration int
	Data     []byte
}

// NewTSItem 新节点
func NewTSItem(name string, duration, seqNum int, b []byte) TSItem {
	var item TSItem
	item.Name = name
	item.SeqNum = seqNum
	item.Duration = duration
	item.Data = make([]byte, len(b))
	copy(item.Data, b)
	return item
}
