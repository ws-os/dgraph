package query

import (
	"fmt"

	"github.com/dgraph-io/dgraph/protos"
	"github.com/dgraph-io/dgraph/x"
)

func DebugSubgraph(sg *SubGraph, indent string) {
	fmt.Printf("%sAttr=%q\n", indent, sg.Attr)
	fmt.Printf("%s  SrcUids=%v\n", indent, sg.SrcUIDs)
	for i := range sg.Children {
		fmt.Printf("%s  Child=%d\n", indent, i)
		DebugSubgraph(sg.Children[i], indent+"    ")
	}
}

func AssertSorted(sg *SubGraph) {
	if sg == nil {
		return
	}
	AssertUidListSorted(sg.SrcUIDs, "src uid")
	AssertUidListSorted(sg.DestUIDs, "dest uid")
	//for i := range sg.uidMatrix {
	//AssertUidListSorted(sg.uidMatrix[i], "uid matrix")
	//}
	for _, ch := range sg.Children {
		AssertSorted(ch)
	}
}

func AssertUidListSorted(pl *protos.List, msg string) {
	for i := 0; i+1 < len(pl.GetUids()); i++ {
		x.AssertTruef(pl.Uids[i] < pl.Uids[i+1],
			"%s list not sorted: %v", msg, pl.GetUids())
	}
}
