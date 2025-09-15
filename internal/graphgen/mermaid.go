package graphgen

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/emicklei/dot"
)

type mermaidNodeAttribute struct {
	nstype string
	//	color  string
	shape string
	style string
}

type mermaidEdgeAttribute struct {
	value string
	color string
}

var mermaidEdgeAttrs = []mermaidEdgeAttribute{
	{value: "*", color: "red"}, // any
	{value: "ANY", color: "red"},
	{value: "25", color: "darkorange"}, // SMTP
	{value: "SMTP", color: "darkorange"},
	{value: "HTTP", color: "red"},
	{value: "443", color: "green"}, // HTTPS
	{value: "HTTPS", color: "green"},
	{value: "SSL", color: "green"},
	{value: "SSL_TCP", color: "green"},
	{value: "SAML", color: "green"},
	{value: "53", color: "hotpink"}, // DNS
	{value: "DNS", color: "hotpink"},
	{value: "389", color: "orange"}, // LDAP
	{value: "LDAP", color: "orange"},
	{value: "636", color: "navy"},     // LDAPS
	{value: "1812", color: "magenta"}, // RADIUS
	{value: "RADIUS", color: "magenta"},
	{value: "CERT", color: "greenyellow"},
	{value: "STA", color: "cadetblue"},
	{value: "BASETHEME", color: "black"},
	{value: "LOGINSCHEMA", color: "violet"},
	{value: "NFACTOR", color: "pink"},
}

var mermaidNodeAttrs = []mermaidNodeAttribute{
	{nstype: "Unknown", shape: "stadium", style: "fill:#ff00ff"},
	{nstype: "AuthAction", shape: "trapezoid-alt", style: "fill:#ffcc99"},
	{nstype: "AuthPolicy", shape: "trapezoid", style: "fill:#ffcc99"},
	{nstype: "AuthVServer", shape: "hexagon", style: "fill:#ffcc99"},
	{nstype: "Cert", shape: "asymmetric", style: "fill:#00ff00"},
	{nstype: "CSAction", shape: "trapezoid-alt", style: "fill:#ffb366"},
	{nstype: "CSPolicy", shape: "trapezoid", style: "fill:#ffb366"},
	{nstype: "CSVServer", shape: "hexagon", style: "fill:#ffb366"},
	{nstype: "DomainName", shape: "stadium", style: "fill:#ff00ff"},
	{nstype: "GSLBService", shape: "box", style: "fill:#66ccff"},
	{nstype: "GSLBGroup", shape: "box", style: "fill:#66ccff"},
	{nstype: "GSLBVServer", shape: "hexagon", style: "fill:#66ccff"},
	{nstype: "LBGroup", shape: "stadium", style: "fill:#ffff99"},
	{nstype: "LBVServer", shape: "hexagon", style: "fill:#ffff99"},
	{nstype: "Netscaler", shape: "circle", style: "fill:#00ffff"},
	{nstype: "Policy", shape: "folder", style: "fill:#ff99ff"},
	{nstype: "PolicyLabel", shape: "tab", style: "fill:#ff99ff"},
	{nstype: "PortalTheme", shape: "note", style: "fill:#33ccff"},
	{nstype: "ResponderAction", shape: "trapezoid-alt", style: "fill:#ffffcc"},
	{nstype: "ResponderPolicy", shape: "trapezoid", style: "fill:#ffffcc"},
	{nstype: "RewriteAction", shape: "trapezoid-alt", style: "fill:#99ff99"},
	{nstype: "RewritePolicy", shape: "trapezoid", style: "fill:#99ff99"},
	{nstype: "Server", shape: "box", style: "fill:#ffffff"},
	{nstype: "Service", shape: "box", style: "fill:f2f2f2"},
	{nstype: "ServiceGroup", shape: "box", style: "fill:#e6e6e6"},
	{nstype: "SessionAction", shape: "trapezoid-alt", style: "fill:#00ff99"},
	{nstype: "SessionPolicy", shape: "trapezoid", style: "fill:#00ff99"},
	{nstype: "STA", shape: "box", style: "fill:#66ccff"},
	{nstype: "VPNVServer", shape: "hexagon", style: "fill:#33ccff"},
	{nstype: "WI", shape: "hexagon", style: "fill:#33ccff"},
	{nstype: "VIP", shape: "circle", style: "fill:#ffff00"},
}

var mermaidHighlightColor = "magenta"

func (ns *NSGraph) ExportMermaid(outputFile string, stdout bool) {
	ns.Graph = dot.NewGraph(dot.Directed)

	for _, v := range ns.Nodes {

		attr := getMermaidNodeAttribute(v.nstype)
		style := ""
		// style := attr.style
		if v.highlighted {
			style = fmt.Sprintf("%s,stroke:%s", attr.style, mermaidHighlightColor)
		}
		ns.Graph.Node(v.label).Attr("shape", attr.shape).Attr("style", style)
	}
	for _, v := range ns.Edges {
		attr := getMermaidEdgeAttribute(v.port, v.protocol)
		from, found_from := ns.Graph.FindNodeById(v.from)
		if !found_from {
			slog.Debug("from node not found, skipping", "label", v.from)
			continue
		}
		to, found_to := ns.Graph.FindNodeById(v.to)
		if !found_to {
			slog.Debug("to node not found, skipping", "label", v.to)
			continue
		}
		ns.Graph.Edge(from, to, v.label).Attr("color", attr.color)
	}

	var orientation int
	switch ns.Rankdir {
	case "LR":
		orientation = dot.MermaidLeftToRight
	case "RL":
		orientation = dot.MermaidRightToLeft
	case "BT":
		orientation = dot.MermaidBottomToTop
	case "TB":
		orientation = dot.MermaidTopToBottom
	}

	if stdout {
		slog.Info("generating Mermaid export to STDOUT")
		fmt.Print(dot.MermaidFlowchart(ns.Graph, orientation))
	} else {
		slog.Info("generating Mermaid export", "output-file", outputFile)
		f, err := os.Create(outputFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		_, err = w.WriteString(dot.MermaidFlowchart(ns.Graph, orientation))
		if err != nil {
			panic(err)
		}
		w.Flush()
	}
}

func getMermaidNodeAttribute(nstype string) mermaidNodeAttribute {
	for _, v := range mermaidNodeAttrs {
		if v.nstype == nstype {
			return v
		}
	}

	// nstype: "Unknown",  shape: "stadium", style: "fill:#ff00ff"
	return mermaidNodeAttribute{nstype: "Unknown", shape: "stadium", style: "fill:#ff00ff"}
}

func getMermaidEdgeAttribute(port, protocol string) mermaidEdgeAttribute {
	for _, v := range mermaidEdgeAttrs {
		if v.value == port || v.value == protocol {
			return v
		}
	}
	return mermaidEdgeAttribute{value: "*", color: "black"}
}
