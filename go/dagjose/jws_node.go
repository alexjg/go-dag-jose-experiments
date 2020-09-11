package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
    ipldBasicNode "github.com/ipld/go-ipld-prime/node/basic"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func (d *DagJOSE) ReprKind() ipld.ReprKind {
    return ipld.ReprKind_Map
}
func (d *DagJOSE) LookupByString(key string) (ipld.Node, error) {
    if key == "payload" {
        return ipldBasicNode.NewLink(cidlink.Link{Cid: *d.payload}), nil
    }
    return nil, nil
}
func (d *DagJOSE) LookupByNode(key ipld.Node) (ipld.Node, error) {
	ks, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(ks)
}
func (d *DagJOSE) LookupByIndex(idx int) (ipld.Node, error) {
    return nil, nil
}
func (d *DagJOSE) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	return d.LookupByString(seg.String())
}
func (d *DagJOSE) MapIterator() ipld.MapIterator {
    return nil
}
func (d *DagJOSE) ListIterator() ipld.ListIterator{
    return nil
}
func (d *DagJOSE) Length() int{
    return 0
}
func (d *DagJOSE) IsAbsent() bool{
    return false
}
func (d *DagJOSE) IsNull() bool{
    return false
}
func (d *DagJOSE) AsBool() (bool, error) {
    return false, nil
}
func (d *DagJOSE) AsInt() (int, error) {
    return 0, nil
}
func (d *DagJOSE) AsFloat() (float64, error) {
    return 0, nil
}
func (d *DagJOSE) AsString() (string, error) {
    return "", nil
}
func (d *DagJOSE) AsBytes() ([]byte, error) {
    return nil, nil
}
func (d *DagJOSE) AsLink() (ipld.Link, error) {
    return nil, nil
}
func (d *DagJOSE) Prototype() ipld.NodePrototype{
    return nil
}

// end ipld.Node implementation

