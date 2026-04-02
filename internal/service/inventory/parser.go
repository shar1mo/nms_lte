package inventory

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"nms_lte/internal/model"
)

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Text    string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

var dnKeyCandidates = []string{
	"dn",
	"distinguishedName",
	"id",
	"name",
	"localId",
	"userLabel",
	"uuid",
}

func buildInventoryObjects(replies ...[]byte) ([]model.InventoryObject, error) {
	objectsByDN := make(map[string]model.InventoryObject)

	for _, raw := range replies {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			continue
		}

		nodes, err := parseXMLNodes(raw)
		if err != nil {
			return nil, fmt.Errorf("parse inventory xml: %w", err)
		}

		dataRoots := collectDataRoots(nodes)
		if len(dataRoots) == 0 {
			dataRoots = nodes
		}

		for _, root := range dataRoots {
			walkInventoryNodes(root.Nodes, nil, objectsByDN)
		}
	}

	out := make([]model.InventoryObject, 0, len(objectsByDN))
	for _, object := range objectsByDN {
		out = append(out, object)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DN < out[j].DN
	})

	return out, nil
}

func parseXMLNodes(raw []byte) ([]xmlNode, error) {
	wrapped := append([]byte("<root>"), raw...)
	wrapped = append(wrapped, []byte("</root>")...)

	var root xmlNode
	if err := xml.Unmarshal(wrapped, &root); err != nil {
		return nil, err
	}

	return root.Nodes, nil
}

func collectDataRoots(nodes []xmlNode) []xmlNode {
	out := make([]xmlNode, 0)

	var walk func(node xmlNode)
	walk = func(node xmlNode) {
		switch node.XMLName.Local {
		case "data":
			out = append(out, node)
			return
		case "rpc-reply", "rpc", "root":
		default:
			for _, child := range node.Nodes {
				walk(child)
			}
			return
		}

		for _, child := range node.Nodes {
			walk(child)
		}
	}

	for _, node := range nodes {
		walk(node)
	}

	return out
}

func walkInventoryNodes(nodes []xmlNode, parentDN []string, objects map[string]model.InventoryObject) {
	siblingTotals := make(map[string]int)
	for _, node := range nodes {
		if hasElementChildren(node) {
			siblingTotals[node.XMLName.Local]++
		}
	}

	siblingIndex := make(map[string]int)
	for _, node := range nodes {
		if !hasElementChildren(node) {
			continue
		}

		className := node.XMLName.Local
		siblingIndex[className]++
		attributes := directLeafAttributes(node)
		segment := buildDNSegment(className, attributes, siblingIndex[className], siblingTotals[className] > 1)
		dnParts := append(append([]string(nil), parentDN...), segment)
		dn := strings.Join(dnParts, ",")

		if len(attributes) > 0 || len(node.Attrs) > 0 {
			object := objects[dn]
			object.DN = dn
			object.Class = className
			if object.Attributes == nil {
				object.Attributes = make(map[string]string, len(attributes)+len(node.Attrs))
			}
			for _, attr := range node.Attrs {
				object.Attributes["@"+attr.Name.Local] = attr.Value
			}
			for key, value := range attributes {
				object.Attributes[key] = value
			}
			objects[dn] = object
		}

		complexChildren := complexChildNodes(node)
		if len(complexChildren) > 0 {
			walkInventoryNodes(complexChildren, dnParts, objects)
		}
	}
}

func complexChildNodes(node xmlNode) []xmlNode {
	children := make([]xmlNode, 0, len(node.Nodes))
	for _, child := range node.Nodes {
		if hasElementChildren(child) {
			children = append(children, child)
		}
	}
	return children
}

func directLeafAttributes(node xmlNode) map[string]string {
	attributes := make(map[string]string)

	for _, child := range node.Nodes {
		if hasElementChildren(child) {
			continue
		}

		key := strings.TrimSpace(child.XMLName.Local)
		if key == "" {
			continue
		}

		value := strings.TrimSpace(child.Text)
		if current, ok := attributes[key]; ok && current != "" && value != "" {
			attributes[key] = current + "," + value
			continue
		}
		attributes[key] = value
	}

	return attributes
}

func buildDNSegment(className string, attributes map[string]string, index int, repeated bool) string {
	for _, key := range dnKeyCandidates {
		value := strings.TrimSpace(attributes[key])
		if value != "" {
			return fmt.Sprintf("%s=%s", className, sanitizeDNValue(value))
		}
	}

	if repeated {
		return fmt.Sprintf("%s=%d", className, index)
	}

	return className
}

func sanitizeDNValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "_")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	if value == "" {
		return "unknown"
	}
	return value
}

func hasElementChildren(node xmlNode) bool {
	for _, child := range node.Nodes {
		if child.XMLName.Local != "" {
			return true
		}
	}
	return false
}

func validateInventoryObjects(objects []model.InventoryObject) error {
	if len(objects) == 0 {
		return errors.New("inventory reply did not contain any objects")
	}
	return nil
}
