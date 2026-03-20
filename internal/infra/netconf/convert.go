package netconf

import "encoding/xml"

type GetReply struct {
	XMLName xml.Name `xml:"get"`
	Data    Data     `xml:"data"`
}

type Data struct {
	YangLibrary  YangLibrary  `xml:"yang-library"`
	ModulesState ModulesState `xml:"modules-state"`
}

type YangLibrary struct {
	XMLName   xml.Name    `xml:"yang-library"`
	ModuleSet []ModuleSet `xml:"module-set"`
	Schema    []SchemaRef `xml:"schema"`
	ContentID string      `xml:"content-id"`
}

type ModuleSet struct {
	Name             string             `xml:"name"`
	Modules          []ModuleEntry      `xml:"module"`
	ImportOnlyModule []ImportOnlyModule `xml:"import-only-module"`
}

type SchemaRef struct {
	Name      string `xml:"name"`
	ModuleSet string `xml:"module-set"`
}

type ModuleEntry struct {
	Name      string   `xml:"name"`
	Revision  string   `xml:"revision"`
	Namespace string   `xml:"namespace"`
	Location  string   `xml:"location,omitempty"`
	Features  []string `xml:"feature"`
}

type ImportOnlyModule struct {
	Name      string `xml:"name"`
	Revision  string `xml:"revision"`
	Namespace string `xml:"namespace"`
	Location  string `xml:"location,omitempty"`
}

type ModulesState struct {
	XMLName     xml.Name           `xml:"modules-state"`
	ModuleSetID string             `xml:"module-set-id"`
	Modules     []ModulesStateItem `xml:"module"`
}

type ModulesStateItem struct {
	Name            string   `xml:"name"`
	Revision        string   `xml:"revision"`
	Namespace       string   `xml:"namespace"`
	Schema          string   `xml:"schema,omitempty"`
	Features        []string `xml:"feature"`
	ConformanceType string   `xml:"conformance-type"`
}

func StringToXml(raw []byte) (GetReply, error) {
	var get GetReply
	err := xml.Unmarshal(raw, &get)
	if err != nil {
		return GetReply{}, err
	}
	return get, nil
}
