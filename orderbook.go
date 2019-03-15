package orderbook

import "container/heap"

type Item interface {
	Peek() *Order
}

type Node struct {
	Item
	Key    string
	weight float64
	index  int
}

func NewNode(key string, i Item, weight float64) Node {
	return Node{
		Item:   i,
		Key:    key,
		weight: weight,
	}
}

type Order struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	OrderId  string  `json:"orderId"`
	Country  string  `json:"country"`
}

func (o *Order) Peek() *Order {
	return o
}

type Quote struct {
	Ask *Order `json:"ask,omitempty"`
	Bid *Order `json:"bid,omitempty"`
}

type TradeEvent struct {
	Price    float64
	Quantity float64
}

type BaseHeap []*Node
type AskOrders struct {
	BaseHeap
}
type BidOrders struct {
	BaseHeap
}
type OrdersMap map[string]*Node

func (ob AskOrders) Less(i, j int) bool {
	return ob.BaseHeap[i].Peek().Price < ob.BaseHeap[j].Peek().Price
}

func (ob BidOrders) Less(i, j int) bool {
	return ob.BaseHeap[i].Peek().Price > ob.BaseHeap[j].Peek().Price
}

func (h BaseHeap) Len() int { return len(h) }

func (h BaseHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *BaseHeap) Push(x interface{}) {
	*h = append(*h, x.(*Node))
	(*h)[len(*h)-1].index = len(*h) - 1
}

func (h *BaseHeap) Pop() interface{} {
	x := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return x
}

type BidBook struct {
	Orders BidOrders
	OrdersMap
}

func (bb *BidBook) Peek() *Order {
	if bb.Orders.Len() > 0 {
		return bb.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (bb *BidBook) Push(n *Node) {
	bb.Remove(n.Key) // ensure Key does not already exist
	heap.Push(&bb.Orders, n)
	bb.OrdersMap[n.Key] = n
}

func (bb *BidBook) Pop() *Node {
	node := heap.Pop(&bb.Orders).(*Node)
	delete(bb.OrdersMap, node.Key)
	return node
}

func (bb *BidBook) Remove(key string) {
	if _, ok := bb.OrdersMap[key]; ok {
		heap.Remove(&bb.Orders, bb.OrdersMap[key].index)
		delete(bb.OrdersMap, key)
	}
}

func (bb *BidBook) volume() float64 {
	var total float64 = 0
	for _, node := range bb.Orders.BaseHeap {
		total += node.Peek().Quantity
	}
	return total
}

type AskBook struct {
	Orders AskOrders
	OrdersMap
}

func (ab *AskBook) Peek() *Order {
	if ab.Orders.Len() > 0 {
		return ab.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (ab *AskBook) Push(n *Node) {
	ab.Remove(n.Key) // ensure Key does not already exist
	heap.Push(&ab.Orders, n)
	ab.OrdersMap[n.Key] = n
}

func (ab *AskBook) Pop() *Node {
	node := heap.Pop(&ab.Orders).(*Node)
	delete(ab.OrdersMap, node.Key)
	return node
}

func (ab *AskBook) Remove(key string) {
	if _, ok := ab.OrdersMap[key]; ok {
		heap.Remove(&ab.Orders, ab.OrdersMap[key].index)
		delete(ab.OrdersMap, key)
	}
}

func (ab *AskBook) volume() float64 {
	var total float64 = 0
	for _, node := range ab.Orders.BaseHeap {
		total += node.Peek().Quantity
	}
	return total
}

type OrderBook struct {
	AskBook
	BidBook
	quotes     chan *Quote
	buyEvents  chan *TradeEvent
	sellEvents chan *TradeEvent
}

func (ob *OrderBook) Init() {
	heap.Init(&ob.AskBook.Orders)
	heap.Init(&ob.BidBook.Orders)
	ob.AskBook.OrdersMap = make(OrdersMap)
	ob.BidBook.OrdersMap = make(OrdersMap)
	ob.quotes = make(chan *Quote)
	ob.buyEvents = make(chan *TradeEvent)
	ob.sellEvents = make(chan *TradeEvent)
}

func (ob OrderBook) Midpoint() float64 {
	if !ob.HasBoth() {
		return 0
	}
	return (float64(ob.AskBook.Peek().Price) + float64(ob.BidBook.Peek().Price)) / 2
}

func (ob OrderBook) Spread() float64 {
	if !ob.HasBoth() {
		return 0
	}
	return (float64(ob.AskBook.Peek().Price) - float64(ob.BidBook.Peek().Price))
}

func (ob OrderBook) HasBoth() bool {
	return ob.AskBook.Orders.Len() > 0 && ob.BidBook.Orders.Len() > 0
}

func (ob OrderBook) Volume() float64 {
	return ob.AskBook.volume() + ob.BidBook.volume()
}
