package clientB

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Node struct {
	Name string

	ReqNum   int
	TaskNum  int   // 任务数
	ColdTime int64 //unix time 表示这个节点在这个时间前都不允许使用
}

// TODO: 锁争用太严重了，几乎无并发
// Fix: pprof看几乎没有争用，应该是nodes太少，速度很快。
type WeightedRandom struct {
	Nodes []*Node
	//currentIndex int
	mu sync.Mutex
}

// 仅用于排序
// ByReqNum 实现sort.Interface接口
type ByReqNum []*Node

// Len 返回切片的长度
func (a ByReqNum) Len() int { return len(a) }

// Swap 交换切片中的两个元素
func (a ByReqNum) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less 比较两个元素的ReqNum大小，从小到大排序
func (a ByReqNum) Less(i, j int) bool { return a[i].ReqNum < a[j].ReqNum }

// 仅用于排序
// ByReqNum 实现sort.Interface接口
type ByTaskNum []*Node

// Len 返回切片的长度
func (a ByTaskNum) Len() int { return len(a) }

// Swap 交换切片中的两个元素
func (a ByTaskNum) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less 比较两个元素的ReqNum大小，从小到大排序
func (a ByTaskNum) Less(i, j int) bool { return a[i].TaskNum < a[j].TaskNum }

func NewWeightedRandom() *WeightedRandom {
	return &WeightedRandom{Nodes: make([]*Node, 0)}
}

func (wr *WeightedRandom) UpdateNodeReqNum(name string, num int, taskNum int) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	exist := false
	idx := -1
	for i, node := range wr.Nodes {
		if node.Name == name {
			exist = true
			idx = i
			break
		}
	}
	if exist {
		wr.Nodes[idx].ReqNum += num
		wr.Nodes[idx].TaskNum += taskNum
	} else {
		node := &Node{Name: name, ColdTime: time.Now().Unix(), ReqNum: num}
		wr.Nodes = append(wr.Nodes, node)
	}
}

// 使用coldtime的插入方法
func (wr *WeightedRandom) UpdateNodeColdTime(name string, coldTime int64) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	exist := false
	idx := -1
	for i, node := range wr.Nodes {
		if node.Name == name {
			exist = true
			idx = i
			break
		}
	}
	if exist {
		wr.Nodes[idx].ColdTime = coldTime
	} else {
		node := &Node{Name: name, ColdTime: coldTime}
		wr.Nodes = append(wr.Nodes, node)
	}
}

func (wr *WeightedRandom) Choose() string {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	if len(wr.Nodes) == 0 {
		return ""
	}
	//chosenNode := wr.RoundRobinChoice()
	chosenNode := wr.MinReqNumChoice()
	if chosenNode == nil {
		return ""
	}

	return chosenNode.Name
}

// 简单轮询
var rbIdx int = 0

func (wr *WeightedRandom) RoundRobinChoice() *Node {
	count := 0
	for count < len(wr.Nodes) {
		rbIdx = (rbIdx + 1) % len(wr.Nodes)
		if wr.Nodes[rbIdx].ColdTime > time.Now().Unix() {
			count++
			continue
		}
		return wr.Nodes[rbIdx]
	}
	//log.Warn().Msg("randomChoice失败, 可能是全部节点冷却")
	return nil
}

// 先保证大家都有活干吧
func (wr *WeightedRandom) TaskbiggerThan40Choice() *Node {
	sort.Sort(ByTaskNum(wr.Nodes))
	for i := 0; i < len(wr.Nodes); i++ {
		if wr.Nodes[i].TaskNum > 40 {
			break
		}
		return wr.Nodes[i]
	}

	return wr.MinReqNumChoice()
}

func (wr *WeightedRandom) MinReqNumChoice() *Node {
	sort.Sort(ByReqNum(wr.Nodes))
	for i := 0; i < len(wr.Nodes); i++ {
		if wr.Nodes[i].ColdTime > time.Now().Unix() {
			continue
		}
		return wr.Nodes[i]
	}
	//log.Warn().Msg("randomChoice失败, 可能是全部节点冷却")
	return nil
}

// 无锁
func (wr *WeightedRandom) randomChoiceNew() *Node {
	r := rand.Intn(len(wr.Nodes))
	count := 0

	for i := r; count < len(wr.Nodes); i++ {
		curIndex := i % len(wr.Nodes)
		if wr.Nodes[curIndex].ColdTime > time.Now().Unix() {
			count++
			continue
		}
		return wr.Nodes[curIndex]
	}
	log.Warn().Msg("randomChoice失败, 可能是全部节点冷却")
	return nil
}

func (wr *WeightedRandom) Clear() {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.Nodes = make([]*Node, 0)
}

func (wr *WeightedRandom) Reset(cl map[string]*ClientPool) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.Nodes = make([]*Node, 0)
	for nodeName := range cl {
		node := &Node{Name: nodeName, ColdTime: time.Now().Unix(), ReqNum: 0}
		wr.Nodes = append(wr.Nodes, node)
	}
}

func (wr *WeightedRandom) CopyNodes() []Node {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	ret := make([]Node, len(wr.Nodes))
	for i := 0; i < len(wr.Nodes); i++ {
		ret[i] = *wr.Nodes[i]
	}
	return ret
}

// func (c *Client) ResetWeightedRandomNew() {
// 	c.Wr.Clear()

// 	// 小心死锁
// 	c.CpMutex.Lock()
// 	defer c.CpMutex.Unlock()
// 	for nodeName := range c.ClientPools {
// 		c.Wr.UpdateNodeColdTime(nodeName, time.Now().Unix())
// 	}
// }

func (c *Client) ResetWr() {
	// 小心死锁
	c.CpMutex.Lock()
	defer c.CpMutex.Unlock()
	c.Wr.Reset(c.ClientPools)
}
