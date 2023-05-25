package main

import (
	"container/list"
	"fmt"
	"math"
)

type transport struct {
	supply, demand []int
	costs          [][]float64
	mx             [][]shipment
}

type shipment struct {
	q, unitCost float64
	i, j        int
}

var counter int = 0

func main() {
	t := initializeProblem()
	t.northWest()
	t.result()
	t.ssm()
	t.result()

}

func initializeProblem() *transport {
	src := []int{60, 62, 55, 65, 58, 50, 50, 50, 50, 50}
	dst := []int{50, 50, 50, 50, 50, 58, 61, 63, 60, 58}
	costs := [][]float64{
		{0, 1, 2, 3, 4, 7, 6, 5, 4, 3},
		{1, 0, 1, 2, 3, 6, 5, 4, 3, 2},
		{2, 3, 0, 1, 2, 4, 3, 2, 1, 2},
		{3, 2, 1, 0, 1, 8, 7, 4, 3, 6},
		{4, 3, 2, 1, 0, 2, 1, 3, 4, 2},
		{7, 6, 4, 8, 2, 0, 4, 3, 2, 1},
		{6, 5, 3, 7, 1, 4, 0, 4, 3, 2},
		{5, 4, 2, 4, 3, 3, 4, 0, 4, 3},
		{4, 3, 1, 3, 4, 2, 3, 4, 0, 4},
		{3, 2, 2, 6, 2, 1, 2, 3, 4, 0},
	}

	/*
		src := []int{10, 12, 5, 15, 8, 0, 0, 0, 0, 0}
			dst := []int{0, 0, 0, 0, 0, 8, 11, 13, 10, 8}
			costs := [][]float64{
				{0, 1, 2, 3, 4, 7, 6, 5, 4, 3},
				{1, 0, 1, 2, 3, 6, 5, 4, 3, 2},
				{2, 3, 0, 1, 2, 4, 3, 2, 1, 2},
				{3, 2, 1, 0, 1, 8, 7, 4, 3, 6},
				{4, 3, 2, 1, 0, 2, 1, 3, 4, 2},
				{7, 6, 4, 8, 2, 0, 4, 3, 2, 1},
				{6, 5, 3, 7, 1, 4, 0, 4, 3, 2},
				{5, 4, 2, 4, 3, 3, 4, 0, 4, 3},
				{4, 3, 1, 3, 4, 2, 3, 4, 0, 4},
				{3, 2, 2, 6, 2, 1, 2, 3, 4, 0},
			}
	*/

	matrix := make([][]shipment, len(src))
	for i := 0; i < len(src); i++ {
		matrix[i] = make([]shipment, len(dst))
	}
	return &transport{src, dst, costs, matrix}
}

func (t *transport) northWest() {
	r, northwest := 0, 0
	for r < len(t.supply) {
		c := northwest
		for c < len(t.demand) {
			quantity := t.supply[r]
			if t.supply[r] > t.demand[c] {
				quantity = t.demand[c]
			}
			if quantity > 0 {
				t.mx[r][c] = shipment{float64(quantity), t.costs[r][c], r, c}
				t.supply[r] -= quantity
				t.demand[c] -= quantity
				if t.supply[r] == 0 {
					northwest = c
					break
				}
			}
			c++
		}
		r++
	}
}

func (t *transport) ssm() {
	t.result()
	_, move, leaving := t.reduction()

	if move != nil {
		q := leaving.q
		plus := true
		for _, s := range move {
			if plus {
				s.q += q
			} else {
				s.q -= q
			}
			if s.q == 0 {
				t.mx[s.i][s.j] = shipment{}
			} else {
				t.mx[s.i][s.j] = s
			}
			plus = !plus
		}
		t.ssm()
	}

}

func (t *transport) reduction() (float64, []shipment, shipment) {
	maxReduction := 0.0
	var move []shipment = nil
	leaving := shipment{}
	t.degen()

	for r := 0; r < len(t.supply); r++ {
		for c := 0; c < len(t.demand); c++ {
			if (t.mx[r][c] != shipment{}) {
				continue
			}
			trial := shipment{0, t.costs[r][c], r, c}
			path := t.path(trial)

			reduction := 0.0
			lowestQuantity := float64(math.MaxInt32)
			leavingCandidate := shipment{}
			plus := true
			for _, s := range path {
				if plus {
					reduction += s.unitCost
				} else {
					reduction -= s.unitCost
					if s.q < lowestQuantity {
						leavingCandidate = s
						lowestQuantity = s.q
					}
				}
				plus = !plus
			}

			if reduction < maxReduction {
				move = path
				leaving = leavingCandidate
				maxReduction = reduction
			}
		}
	}

	return maxReduction, move, leaving
}

func (t *transport) path(s shipment) []shipment {

	path := list.New()
	for _, m := range t.mx {
		for _, s := range m {
			if (s != shipment{}) {
				path.PushBack(s)
			}
		}
	}
	path.PushFront(s)
	var next *list.Element
	for {
		removals := 0
		for e := path.Front(); e != nil; e = next {
			next = e.Next()
			shipments := t.findNeighbors(e.Value.(shipment), path)
			if (shipments[0] == shipment{} || shipments[1] == shipment{}) {
				path.Remove(e)
				removals++
			}
		}
		if removals == 0 {
			break
		}
	}

	shipments := make([]shipment, path.Len())
	prev := s
	for i := 0; i < len(shipments); i++ {
		shipments[i] = prev
		prev = t.findNeighbors(prev, path)[i%2]
	}
	return shipments
}

func (t *transport) findNeighbors(s shipment, lst *list.List) [2]shipment {
	var shipments [2]shipment
	e := lst.Front()
	for e != nil && (shipments[0] == shipment{} || shipments[1] == shipment{}) {
		o := e.Value.(shipment)
		if o != s {
			if (o.i == s.i && shipments[0] == shipment{}) {
				shipments[0] = o
			} else if (o.j == s.j && shipments[1] == shipment{}) {
				shipments[1] = o
			}
			if (shipments[0] != shipment{} && shipments[1] != shipment{}) {
				break
			}
		}
		e = e.Next()
	}
	return shipments
}

func (t *transport) degen() {
	eps := math.SmallestNonzeroFloat64
	list := list.New()
	for _, m := range t.mx {
		for _, s := range m {
			if (s != shipment{}) {
				list.PushBack(s)
			}
		}
	}

	if len(t.supply)+len(t.demand)-1 != list.Len() {
		i, i2 := 0, 0
		for i < len(t.supply) {
			i2 = 0
			for i2 < len(t.demand) {
				if (t.mx[i][i2] == shipment{}) {
					dummy := shipment{eps, t.costs[i][i2], i, i2}
					if len(t.path(dummy)) == 0 {
						t.mx[i][i2] = dummy
						return
					}
				}
				i2++
			}
			i++
		}
	}
}

func (t *transport) result() {
	if counter == 0 {
		fmt.Printf("Answer after NorthWest \n")
	} else if counter != 7 {
		fmt.Printf("Iteration %d \n", counter)
	} else {
		fmt.Printf("Final Answer \n")
	}

	totalCosts := 0.0
	for r := 0; r < len(t.supply); r++ {
		for c := 0; c < len(t.demand); c++ {
			s := t.mx[r][c]
			if (s != shipment{} && s.i == r && s.j == c) {
				fmt.Printf("%3d ", int(s.q))
				totalCosts += s.q * s.unitCost
			} else {
				fmt.Printf("  * ")
			}
		}
		fmt.Println()
	}
	fmt.Printf("\nCost: %g\n\n", totalCosts)
	counter++
}
