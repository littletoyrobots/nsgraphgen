package graphgen

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/emicklei/dot"
)

var NodeTypes = []string{
	"Unknown",
	"AuthAction",
	"AuthPolicy",
	"AuthVServer",
	"Cert",
	"CSAction",
	"CSPolicy",
	"CSVServer",
	"DomainName",
	"GSLBService",
	"GSLBGroup",
	"GSLBVServer",
	"LBGroup",
	"LBVServer",
	"Netscaler",
	"Policy",
	"PolicyLabel",
	"PortalTheme",
	"ResponderAction",
	"ResponderPolicy",
	"RewriteAction",
	"RewritePolicy",
	"Server",
	"Service",
	"ServiceGroup",
	"SessionAction",
	"SessionPolicy",
	"STA",
	"VPNVServer",
	"WI",
	"VIP",
}

type nsNode struct {
	nstype      string
	name        string
	ip          string
	port        string
	protocol    string
	label       string
	isolated    bool
	highlighted bool
}
type nsEdge struct {
	from     string
	to       string
	port     string
	protocol string
	label    string
}

type NSGraph struct {
	Rankdir       string
	IgnoreNames   []string
	IgnoreTypes   []string
	IsolatedNames []string
	Nodes         []nsNode
	Edges         []nsEdge
	Graph         *dot.Graph
}

func isIPAddress(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil
}

func isValidNSType(str string) bool {
	for _, v := range NodeTypes {
		if strings.EqualFold(str, v) {
			return true
		}
	}
	return false
}

func makeEdgeLabel(port, protocol string) string {
	if port != "" && protocol != "" {
		return fmt.Sprintf("%s | %s", port, protocol)
	}
	if port != "" {
		return port
	}
	return protocol
}

func makeNodeLabel(name, ip string) string {
	if name != "" && ip != "" {
		return fmt.Sprintf("%s | %s", name, ip)
	}
	if name != "" {
		return name
	}
	return ip
}

func New(rankDir string, ignoreNames []string, ignoreTypes []string, isolateNames []string) *NSGraph {
	for _, v := range ignoreNames {
		slog.Info("adding to ignore list", "name", v)
	}
	for _, v := range ignoreTypes {
		slog.Info("adding to ignore list", "type", v)
	}
	for _, v := range isolateNames {
		slog.Info("adding to isolation list", "name", v)
	}
	ns := &NSGraph{
		Rankdir:       rankDir,
		IgnoreNames:   ignoreNames,
		IgnoreTypes:   ignoreTypes,
		IsolatedNames: isolateNames,
		Graph:         nil,
	}
	ns.Nodes = []nsNode{}
	ns.Edges = []nsEdge{}
	// ns.Graph = dot.NewGraph(dot.Directed)

	return ns
}

func (ns *NSGraph) Parse(inputFile string) error {
	re := regexp.MustCompile(`"((?:\\.|[^"\\])*)"|(\S+)`)

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	slog.Debug("adding global (0.0.0.0) node")
	ns.addNode("VIP", "Global", "0.0.0.0", "", "")

	for scanner.Scan() {
		line := scanner.Text()
		// term := re.Split(line, -1)
		term := re.FindAllString(line, -1)
		for i, v := range term {
			term[i] = strings.Trim(v, "\"")
		}
		ns.parseNSline(line)
	}

	ns.updateEdges()
	ns.pruneIgnored()
	ns.pruneNonIsolated()
	// ns.pruneNonIsolatedOld()
	slog.Info("parse complete")
	return nil
}

func (ns *NSGraph) pruneIgnored() {
	slog.Info("pruning ignored names and types")
	if len(ns.IgnoreNames) < 1 && len(ns.IgnoreTypes) < 1 {
		slog.Debug("nothing to ignore. skipping")
		return
	}
	newNodes := []nsNode{}
	newEdges := []nsEdge{}

	for _, n := range ns.Nodes {
		if slices.Contains(ns.IgnoreNames, n.name) {
			slog.Debug("ignoring node by name", "node", n)
			continue
		}
		if slices.Contains(ns.IgnoreNames, n.ip) {
			slog.Debug("ignoring node by ip", "node", n)
			continue
		}
		if slices.Contains(ns.IgnoreNames, n.label) {
			slog.Debug("ignoring node by named label", "node", n)
			continue
		}

		if slices.Contains(ns.IgnoreTypes, n.nstype) {
			slog.Debug("ignoring node by type", "node", n)
			continue
		}
		newNodes = append(newNodes, n)
	}
	ns.Nodes = newNodes

	for _, e := range ns.Edges {
		if slices.Contains(ns.IgnoreNames, e.from) {
			slog.Debug("ignoring edge by from node", "edge", e)
			continue
		}
		if slices.Contains(ns.IgnoreNames, e.to) {
			slog.Debug("ignoring edge by to node", "edge", e)
			continue
		}
		newEdges = append(newEdges, e)
	}

	ns.Edges = newEdges
}

func (ns *NSGraph) pruneNonIsolated() {
	slog.Info("pruning non-isolated nodes")
	if len(ns.IsolatedNames) < 1 {
		slog.Info("no names to isolate. skipping")
		return
	}

	markedNames := ns.markIsolated(ns.IsolatedNames, true)

	if len(markedNames) < 1 {
		slog.Warn("no matching names to isolate", "isolate", ns.IsolatedNames)
		return
	}

	newNodes := []nsNode{}
	newEdges := []nsEdge{}

	markedNamesTo := []string{}
	markedNamesFrom := []string{}

	for _, name := range markedNames {
		//isolatedLabelsTo := []string{}
		// isolatedLabelsTo = append(isolatedLabelsTo, name)
		toQueue := []string{name}
		for {
			if len(toQueue) < 1 {
				break
			}

			for _, e := range ns.Edges {
				if slices.Contains(toQueue, e.from) {
					if !slices.Contains(newEdges, e) {
						newEdges = append(newEdges, e)
					}
					if !slices.Contains(markedNamesTo, e.to) {
						markedNamesTo = append(markedNamesTo, e.to)
					}
				}
			}
			toQueue = ns.markIsolated(markedNamesTo, false)
		}

		fromQueue := []string{name}
		for {
			if len(fromQueue) < 1 {
				break
			}

			for _, e := range ns.Edges {
				if slices.Contains(fromQueue, e.to) {
					if !slices.Contains(newEdges, e) {
						newEdges = append(newEdges, e)
					}
					if !slices.Contains(markedNamesFrom, e.from) {
						markedNamesFrom = append(markedNamesFrom, e.from)
					}
				}
			}
			fromQueue = ns.markIsolated(markedNamesFrom, false)
		}
	}

	markedNames = append(markedNames, markedNamesTo...)
	markedNames = append(markedNames, markedNamesFrom...)

	for _, n := range ns.Nodes {
		if slices.Contains(markedNames, n.label) {
			if !slices.Contains(newNodes, n) {
				newNodes = append(newNodes, n)
			}
		}
	}

	ns.Nodes = newNodes
	ns.Edges = newEdges
}

func (ns *NSGraph) markIsolated(isolatedNames []string, highlight bool) []string {
	newIsolated := []string{}
	for _, name := range isolatedNames {
		idx := -1
		nodeIdx := ns.getNodeIndex(name)

		if nodeIdx != nil {
			idx = *nodeIdx
		} else {
			slog.Warn("could not find target to isolate", "name", name)
		}

		if idx != -1 {
			if ns.Nodes[idx].isolated {
				continue
			}

			ns.Nodes[idx].isolated = true
			slog.Debug("isolating", "node", ns.Nodes[idx])
			if highlight {
				ns.Nodes[idx].highlighted = true
				slog.Debug("highlighting isolated node", "node", ns.Nodes[idx])
			}
			newIsolated = append(newIsolated, ns.Nodes[idx].label)
		}
	}
	return newIsolated
}

func (ns *NSGraph) addEdge(from, to, port, protocol string) {
	protocol = strings.ToUpper(protocol)
	if isIPAddress(from) {
		ns.addNode("Unknown", "", from, "", "")
	} else {
		ns.addNode("Unknown", from, "", "", "")
	}
	if isIPAddress(to) {
		ns.addNode("Unknown", "", to, "", "")
	} else {
		ns.addNode("Unknown", to, "", "", "")
	}

	label := makeEdgeLabel(port, protocol)
	e := nsEdge{
		from:     from,
		to:       to,
		port:     port,
		protocol: protocol,
		label:    label,
	}
	slog.Debug("add new", "edge", e)
	ns.Edges = append(ns.Edges, e)
}

func (ns *NSGraph) addNode(nstype, name, ip, port, protocol string) {

	if !isValidNSType(nstype) {
		log.Fatal("Add node: invalid type", "type", nstype)
	}
	if name == "" && ip == "" {
		log.Fatal("Add node, no name or ip")
	}
	protocol = strings.ToUpper(protocol)
	if isIPAddress(name) {
		ip = name
		name = ""
	}
	nameIdx := ns.getNodeIndex(name)
	ipIdx := ns.getNodeIndex(ip)
	idx := -1

	if nameIdx != nil {
		idx = *nameIdx
	} else if ipIdx != nil {
		idx = *ipIdx
	}

	if idx != -1 {
		if nstype != "Unknown" && ns.Nodes[idx].nstype == "Unknown" {
			ns.Nodes[idx].nstype = nstype
			slog.Debug("update nstype", "node", ns.Nodes[idx])
		}
		if nstype == "VIP" && ns.Nodes[idx].nstype == "Server" {
			ns.Nodes[idx].nstype = nstype
			slog.Debug("update nstype", "node", ns.Nodes[idx])
		}
		if name != "" && ns.Nodes[idx].name == "" {
			ns.Nodes[idx].ip = ip
			slog.Debug("update name", "node", ns.Nodes[idx])
		}
		if ip != "" && ns.Nodes[idx].ip == "" {
			ns.Nodes[idx].ip = ip
			slog.Debug("update ip", "node", ns.Nodes[idx])
		}
		if port != "" && ns.Nodes[idx].port == "" {
			ns.Nodes[idx].port = port
			slog.Debug("update port", "node", ns.Nodes[idx])
		}
		if protocol != "" && ns.Nodes[idx].protocol == "" {
			ns.Nodes[idx].protocol = protocol
			slog.Debug("update protocol", "node", ns.Nodes[idx])
		}

		// update label
		label := makeNodeLabel(ns.Nodes[idx].name, ns.Nodes[idx].ip)
		if label != ns.Nodes[idx].label {
			ns.Nodes[idx].label = label
			slog.Debug("update label", "node", ns.Nodes[idx])
		}
	} else {
		label := makeNodeLabel(name, ip)

		n := nsNode{
			nstype:   nstype,
			name:     name,
			ip:       ip,
			port:     port,
			protocol: protocol,
			label:    label,
		}
		slog.Debug("add new", "node", n)
		ns.Nodes = append(ns.Nodes, n)
	}

}

func (ns *NSGraph) getNodeIndex(target string) *int {
	if target == "" {
		return nil
	}

	for i, v := range ns.Nodes {
		if v.label == target || v.name == target || v.ip == target {
			idx := i
			return &idx
		}
	}

	return nil

	/*
		idx := ns.getNodeIndexByLabel(target)
		if idx != nil {
			return idx
		}
		idx = ns.getNodeIndexByName(target)
		if idx != nil {
			return idx
		}

		return ns.getNodeIndexByIP(target)
	*/
}

func (ns *NSGraph) updateEdges() {
	slog.Debug("update edges")
	fromIdx := -1
	toIdx := -1
	for i, v := range ns.Edges {
		from_idx := ns.getNodeIndex(v.from)
		// from_ip_idx := ns.getNodeIndexByIP(v.from)
		if from_idx != nil {
			fromIdx = *from_idx
		} else {
			slog.Error("from node not found", "edge", v)
			continue
		}
		fromLabel := makeNodeLabel(ns.Nodes[fromIdx].name, ns.Nodes[fromIdx].ip)

		if v.port == "" && ns.Nodes[fromIdx].port != "" {
			ns.Edges[i].port = ns.Nodes[fromIdx].port
		}
		if v.protocol == "" && ns.Nodes[fromIdx].protocol != "" {
			ns.Edges[i].protocol = ns.Nodes[fromIdx].protocol
		}

		to_idx := ns.getNodeIndex(v.to)
		if to_idx != nil {
			toIdx = *to_idx
		} else {
			slog.Error("to node not found", "edge", v)
			continue
		}

		to_label := makeNodeLabel(ns.Nodes[toIdx].name, ns.Nodes[toIdx].ip)

		ns.Edges[i].from = fromLabel
		ns.Edges[i].to = to_label
		ns.Edges[i].label = makeEdgeLabel(ns.Edges[i].port, ns.Edges[i].protocol)
	}
}
