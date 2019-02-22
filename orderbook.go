package orderbook

import "container/heap"

type Order struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	OrderId  string  `json:"orderId"`
	index    int
}

type Quote struct {
	Ask *Order `json:"ask,omitempty"`
	Bid *Order `json:"bid,omitempty"`
}

type TradeEvent struct {
	Price    float64
	Quantity float64
}

type BaseHeap []*Order
type AskOrders struct {
	BaseHeap
}
type BidOrders struct {
	BaseHeap
}
type OrdersMap map[string]*Order

func (ob AskOrders) Less(i, j int) bool {
	return ob.BaseHeap[i].Price < ob.BaseHeap[j].Price
}

func (ob BidOrders) Less(i, j int) bool {
	return ob.BaseHeap[i].Price > ob.BaseHeap[j].Price
}

func (h BaseHeap) Len() int { return len(h) }

func (h BaseHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *BaseHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
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
		return bb.Orders.BaseHeap[0]
	} else {
		return nil
	}
}

func (bb *BidBook) Push(orderId string, price float64, amount float64) {
	bb.Remove(orderId) // ensure orderId does not already exist
	order := Order{Price: price, Quantity: amount, OrderId: orderId}
	heap.Push(&bb.Orders, &order)
	bb.OrdersMap[orderId] = &order
}

func (bb *BidBook) Pop() *Order {
	order := heap.Pop(&bb.Orders).(*Order)
	delete(bb.OrdersMap, order.OrderId)
	return order
}

func (bb *BidBook) Remove(orderId string) {
	if _, ok := bb.OrdersMap[orderId]; ok {
		heap.Remove(&bb.Orders, bb.OrdersMap[orderId].index)
		delete(bb.OrdersMap, orderId)
	}
}

func (bb *BidBook) volume() float64 {
	var total float64 = 0
	for _, order := range bb.Orders.BaseHeap {
		total += order.Quantity
	}
	return total
}

type AskBook struct {
	Orders AskOrders
	OrdersMap
}

func (ab *AskBook) Peek() *Order {
	if ab.Orders.Len() > 0 {
		return ab.Orders.BaseHeap[0]
	} else {
		return nil
	}
}

func (ab *AskBook) Push(orderId string, price float64, amount float64) {
	ab.Remove(orderId) // ensure orderId does not already exist
	order := Order{Price: price, Quantity: amount, OrderId: orderId}
	heap.Push(&ab.Orders, &order)
	ab.OrdersMap[orderId] = &order
}

func (ab *AskBook) Pop() *Order {
	order := heap.Pop(&ab.Orders).(*Order)
	delete(ab.OrdersMap, order.OrderId)
	return order
}

func (ab *AskBook) Remove(orderId string) {
	if _, ok := ab.OrdersMap[orderId]; ok {
		heap.Remove(&ab.Orders, ab.OrdersMap[orderId].index)
		delete(ab.OrdersMap, orderId)
	}
}

func (ab *AskBook) volume() float64 {
	var total float64 = 0
	for _, order := range ab.Orders.BaseHeap {
		total += order.Quantity
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
