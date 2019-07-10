package orderbook

import (
	"container/heap"
	"sync"
)

type Item interface {
	Peek() *Order
}

type Book interface {
	Item
	Push(*Node)
	Pop() *Node
	Remove(string)
	Fix(string)
	Len() int
}

type Node struct {
	Item
	Key    string
	Weight float64
	index  int
}

func NewNode(key string, i Item, weight float64) Node {
	return Node{
		Item:   i,
		Key:    key,
		Weight: weight,
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

func NewOrder(price float64, quantity float64, orderId string) Order {
	return Order{
		Price:    price,
		Quantity: quantity,
		OrderId:  orderId,
	}
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
	return ob.BaseHeap[i].Peek().Price*ob.BaseHeap[i].Weight <
		ob.BaseHeap[j].Peek().Price*ob.BaseHeap[j].Weight
}

func (ob BidOrders) Less(i, j int) bool {
	return ob.BaseHeap[i].Peek().Price*ob.BaseHeap[i].Weight >
		ob.BaseHeap[j].Peek().Price*ob.BaseHeap[j].Weight
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
	lock sync.Mutex
}

func (bb *BidBook) Peek() *Order {
	if bb.Len() > 0 {
		return bb.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (bb *BidBook) Len() int {
	return bb.Orders.Len()
}

func (bb *BidBook) Push(n *Node) {
	bb.Remove(n.Key) // ensure Key does not already exist
	bb.lock.Lock()
	defer bb.lock.Unlock()

	heap.Push(&bb.Orders, n)
	bb.OrdersMap[n.Key] = n
}

func (bb *BidBook) Pop() *Node {
	bb.lock.Lock()
	defer bb.lock.Unlock()

	node := heap.Pop(&bb.Orders).(*Node)
	delete(bb.OrdersMap, node.Key)
	return node
}

func (bb *BidBook) Remove(key string) {
	bb.lock.Lock()
	defer bb.lock.Unlock()

	if _, ok := bb.OrdersMap[key]; ok {
		heap.Remove(&bb.Orders, bb.OrdersMap[key].index)
		delete(bb.OrdersMap, key)
	}
}

func (bb *BidBook) Fix(key string) {
	bb.lock.Lock()
	defer bb.lock.Unlock()

	if _, ok := bb.OrdersMap[key]; ok {
		heap.Fix(&bb.Orders, bb.OrdersMap[key].index)
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
	lock sync.Mutex
}

func (ab *AskBook) Peek() *Order {
	if ab.Len() > 0 {
		return ab.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (ab *AskBook) Len() int {
	return ab.Orders.Len()
}

func (ab *AskBook) Push(n *Node) {
	ab.Remove(n.Key) // ensure Key does not already exist
	ab.lock.Lock()
	defer ab.lock.Unlock()

	heap.Push(&ab.Orders, n)
	ab.OrdersMap[n.Key] = n
}

func (ab *AskBook) Pop() *Node {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	node := heap.Pop(&ab.Orders).(*Node)
	delete(ab.OrdersMap, node.Key)
	return node
}

func (ab *AskBook) Remove(key string) {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	if _, ok := ab.OrdersMap[key]; ok {
		heap.Remove(&ab.Orders, ab.OrdersMap[key].index)
		delete(ab.OrdersMap, key)
	}
}

func (ab *AskBook) Fix(key string) {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	if _, ok := ab.OrdersMap[key]; ok {
		heap.Fix(&ab.Orders, ab.OrdersMap[key].index)
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
	return ob.AskBook.Len() > 0 && ob.BidBook.Len() > 0
}

func (ob OrderBook) Volume() float64 {
	return ob.AskBook.volume() + ob.BidBook.volume()
}
