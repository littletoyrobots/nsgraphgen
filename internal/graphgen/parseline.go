package graphgen

import (
	"regexp"
	"slices"
	"strings"
)

func (ns *NSGraph) parseNSline(line string) {

	// name := ""
	// ip := ""
	// port := ""
	// protocol := ""
	// to := ""
	// from := ""
	// re := regexp.MustCompile(` (?=(?:[^"]|"[^"]*")*$)`)
	// var term []string
	re := regexp.MustCompile(`"((?:\\.|[^"\\])*)"|(\S+)`)
	term := re.FindAllString(line, -1)
	for i, v := range term {
		term[i] = strings.Trim(v, "\"")
	}

	if strings.HasPrefix(line, "add authentication ldapAction ") {
		name := term[3]
		to := term[5]
		port := term[7]
		protocol := ""
		switch port {
		case "389":
			protocol = "LDAP"
		case "636":
			protocol = "LDAPS"
			//	default: protocol = ""
		}
		ns.addNode("AuthAction", name, "", "", "")
		ns.addEdge(name, to, port, protocol)
	} else if strings.HasPrefix(line, "add authentication ldapPolicy") {
		name := term[3]
		to := term[5]
		ns.addNode("AuthPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication OAuthAction ") {
		name := term[3]
		to := term[5]
		ns.addNode("AuthAction", name, "", "", "")
		ns.addEdge(name, to, "", "OAUTH")
	} else if strings.HasPrefix(line, "add authentication OAuthIdPPolicy ") {
		name := term[3]
		to := term[7]
		ns.addNode("AuthPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication Policy ") {
		name := term[3]
		to := term[7]
		ns.addNode("AuthPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication policyLabel") {
		name := term[3]
		to := term[7]
		ns.addNode("PolicyLabel", name, "", "", "LoginSchema")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication radiusAction ") {
		name := term[3]
		to := term[5]
		port := term[7]
		ns.addNode("AuthAction", name, "", "", "")
		ns.addEdge(name, to, port, "RADIUS")
	} else if strings.HasPrefix(line, "add authentication radiusPolicy ") {
		name := term[3]
		to := term[5]
		ns.addNode("AuthPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication samlAction ") {
		// regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
		name := term[3]
		cert := term[5]
		to := term[9]
		ns.addNode("AuthAction", name, "", "", "")
		ns.addNode("Cert", cert, "", "", "CERT")
		ns.addEdge(name, cert, "", "CERT")
		ns.addEdge(name, to, "", "SAML")
	} else if strings.HasPrefix(line, "add authentication samlPolicy ") {
		name := term[3]
		to := term[5]
		ns.addNode("AuthPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add authentication vserver ") {
		name := term[3]
		protocol := term[4]
		vip := term[5]
		ns.addNode("AuthVServer", name, "", "", protocol)
		if vip == "0.0.0.0" {
			ns.addEdge("Global", name, "", protocol)
		} else {
			port := term[6]
			ns.addNode("VIP", "", vip, "", "")
			ns.addEdge(vip, name, port, protocol)
		}
	} else if strings.HasPrefix(line, "add cs action ") {
		name := term[3]
		ns.addNode("CSAction", name, "", "", "")
		if strings.Contains(line, "targetVserver") {
			to := term[5]
			ns.addEdge(name, to, "", "")
		}
		if strings.Contains(line, "targetLBVserver") {
			to := term[5]
			ns.addNode("LBVServer", to, "", "", "")
			ns.addEdge(name, to, "", "")
		}
	} else if strings.HasPrefix(line, "add cs policy ") {
		name := term[3]
		ns.addNode("CSPolicy", name, "", "", "")
		actionIdx := slices.Index(term[:], "-action")
		if actionIdx != -1 {
			to := term[actionIdx+1]
			ns.addNode("CSAction", to, "", "", "")
			ns.addEdge(name, to, "", "")
		}
		ruleIdx := slices.Index(term[:], "-rule")
		if ruleIdx != -1 {
			pipeRegex := regexp.MustCompile(`\|`)
			hostRegex := regexp.MustCompile(`HTTP\.REQ\.HOSTNAME\.EQ\(\\"([^"]+)\\"\)`)
			rules := pipeRegex.Split(term[ruleIdx+1], -1)
			for _, rule := range rules {
				if strings.Contains(rule, "HTTP.REQ.HOSTNAME.EQ") {
					match := hostRegex.FindStringSubmatch(rule)
					if len(match) > 1 {
						domain := match[1]
						ns.addNode("DomainName", domain, "", "", "")
						ns.addEdge(domain, name, "", "")
					}
				}
			}
		}
	} else if strings.HasPrefix(line, "add cs vserver ") {
		name := term[3]
		protocol := term[4]
		vip := term[5]
		ns.addNode("CSVServer", name, "", "", protocol)
		if vip == "0.0.0.0" {
			ns.addEdge("Global", name, "", protocol)
		} else {
			port := term[6]
			ns.addNode("VIP", "", vip, "", "")
			ns.addEdge(vip, name, port, protocol)
		}
	} else if strings.HasPrefix(line, "add gslb service ") {
		name := term[3]
		to := term[4] // ip
		protocol := term[5]
		port := term[6]
		publicip := term[8]
		ns.addNode("GSLBService", name, "", port, protocol)
		ns.addEdge(name, to, port, protocol)
		if to != publicip {
			public_port := term[10]
			ns.addEdge(publicip, name, public_port, "")
		}
	} else if strings.HasPrefix(line, "add gslb vserver ") {
		name := term[3]
		protocol := term[4]

		ns.addNode("GSLBVServer", name, "", "", protocol)
		idx := slices.Index(term[:], "-backupVServer")
		if idx != -1 {
			backup := term[idx+1]
			backup_protocol := term[slices.Index(term, "-backupLBMethod")+1]
			ns.addEdge(name, backup, "", backup_protocol)
			ns.addEdge(backup, name, "", backup_protocol)
		}

	} else if strings.HasPrefix(line, "add ha node ") {
		ip := term[4]
		ns.addNode("Netscaler", "", ip, "", "")
	} else if strings.HasPrefix(line, "add lb group ") {
		name := term[3]
		ns.addNode("LBGroup", name, "", "", "")
	} else if strings.HasPrefix(line, "add lb vserver ") {
		// fmt.Printf("line: %s\n", line)
		// fmt.Println(term)
		name := term[3]
		protocol := term[4]
		vip := term[5]
		port := term[6]
		ns.addNode("LBVServer", name, "", "", protocol)
		if vip == "0.0.0.0" {
			ns.addEdge(vip, name, "", protocol)
		} else {
			ns.addNode("VIP", "", vip, "", "")
			ns.addEdge(vip, name, port, protocol)
		}
		backupIdx := slices.Index(term[:], "-backupVServer")
		if backupIdx != -1 {
			backup := term[backupIdx+1]
			ns.addNode("LBVServer", backup, "", "", "")
			ns.addEdge(name, backup, "", "")
		}
	} else if strings.HasPrefix(line, "add ns ip ") {
		ip := term[3]
		ns.addNode("Netscaler", "", ip, "", "")
	} else if strings.HasPrefix(line, "add responder action ") {
		name := term[3]
		ns.addNode("ResponderAction", name, "", "", "")
	} else if strings.HasPrefix(line, "add responder policy ") {
		name := term[3]
		to := term[5]
		ns.addNode("ResponderPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add rewrite action ") {
		name := term[3]
		ns.addNode("RewriteAction", name, "", "", "")
		if term[4] != "replace" {
			to := term[5]
			ns.addEdge(name, to, "", "")
		}
	} else if strings.HasPrefix(line, "add rewrite policy ") {
		name := term[3]
		to := term[5]
		ns.addNode("RewritePolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add server ") {
		name := term[2]
		ip := term[3]
		ns.addNode("Server", name, ip, "", "")
	} else if strings.HasPrefix(line, "add service ") {
		name := term[2]
		to := term[3]
		port := term[5]
		protocol := term[4]
		ns.addNode("Service", name, "", "", "")
		ns.addEdge(name, to, port, protocol)
	} else if strings.HasPrefix(line, "add serviceGroup ") {
		name := term[2]
		protocol := term[3]
		ns.addNode("ServiceGroup", name, "", "", protocol)
	} else if strings.HasPrefix(line, "add ssl certkey ") {
		name := term[3]
		ns.addNode("Cert", name, "", "", "CERT")

	} else if strings.HasPrefix(line, "add vpn portaltheme ") {
		name := term[3]
		ns.addNode("PortalTheme", name, "", "", "")
		if strings.Contains(line, "basetheme") {
			basetheme := term[5]
			ns.addNode("PortalTheme", basetheme, "", "", "")
			ns.addEdge(name, basetheme, "", "BaseTheme")
		}
	} else if strings.HasPrefix(line, "add vpn sessionAction ") {
		name := term[3]
		ns.addNode("SessionAction", name, "", "", "")
		// if strings.Contains(line, "-wihome") {
		idx := slices.Index(term[:], "-wihome")
		if idx != -1 {
			wi := term[idx+1]
			ns.addNode("WI", wi, "", "", "")
			ns.addEdge(name, wi, "", "wi")
		}
	} else if strings.HasPrefix(line, "add vpn sessionPolicy ") {
		name := term[3]
		to := term[5]
		ns.addNode("SessionPolicy", name, "", "", "")
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "add vpn vserver ") {
		name := term[3]
		protocol := term[4]
		vip := term[5]
		port := term[6]
		ns.addNode("VPNVServer", name, "", "", protocol)
		if vip == "0.0.0.0" {
			ns.addEdge(vip, name, "", protocol)
		} else {
			ns.addNode("VIP", "", vip, "", "")
			ns.addEdge(vip, name, port, protocol)
		}
	} else if strings.HasPrefix(line, "bind authentication policylabel ") {
		name := term[3]
		to := term[5]
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "bind authentication vserver ") {
		name := term[3]
		idx := slices.Index(term[:], "-policy")
		if idx != -1 {
			policy := term[idx+1]
			if !strings.HasPrefix(policy, "_") {
				nfIdx := slices.Index(term[:], "-nextFactor")
				if nfIdx != -1 {
					ns.addEdge(name, policy, "", "nFactor")
					next := term[nfIdx+1]
					ns.addEdge(policy, next, "", "nFactor")
				} else {
					ns.addEdge(name, policy, "", "")
				}
			}
		}

		idx = slices.Index(term[:], "-policyLabel")
		if idx != -1 {
			label := term[idx+1]
			ns.addEdge(name, label, "", "")
		}

		idx = slices.Index(term[:], "-portaltheme")
		if idx != -1 {
			theme := term[idx+1]
			ns.addEdge(name, theme, "", "")
		}

	} else if strings.HasPrefix(line, "bind cs vserver") {
		name := term[3]
		policy := term[5]
		ns.addNode("CSVServer", name, "", "", "")
		polIdx := slices.Index(term[:], "-policyName")
		if polIdx != -1 {
			policy := term[polIdx+1]
			ns.addEdge(name, policy, "", "")
		}

		lbIdx := slices.Index(term[:], "-targetLBVserver")
		if lbIdx != -1 {
			lbvserver := term[lbIdx+1]
			ns.addNode("LBVServer", lbvserver, "", "", "") // TODO
			ns.addNode("CSPolicy", policy, "", "", "")     // TODO
			ns.addEdge(policy, lbvserver, "", "")
		}
	} else if strings.HasPrefix(line, "bind gslb service ") {
		idx := slices.Index(term[:], "-viewName")
		if idx != -1 {
			name := term[3]
			protocol := term[idx+1]
			target := term[idx+2]
			ns.addEdge(name, target, "", protocol)
		}
	} else if strings.HasPrefix(line, "bind gslb vserver ") {
		name := term[3]
		domainIdx := slices.Index(term[:], "-domainName")
		if domainIdx != -1 {
			domain := term[domainIdx+1]
			ns.addNode("DomainName", domain, "", "", "")
			ns.addEdge(domain, name, "", "gslb")
		}

		serviceIdx := slices.Index(term[:], "-serviceName")
		if serviceIdx != -1 {
			service := term[serviceIdx+1]
			ns.addEdge(name, service, "", "gslb")
		}
	} else if strings.HasPrefix(line, "bind ha ") {
		// TODO:
	} else if strings.HasPrefix(line, "bind lb group ") {
		name := term[3]
		target := term[4]
		ns.addEdge(name, target, "", "")
	} else if strings.HasPrefix(line, "bind lb vserver ") {
		name := term[3]
		idx := slices.Index(term[:], "-policyName")
		if idx != -1 {
			policy := term[idx+1]
			ns.addEdge(name, policy, "", "")
		} else {
			target := term[4] // server or serviceGroup
			ns.addEdge(name, target, "", "")
		}
	} else if strings.HasPrefix(line, "bind ns ") {
		// TODO:
	} else if strings.HasPrefix(line, "bind responder cs vserver ") {
		name := term[3]
		to := term[5]
		ns.addEdge(name, to, "", "")
	} else if strings.HasPrefix(line, "bind responder global ") {
		name := term[3]
		ns.addEdge("Global", name, "", "") // global to policy
	} else if strings.HasPrefix(line, "bind responder ssl vserver ") {
		if !strings.Contains(line, "eccCurveName") && !strings.Contains(line, "cipherName") {
			name := term[3]
			cert := term[5]
			ns.addNode("Cert", cert, "", "", "CERT")
			ns.addEdge(name, cert, "", "CERT")
		}
	} else if strings.HasPrefix(line, "bind responder vpn ") {
		// TODO:
	} else if strings.HasPrefix(line, "bind server ") {
		// TODO:
		// name := term[3]
		// ns.add_edge(name, cert, "", "CERT")

	} else if strings.HasPrefix(line, "bind service ") {
		name := term[2]
		idx := slices.Index(term[:], "-monitorName")
		if idx == -1 {
			target := term[3]
			ns.addEdge(name, target, "", "")
		}
	} else if strings.HasPrefix(line, "bind serviceGroup ") {
		name := term[2]
		idx := slices.Index(term[:], "-monitorName")
		if idx == -1 {
			target := term[3]
			port := term[4]
			ns.addEdge(name, target, port, "")
		}
	} else if strings.HasPrefix(line, "bind ssl vserver ") {
		name := term[3]
		cert := term[5]
		ns.addNode("Cert", cert, "", "", "CERT")
		ns.addEdge(name, cert, "", "CERT")
	} else if strings.HasPrefix(line, "bind vpn global ") {
		name := term[4]
		ns.addEdge("0.0.0.0", name, "", "")
	} else if strings.HasPrefix(line, "bind vpn vserver ") {
		name := term[3]
		staIdx := slices.Index(term[:], "-staServer")
		if staIdx != -1 {
			sta := term[staIdx+1]
			ns.addNode("STA", name, "", "", "STA")
			ns.addNode("STA", sta, "", "", "")
			ns.addEdge(name, sta, "", "STA")
		}
		polIdx := slices.Index(term[:], "-policy")
		if polIdx != -1 {
			policy := term[polIdx+1]
			if !strings.HasPrefix(policy, "_") {
				ns.addEdge(name, policy, "", "")
			}
		}
		porIdx := slices.Index(term[:], "-portaltheme")
		if porIdx != -1 {
			portaltheme := term[porIdx+1]
			ns.addEdge(name, portaltheme, "", "")
		}
	} else if strings.HasPrefix(line, "link ssl certkey ") {
		name := term[3]
		cert := term[4]
		ns.addNode("Cert", cert, "", "", "CERT")
		ns.addEdge(name, cert, "", "CERT")
	} else if strings.HasPrefix(line, "set ns config ") {
		if term[3] == "-IPAddress" {
			ip := term[4]
			ns.addNode("Netscaler", "", ip, "", "")
		}
	}

	// return nil
}
