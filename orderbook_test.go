// Copyright 2019 Lane A. Shetron
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package orderbook

import (
	"fmt"
	"testing"
)

func TestAskBook(t *testing.T) {
	orders := []struct {
		Id    string
		Price float64
		Peek  float64
	}{
		{"a", 123.45, 123.45},
		{"b", 155.45, 123.45},
		{"c", 122.00, 122.00},
		{"d", 136.00, 122.00},
		{"e", 121.00, 121.00},
		{"f", 333.00, 121.00},
		{"g", 120.999, 120.999},
	}
	ob := NewOrderBook()
	for _, order := range orders {
		t.Run(fmt.Sprintf("%s-%f", order.Id, order.Price), func(t *testing.T) {
			o := NewOrder(order.Price, 1, order.Id)
			node := NewNode(order.Id, &o, 1)
			ob.AskBook.Push(&node)
			if ob.AskBook.Peek().Price != order.Peek {
				t.Errorf("Expected lowest ask %f, got %f", order.Peek, ob.AskBook.Peek().Price)
			}
		})
	}
	expected := []float64{120.999, 121.00, 122.00, 123.45, 136.00, 155.45, 333.00}
	for ob.AskBook.Len() > 0 {
		t.Run(fmt.Sprintf("next-lowest-%f", expected[0]), func(t *testing.T) {
			o := ob.AskBook.Pop().Peek()
			if o.Price != expected[0] {
				t.Errorf("Expected next lowest ask %f, got %f", expected[0], o.Price)
			}
			expected = expected[1:]
		})
	}
}

func TestBidBook(t *testing.T) {
	orders := []struct {
		Id    string
		Price float64
		Peek  float64
	}{
		{"a", 123.45, 123.45},
		{"b", 155.45, 155.45},
		{"c", 122.00, 155.45},
		{"d", 136.00, 155.45},
		{"e", 121.00, 155.45},
		{"f", 333.00, 333.00},
		{"g", 120.999, 333.00},
	}
	ob := NewOrderBook()
	for _, order := range orders {
		t.Run(fmt.Sprintf("%s-%f", order.Id, order.Price), func(t *testing.T) {
			o := NewOrder(order.Price, 1, order.Id)
			node := NewNode(order.Id, &o, 1)
			ob.BidBook.Push(&node)
			if ob.BidBook.Peek().Price != order.Peek {
				t.Errorf("Expected highest bid %f, got %f", order.Peek, ob.BidBook.Peek().Price)
			}
		})
	}
	expected := []float64{333.00, 155.45, 136.00, 123.45, 122.00, 121.00, 120.999}
	for ob.BidBook.Len() > 0 {
		t.Run(fmt.Sprintf("next-highest-%f", expected[0]), func(t *testing.T) {
			o := ob.BidBook.Pop().Peek()
			if o.Price != expected[0] {
				t.Errorf("Expected next highest bid %f, got %f", expected[0], o.Price)
			}
			expected = expected[1:]
		})
	}
}

func TestCopy(t *testing.T) {
	src := NewOrderBook()
	dst := NewOrderBook()

	ask := NewOrder(1234.0, 100, "a")
	bid := NewOrder(1232.0, 100, "b")
	node1 := NewNode("a", &ask, 1)
	node2 := NewNode("b", &bid, 1)
	src.AskBook.Push(&node1)
	src.BidBook.Push(&node2)

	Copy(src, dst)
	if dst.AskBook.Peek() == nil || dst.BidBook.Peek() == nil {
		t.Fatal("Expected src entries to be copied to dst OrderBook.")
	}
	dst.AskBook.Peek().Quantity -= 10
	dst.BidBook.Peek().Price -= 2
	if src.AskBook.Peek().Quantity == dst.AskBook.Peek().Quantity {
		t.Errorf("Expected source order to be unaltered. Expected %f, got %f", 100.0, src.AskBook.Peek().Quantity)
	}
	if src.BidBook.Peek().Price == dst.BidBook.Peek().Price {
		t.Errorf("Expected source order to be unaltered. Expected %f, got %f", 1232.0, src.BidBook.Peek().Price)
	}

	srcNode := src.AskBook.Pop()
	dstNode := dst.AskBook.Pop()
	dstNode.Weight = 2
	if srcNode.Weight == dstNode.Weight {
		t.Errorf("Expected source node weight to be unaltered. Expected %f, got %f", 1.0, srcNode.Weight)
	}
}
