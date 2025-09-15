package graphgen

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/emicklei/dot"
)

type dotNodeAttribute struct {
	nstype    string
	fillcolor string
	shape     string
	style     string
}

type dotEdgeAttribute struct {
	value string
	color string
}

var dotHighlightColor = "magenta"

var dotEdgeAttrs = []dotEdgeAttribute{
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

var dotNodeAttrs = []dotNodeAttribute{
	{nstype: "Unknown", fillcolor: "magenta", shape: "rectangle", style: "rounded"},
	{nstype: "AuthAction", fillcolor: "lightcoral", shape: "invhouse", style: "rounded,filled"},
	{nstype: "AuthPolicy", fillcolor: "lightcoral", shape: "house", style: "rounded,filled"},
	{nstype: "AuthVServer", fillcolor: "lightcoral", shape: "house", style: "rounded,filled"},
	{nstype: "Cert", fillcolor: "greenyellow", shape: "cds", style: "rounded,filled"},
	{nstype: "CSAction", fillcolor: "lightsalmon", shape: "invhouse", style: "rounded,filled"},
	{nstype: "CSPolicy", fillcolor: "lightsalmon", shape: "house", style: "rounded,filled"},
	{nstype: "CSVServer", fillcolor: "lightsalmon", shape: "house", style: "rounded,filled"},
	{nstype: "DomainName", fillcolor: "aqua", shape: "house", style: "rounded,filled"},
	{nstype: "GSLBService", fillcolor: "lightblue", shape: "rectangle", style: "rounded,filled"},
	{nstype: "GSLBGroup", fillcolor: "lightblue", shape: "rectangle", style: "rounded,filled"},
	{nstype: "GSLBVServer", fillcolor: "lightblue", shape: "house", style: "rounded,filled"},
	{nstype: "LBGroup", fillcolor: "lightgoldenrodyellow", shape: "rectangle", style: "rounded,filled"},
	{nstype: "LBVServer", fillcolor: "lightgoldenrodyellow", shape: "house", style: "rounded,filled"},
	{nstype: "Netscaler", fillcolor: "aqua", shape: "doublecircle", style: "filled"},
	{nstype: "Policy", fillcolor: "lightpink", shape: "folder", style: "rounded,filled"},
	{nstype: "PolicyLabel", fillcolor: "lightpink", shape: "tab", style: "rounded,filled"},
	{nstype: "PortalTheme", fillcolor: "turquoise", shape: "note", style: "rounded,filled"},
	{nstype: "ResponderAction", fillcolor: "lightyellow", shape: "invhouse", style: "rounded,filled"},
	{nstype: "ResponderPolicy", fillcolor: "lightyellow", shape: "house", style: "rounded,filled"},
	{nstype: "RewriteAction", fillcolor: "lightgreen", shape: "invhouse", style: "rounded,filled"},
	{nstype: "RewritePolicy", fillcolor: "lightgreen", shape: "house", style: "rounded,filled"},
	{nstype: "Server", fillcolor: "white", shape: "rectangle", style: "rounded"},
	{nstype: "Service", fillcolor: "white", shape: "rectangle", style: "rounded"},
	{nstype: "ServiceGroup", fillcolor: "lightgrey", shape: "rectangle", style: "rounded,filled"},
	{nstype: "SessionAction", fillcolor: "palegreen", shape: "invhouse", style: "rounded,filled"},
	{nstype: "SessionPolicy", fillcolor: "palegreen", shape: "house", style: "rounded,filled"},
	{nstype: "STA", fillcolor: "lightblue", shape: "rectangle", style: "rounded,filled"},
	{nstype: "VPNVServer", fillcolor: "turquoise", shape: "house", style: "rounded,filled"},
	{nstype: "WI", fillcolor: "turquoise", shape: "polygon", style: "rounded,filled"},
	{nstype: "VIP", fillcolor: "yellow", shape: "doublecircle", style: "filled"},
}

func (ns *NSGraph) ExportDot(outputFile string, stdout bool) {

	ns.Graph = dot.NewGraph(dot.Directed)
	ns.Graph.Attr("rankdir", ns.Rankdir)

	for _, v := range ns.Nodes {
		attr := getDotNodeAttribute(v.nstype)
		if v.highlighted {

			ns.Graph.Node(v.label).Attr("fillcolor", attr.fillcolor).Attr("color", dotHighlightColor).Attr("shape", attr.shape).Attr("style", attr.style+",bold")
		} else {
			ns.Graph.Node(v.label).Attr("fillcolor", attr.fillcolor).Attr("shape", attr.shape).Attr("style", attr.style)
		}
	}

	for _, v := range ns.Edges {
		attr := getDotEdgeAttribute(v.port, v.protocol)
		from, found_from := ns.Graph.FindNodeById(v.from)
		if !found_from {
			continue
		}
		to, found_to := ns.Graph.FindNodeById(v.to)
		if !found_to {
			continue
		}
		ns.Graph.Edge(from, to, v.label).Attr("color", attr.color)
	}

	if stdout {
		slog.Info("generating DOT export to STDOUT")
		fmt.Print(ns.Graph.String())
	} else {
		slog.Info("generating DOT export", "output-file", outputFile)
		f, err := os.Create(outputFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		_, err = w.WriteString(ns.Graph.String())
		if err != nil {
			panic(err)
		}
		w.Flush()
	}

}

func getDotNodeAttribute(nstype string) dotNodeAttribute {
	for _, v := range dotNodeAttrs {
		if v.nstype == nstype {
			return v
		}
	}
	return dotNodeAttribute{nstype: "Unknown", fillcolor: "magenta", shape: "rectangle", style: "rounded,filled"}
}

func getDotEdgeAttribute(port, protocol string) dotEdgeAttribute {
	for _, v := range dotEdgeAttrs {
		if v.value == port || v.value == protocol {
			return v
		}
	}
	return dotEdgeAttribute{value: "*", color: "black"}
}
