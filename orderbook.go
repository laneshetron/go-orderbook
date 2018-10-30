package orderbook

import "container/heap"

type Order struct {
    price float64
    quantity float64
    orderId string
    index int
}

type Quote struct {
    ask *Order
    bid *Order
}

type TradeEvent struct {
    price float64
    quantity float64
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
    return ob.BaseHeap[i].price < ob.BaseHeap[j].price
}

func (ob BidOrders) Less(i, j int) bool {
    return ob.BaseHeap[i].price > ob.BaseHeap[j].price
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
    BidOrders
    OrdersMap
}

func (bb *BidBook) Push(orderId string, price float64, amount float64) {
    bb.remove(orderId) // ensure orderId does not already exist
    order := Order{price: price, quantity: amount, orderId: orderId}
    heap.Push(&bb.BidOrders, &order)
    bb.OrdersMap[orderId] = &order
}

func (bb *BidBook) Pop() *Order {
    order := heap.Pop(&bb.BidOrders).(*Order)
    delete(bb.OrdersMap, order.orderId)
    return order
}

func (bb *BidBook) remove(orderId string) {
    if _, ok := bb.OrdersMap[orderId]; ok {
        heap.Remove(&bb.BidOrders, bb.OrdersMap[orderId].index)
        delete(bb.OrdersMap, orderId)
    }
}

func (bb *BidBook) volume() float64 {
    var total float64 = 0
    for _, order := range bb.BidOrders.BaseHeap {
        total += order.quantity
    }
    return total
}

type AskBook struct {
    AskOrders
    OrdersMap
}

func (ab *AskBook) Push(orderId string, price float64, amount float64) {
    ab.remove(orderId) // ensure orderId does not already exist
    order := Order{price: price, quantity: amount, orderId: orderId}
    heap.Push(&ab.AskOrders, &order)
    ab.OrdersMap[orderId] = &order
}

func (ab *AskBook) Pop() *Order {
    order := heap.Pop(&ab.AskOrders).(*Order)
    delete(ab.OrdersMap, order.orderId)
    return order
}

func (ab *AskBook) remove(orderId string) {
    if _, ok := ab.OrdersMap[orderId]; ok {
        heap.Remove(&ab.AskOrders, ab.OrdersMap[orderId].index)
        delete(ab.OrdersMap, orderId)
    }
}

func (ab *AskBook) volume() float64 {
    var total float64 = 0
    for _, order := range ab.AskOrders.BaseHeap {
        total += order.quantity
    }
    return total
}

type OrderBook struct {
    AskBook
    BidBook
    quotes chan *Quote
    buyEvents chan *TradeEvent
    sellEvents chan *TradeEvent
}

func (ob *OrderBook) Init() {
    heap.Init(&ob.AskBook.AskOrders)
    heap.Init(&ob.BidBook.BidOrders)
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
    return (float64(ob.AskOrders.BaseHeap[0].price) +
            float64(ob.BidOrders.BaseHeap[0].price)) / 2
}

func (ob OrderBook) Spread() float64 {
    if !ob.HasBoth() {
        return 0
    }
    return (float64(ob.AskOrders.BaseHeap[0].price) -
            float64(ob.BidOrders.BaseHeap[0].price))
}

func (ob OrderBook) HasBoth() bool {
    return ob.AskOrders.Len() > 0 && ob.BidOrders.Len() > 0
}

func (ob OrderBook) Volume() float64 {
    return ob.AskBook.volume() + ob.BidBook.volume()
}
