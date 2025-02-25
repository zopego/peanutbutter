package peanutbutter

import (
	tea "github.com/charmbracelet/bubbletea"
)

type IMovementNode interface {
	IsListOfPanels() bool
	IsSelector() bool
	Next(*ShortCutPanel) *ShortCutPanel
	Previous(*ShortCutPanel) *ShortCutPanel
	CreateKeyBindings(IMovementNode, *KeyBinding, *KeyBinding)
	contains(*ShortCutPanel) bool
	allPanels() []*ShortCutPanel
	First() *ShortCutPanel
	Last() *ShortCutPanel
}

type PanelSequence struct {
	panelList     []*ShortCutPanel
	panelPosition map[*ShortCutPanel]int
}

func NewPanelSequence(panels ...*ShortCutPanel) *PanelSequence {
	panelPosition := make(map[*ShortCutPanel]int)
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
func (m *PanelSequence) CreateKeyBindings(root IMovementNode, kb *KeyBinding, rev_kb *KeyBinding) {
	for _, panel := range m.panelList {
		if kb != nil {
			m.keybindingHelper(panel, kb, root.Next)
		}
		if rev_kb != nil {
			m.keybindingHelper(panel, rev_kb, root.Previous)
		}
	}
}

func (m *PanelSequence) keybindingHelper(panel *ShortCutPanel, kb *KeyBinding, fn func(*ShortCutPanel) *ShortCutPanel) {
	kb.Func = func() tea.Cmd {
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
	panel.AddKeyBinding(*kb)
}

func (m *PanelSequence) contains(panel *ShortCutPanel) bool {
	_, ok := m.panelPosition[panel]
	return ok
}

func (m *PanelSequence) Next(panel *ShortCutPanel) *ShortCutPanel {
	index := m.panelPosition[panel]
	if index == len(m.panelList)-1 {
		return nil
	}
	return m.panelList[index+1]
}

func (m *PanelSequence) Previous(panel *ShortCutPanel) *ShortCutPanel {
	index := m.panelPosition[panel]
	if index == 0 {
		return nil
	}
	return m.panelList[index-1]
}

func (m *PanelSequence) First() *ShortCutPanel {
	return m.panelList[0]
}

func (m *PanelSequence) Last() *ShortCutPanel {
	return m.panelList[len(m.panelList)-1]
}

func (m *PanelSequence) allPanels() []*ShortCutPanel {
	return m.panelList
}

type SelectorNode struct {
	chains        []IMovementNode
	chainMap      map[*ShortCutPanel]IMovementNode
	chainPosition map[*ShortCutPanel]int
	selectionFunc func([]*ShortCutPanel) *ShortCutPanel
	loopAround    bool
	sequential    bool
}

func NewLoopAroundTabMap(chains ...IMovementNode) *SelectorNode {
	k := NewVisibleSelectorNode(chains...)
	k.loopAround = true
	k.sequential = true
	return k
}

func NewVisibleSelectorNode(chains ...IMovementNode) *SelectorNode {
	chainMap := make(map[*ShortCutPanel]IMovementNode)
	chainPosition := make(map[*ShortCutPanel]int)
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

func (m *SelectorNode) CreateKeyBindings(root IMovementNode, kb *KeyBinding, rev_kb *KeyBinding) {
	if root == nil {
		root = m
	}
	for _, chain := range m.chains {
		chain.CreateKeyBindings(root, kb, rev_kb)
	}
}

func (m *SelectorNode) Next(panel *ShortCutPanel) *ShortCutPanel {
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

func (m *SelectorNode) Previous(panel *ShortCutPanel) *ShortCutPanel {
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

func (m *SelectorNode) First() *ShortCutPanel {
	firstOptions := make([]*ShortCutPanel, 0)
	if m.sequential {
		return m.chains[0].First()
	} else {
		for _, chain := range m.chains {
			firstOptions = append(firstOptions, chain.First())
		}
		return m.selectionFunc(firstOptions)
	}
}

func (m *SelectorNode) Last() *ShortCutPanel {
	lastOptions := make([]*ShortCutPanel, 0)
	if m.sequential {
		return m.chains[len(m.chains)-1].Last()
	} else {
		for _, chain := range m.chains {
			lastOptions = append(lastOptions, chain.Last())
		}
		return m.selectionFunc(lastOptions)
	}
}

func (m *SelectorNode) contains(panel *ShortCutPanel) bool {
	_, ok := m.chainMap[panel]
	return ok
}

func (m *SelectorNode) allPanels() []*ShortCutPanel {
	allPanels := make([]*ShortCutPanel, 0)
	for _, chain := range m.chains {
		allPanels = append(allPanels, chain.allPanels()...)
	}
	return allPanels
}

func visiblePaneSelectionLogic(panels []*ShortCutPanel) *ShortCutPanel {
	for _, panel := range panels {
		if !panel.IsInHiddenTab() {
			return panel
		}
	}
	return nil
}
