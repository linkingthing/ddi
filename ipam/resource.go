package ipam

import (
//"github.com/zdnscloud/gorest/resource"
)

type Subtree struct {
	//resource.ResourceBase `json:",inline"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BeginSubnet    string    `json:"beginsubnet"`
	EndSubnet      string    `json:"endsubnet"`
	BeginNodeCode  byte      `json:"beginnodecode"`
	EndNodeCode    byte      `json:"endnodecode"`
	SubtreeBitNum  byte      `json:"subtreebitnum"`
	Depth          int       `json:"depth"`
	SubtreeUseDFor string    `json:"usedfor"`
	Nodes          []Subtree `json:"nodes"`
}

type SplitSubnet struct {
	ID         string `json:"todeleteid"`
	NameFirst  string `json:"namefirst"`
	NameSecond string `json:"namesecond"`
	Bitwith    byte   `json:"bitwith"`
}

type SplitSubnetResult struct {
	First  Subtree `json:"firstnode"`
	Second Subtree `json:"secondnode"`
}
