package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

//undirected weighted edge
type Edge struct {
	U, V, W int
}

type DisjointSet struct {
	parent []int
	rank   []int
}

//intialization
func NewDisjointSet(n int) *DisjointSet {
	p := make([]int, n)
	r := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
	}
	return &DisjointSet{p, r}
}

func (d *DisjointSet) Find(x int) int {
	if d.parent[x] != x {
		d.parent[x] = d.Find(d.parent[x])
	}
	return d.parent[x]
}

func (d *DisjointSet) Union(x, y int) bool {
	rx, ry := d.Find(x), d.Find(y)
	if rx == ry {
		return false
	}
	if d.rank[rx] < d.rank[ry] {
		d.parent[rx] = ry
	} else if d.rank[rx] > d.rank[ry] {
		d.parent[ry] = rx
	} else {
		d.parent[ry] = rx
		d.rank[rx]++
	}
	return true
}


func Kruskal(edges []Edge, n int) ([]Edge, int) {
	sort.Slice(edges, func(i, j int) bool { return edges[i].W < edges[j].W })

	dsu := NewDisjointSet(n)
	var mst []Edge
	total := 0

	for _, e := range edges {
		if dsu.Union(e.U, e.V) {
			mst = append(mst, e)
			total += e.W
			if len(mst) == n-1 {
				break
			}
		}
	}
	return mst, total
}

func EdgePartitionMST(edges []Edge, n int, mem int) ([]Edge, int) {
	rand.Seed(time.Now().UnixNano())
	current := edges

	for len(current) > mem {
		// spliting edges into l partitions
		l := int(math.Ceil(float64(len(current)) / float64(mem)))
		partitions := make([][]Edge, l)
		for _, e := range current {
			pid := rand.Intn(l)
			partitions[pid] = append(partitions[pid], e)
		}

		// Local MSTs
		var merged []Edge
		var wg sync.WaitGroup
		ch := make(chan []Edge, len(partitions))

		for i, part := range partitions {
			if len(part) == 0 {
				continue
			}
			wg.Add(1)
			go func(pid int, p []Edge){
				defer wg.Done()
				fmt.Printf("→ Machine %d computing MST on %d edges\n", pid, len(p))
				localMST, _ := Kruskal(p, n);
				ch <- localMST
			}(i, part)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for mst := range ch {
			merged = append(merged, mst...)
		}

		//Deduplicating edges
		unique := map[[3]int]bool{}
		var newEdges []Edge
		for _, e := range merged {
			key := [3]int{min(e.U, e.V), max(e.U, e.V), e.W}
			if !unique[key] {
				unique[key] = true
				newEdges = append(newEdges, e)
			}
		}

		current = newEdges
		fmt.Printf("Iteration done, edges reduced to %d\n", len(current))
	}

	finalMST, total := Kruskal(current, n)
	return finalMST, total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	edges := []Edge{
		{0, 1, 4},
		{0, 2, 3},
		{1, 2, 2},
		{1, 3, 7},
		{2, 3, 4},
		{2, 4, 6},
		{3, 4, 1},
	}
	n := 5 // number of vertices
	mem := 4 // max edges each machine can hold (η)

	fmt.Println("Initial edges:", len(edges))
	mst, total := EdgePartitionMST(edges, n, mem)

	fmt.Println("\nFinal MST edges (u v w):")
	for _, e := range mst {
		fmt.Printf("%d %d %d\n", e.U, e.V, e.W)
	}
	fmt.Println("Total MST weight-go =", total)
}
