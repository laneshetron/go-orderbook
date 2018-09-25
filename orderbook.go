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

type AskOrders []*Order
type BidOrders []*Order
type OrdersMap map[string]*Order

func (ob AskOrders) Len() int { return len(ob) }

func (ob BidOrders) Len() int { return len(ob) }

func (ob AskOrders) Less(i, j int) bool {
    return ob[i].price < ob[j].price
}

func (ob BidOrders) Less(i, j int) bool {
    return ob[i].price > ob[j].price
}

func (ob AskOrders) Swap(i, j int) {
    ob[i], ob[j] = ob[j], ob[i]
    ob[i].index = i
    ob[j].index = j
}

func (ob BidOrders) Swap(i, j int) {
    ob[i], ob[j] = ob[j], ob[i]
    ob[i].index = i
    ob[j].index = j
}

func (ob *AskOrders) Push(x interface{}) {
    *ob = append(*ob, x.(*Order))
    (*ob)[len(*ob)-1].index = len(*ob) - 1
}

func (ob *BidOrders) Push(x interface{}) {
    *ob = append(*ob, x.(*Order))
    (*ob)[len(*ob)-1].index = len(*ob) - 1
}

func (ob *AskOrders) Pop() interface{} {
    x := (*ob)[len(*ob)-1]
    *ob = (*ob)[:len(*ob)-1]
    return x
}

func (ob *BidOrders) Pop() interface{} {
    x := (*ob)[len(*ob)-1]
    *ob = (*ob)[:len(*ob)-1]
    return x
}

type BidBook struct {
    BidOrders
    OrdersMap
}

func (bb *BidBook) push(orderId string, price float64, amount float64) {
    bb.remove(orderId) // ensure orderId does not already exist
    order := Order{price: price, quantity: amount, orderId: orderId}
    heap.Push(&bb.BidOrders, &order)
    bb.OrdersMap[orderId] = &order
}

func (bb *BidBook) pop() *Order {
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
    for _, order := range bb.BidOrders {
        total += order.quantity
    }
    return total
}

type AskBook struct {
    AskOrders
    OrdersMap
}

func (ab *AskBook) push(orderId string, price float64, amount float64) {
    ab.remove(orderId) // ensure orderId does not already exist
    order := Order{price: price, quantity: amount, orderId: orderId}
    heap.Push(&ab.AskOrders, &order)
    ab.OrdersMap[orderId] = &order
}

func (ab *AskBook) pop() *Order {
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
    for _, order := range ab.AskOrders {
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
}

func (ob OrderBook) midpoint() float64 {
    if !ob.hasBoth() {
        return 0
    }
    return (float64(ob.AskOrders[0].price) +
            float64(ob.BidOrders[0].price)) / 2
}

func (ob OrderBook) spread() float64 {
    if !ob.hasBoth() {
        return 0
    }
    return (float64(ob.AskOrders[0].price) -
            float64(ob.BidOrders[0].price))
}

func (ob OrderBook) hasBoth() bool {
    return ob.AskOrders.Len() > 0 && ob.BidOrders.Len() > 0
}

func (ob OrderBook) volume() float64 {
    return ob.AskBook.volume() + ob.BidBook.volume()
}
