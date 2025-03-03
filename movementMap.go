package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
)

type IMovementNode interface {
	IsListOfPanels() bool
	IsSelector() bool
	Next(IPanel) IPanel
	Previous(IPanel) IPanel
	CreateKeyBindings(IMovementNode, func(IPanel) KeyBinding, func(IPanel) KeyBinding)
	contains(IPanel) bool
	allPanels() []IPanel
	First() IPanel
	Last() IPanel
}

type PanelSequence struct {
	panelList     []IPanel
	panelPosition map[IPanel]int
}

var _ IMovementNode = &PanelSequence{}

func NewPanelSequence(panels ...IPanel) *PanelSequence {
	panelPosition := make(map[IPanel]int)
	for i := 0; i < len(panels); i++ {
		panelPosition[panels[i]] = i
	}
	return &PanelSequence{
		panelList:     panels,
		panelPosition: panelPosition,
	}
}

func (m *PanelSequence) IsListOfPanels() bool {
	return true
}

func (m *PanelSequence) IsSelector() bool {
	return false
}

// Panel list can only create keybindings to go back and forth
// between the panels in the list
// It will leave out the entry/exit keybindings
func (m *PanelSequence) CreateKeyBindings(root IMovementNode, kb func(IPanel) KeyBinding, rev_kb func(IPanel) KeyBinding) {
	for _, panel := range m.panelList {
		if len(kb(panel).KeyDefs) > 0 {
			m.keybindingHelper(panel, kb(panel), root.Next)
		}
		if len(rev_kb(panel).KeyDefs) > 0 {
			m.keybindingHelper(panel, rev_kb(panel), root.Previous)
		}
	}
}

func (m *PanelSequence) keybindingHelper(panel IPanel, kb KeyBinding, fn func(IPanel) IPanel) {
	newKb := kb
	newKb.Func = func() tea.Cmd {
		next := fn(panel)
		if next == nil {
			return nil
		}
		return func() tea.Msg {
			return FocusRequestMsg{
				RequestedPath: next.GetPath(),
				Relation:      Self,
			}
		}
	}
	panel.AddKeyBinding(&newKb)
}

func (m *PanelSequence) contains(panel IPanel) bool {
	_, ok := m.panelPosition[panel]
	return ok
}

func (m *PanelSequence) Next(panel IPanel) IPanel {
	index := m.panelPosition[panel]
	if index == len(m.panelList)-1 {
		return nil
	}
	return m.panelList[index+1]
}

func (m *PanelSequence) Previous(panel IPanel) IPanel {
	index := m.panelPosition[panel]
	if index == 0 {
		return nil
	}
	return m.panelList[index-1]
}

func (m *PanelSequence) First() IPanel {
	return m.panelList[0]
}

func (m *PanelSequence) Last() IPanel {
	return m.panelList[len(m.panelList)-1]
}

func (m *PanelSequence) allPanels() []IPanel {
	return m.panelList
}

type SelectorNode struct {
	chains        []IMovementNode
	chainMap      map[IPanel]IMovementNode
	chainPosition map[IPanel]int
	selectionFunc func([]IPanel) IPanel
	loopAround    bool
	sequential    bool
}

var _ IMovementNode = &SelectorNode{}

func NewLoopAroundTabMap(chains ...IMovementNode) *SelectorNode {
	k := NewVisibleSelectorNode(chains...)
	k.loopAround = true
	k.sequential = true
	return k
}

func NewVisibleSelectorNode(chains ...IMovementNode) *SelectorNode {
	chainMap := make(map[IPanel]IMovementNode)
	chainPosition := make(map[IPanel]int)
	for i := 0; i < len(chains); i++ {
		for _, panel := range chains[i].allPanels() {
			chainMap[panel] = chains[i]
			chainPosition[panel] = i
		}
	}
	return &SelectorNode{
		chains:        chains,
		chainMap:      chainMap,
		chainPosition: chainPosition,
		selectionFunc: visiblePaneSelectionLogic,
		loopAround:    false,
		sequential:    false,
	}
}

func (m *SelectorNode) IsListOfPanels() bool {
	return false
}

func (m *SelectorNode) IsSelector() bool {
	return true
}

func (m *SelectorNode) CreateKeyBindings(root IMovementNode, kb func(IPanel) KeyBinding, rev_kb func(IPanel) KeyBinding) {
	if root == nil {
		root = m
	}
	for _, chain := range m.chains {
		chain.CreateKeyBindings(root, kb, rev_kb)
	}
}

func (m *SelectorNode) Next(panel IPanel) IPanel {
	chain := m.chainMap[panel]
	chainIndex := m.chainPosition[panel]
	next := chain.Next(panel)
	if next == nil {
		// if next is nil, that means it is the last panel in the chain
		if !m.sequential {
			return nil // in sequential mode, we don't look at other chains
		}
		if chainIndex == len(m.chains)-1 {
			if m.loopAround {
				return m.First()
			}
			return nil
		}
		// if it is the last panel in the chain, we need to get the next chain
		nextChain := m.chains[chainIndex+1]
		return nextChain.First()
	}
	return next
}

func (m *SelectorNode) Previous(panel IPanel) IPanel {
	chain := m.chainMap[panel]
	chainIndex := m.chainPosition[panel]
	previous := chain.Previous(panel)
	if previous == nil {
		// if previous is nil, that means it is the first panel in the chain
		if !m.sequential {
			return nil // in sequential mode, we don't look at other chains
		}
		if chainIndex == 0 {
			if m.loopAround {
				return m.Last()
			}
			return nil
		}
		// if it is the first panel in the chain, we need to get the previous chain
		previousChain := m.chains[chainIndex-1]
		return previousChain.Last()
	}
	return previous
}

func (m *SelectorNode) First() IPanel {
	firstOptions := make([]IPanel, 0)
	if m.sequential {
		return m.chains[0].First()
	} else {
		for _, chain := range m.chains {
			firstOptions = append(firstOptions, chain.First())
		}
		return m.selectionFunc(firstOptions)
	}
}

func (m *SelectorNode) Last() IPanel {
	lastOptions := make([]IPanel, 0)
	if m.sequential {
		return m.chains[len(m.chains)-1].Last()
	} else {
		for _, chain := range m.chains {
			lastOptions = append(lastOptions, chain.Last())
		}
		return m.selectionFunc(lastOptions)
	}
}

func (m *SelectorNode) contains(panel IPanel) bool {
	_, ok := m.chainMap[panel]
	return ok
}

func (m *SelectorNode) allPanels() []IPanel {
	allPanels := make([]IPanel, 0)
	for _, chain := range m.chains {
		allPanels = append(allPanels, chain.allPanels()...)
	}
	return allPanels
}

func visiblePaneSelectionLogic(panels []IPanel) IPanel {
	for _, panel := range panels {
		if !panel.IsInHiddenTab() {
			return panel
		}
	}
	return nil
}
