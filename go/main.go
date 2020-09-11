package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"dagjoseroundtrip/dagjose"

	"github.com/ipfs/go-cid"
	ipfsApi "github.com/ipfs/go-ipfs-api"
	ipld "github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	gojose "gopkg.in/square/go-jose.v2"
)

func main() {
	shell := ipfsApi.NewShell("localhost:5001")
    lb := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 48,   // sha3-224 hash has a 48-byte sum.
	}}
    bridge := IPFSBridge{
        shell,
        lb,
    }

    planetsLnk, err := createPlanets(&bridge)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating planets: %v", err)
        os.Exit(-1)
    }

    n, err := bridge.Load(planetsLnk)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading planets link: %v", err)
        os.Exit(-1)
    }

	fmt.Printf("we loaded a %s with %d entries\n", n.ReprKind(), n.Length())

    result, err := shell.BlockGet("bagcqcera6m5ryn2zhfcv6vakblfdicqdnq4fgz4mfquglrigjd6ltqbscebq")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting JWS block: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("We loaded the raw jws: %v\n", result)

    jwsCid, err := cid.Decode("bagcqcera6m5ryn2zhfcv6vakblfdicqdnq4fgz4mfquglrigjd6ltqbscebq")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating JWS link: %v", err)
        os.Exit(-1)
    }
    jwsLnk := cidlink.Link{Cid: jwsCid}

    jwsNode, err := bridge.LoadJWS(jwsLnk)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading jws link: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("Loaded node: %v\n", jwsNode)

    _, err = gojose.ParseSigned(jwsNode.FullSerialization())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing JOSE serialization: %v", err)
        os.Exit(-1)
    }
}

func createPlanets(bridge *IPFSBridge) (ipld.Link, error) {

    earthNode := fluent.MustBuildMap(basicnode.Prototype.Map, 1, func(na fluent.MapAssembler) {
        na.AssembleEntry("name").AssignString("Earth")
        na.AssembleEntry("radius").AssignFloat(6371000)
    })

    earthLink, err := bridge.Build(earthNode) 
    if err != nil {
        return nil, fmt.Errorf("error creating earth: %v", err)
    }

    marsNode := fluent.MustBuildMap(basicnode.Prototype.Map, 1, func(na fluent.MapAssembler) {
        na.AssembleEntry("name").AssignString("Mars")
        na.AssembleEntry("radius").AssignFloat(3389500)
    })

    marsLink, err := bridge.Build(marsNode)
    if err != nil {
        return nil, fmt.Errorf("error creating mars: %v", err)
    }

	np := fluent.MustBuildMap(basicnode.Prototype.Map, 1, func(na fluent.MapAssembler) {
		na.AssembleEntry("planets").AssignNode(fluent.MustBuildMap(basicnode.Prototype.Map, 2, func(na fluent.MapAssembler) {
            na.AssembleEntry("earth").AssignLink(earthLink)
            na.AssembleEntry("mars").AssignLink(marsLink)
        }))
	})

    return bridge.Build(np)
}

type IPFSBridge struct {
    *ipfsApi.Shell
    builder cidlink.LinkBuilder
}

func (i *IPFSBridge) Build(node ipld.Node) (ipld.Link, error) {
	return i.builder.Build(
		context.Background(),
		ipld.LinkContext{},
		node,
		func(ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
			buf := bytes.Buffer{}
			return &buf, func(lnk ipld.Link) error {
                _, err := i.BlockPut(buf.Bytes(), "cbor", "sha3-384", i.builder.MhLength)
                return err
			}, nil
		},
	)
}

func (i *IPFSBridge) Load(lnk ipld.Link) (ipld.Node, error) {
	nb := basicnode.Prototype.Any.NewBuilder()

    err := lnk.Load(
		context.Background(), 
		ipld.LinkContext{},   
		nb,                   
		i.loader,               
	)
	if err != nil {
        return nil, err
	}

    n := nb.Build()
    return n, nil
}

func (i *IPFSBridge) LoadJWS(lnk ipld.Link) (*dagjose.DagJOSE, error) {
    builder := dagjose.NewBuilder()
    err := lnk.Load(
		context.Background(), 
		ipld.LinkContext{},   
		builder,                   
		i.loader,               
	)
	if err != nil {
        return nil, err
	}

    n := builder.Build()
    return n.(*dagjose.DagJOSE), nil
}

func (i *IPFSBridge) loader(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
    theCid, ok := lnk.(cidlink.Link)
    if !ok {
        return nil, fmt.Errorf("Attempted to load a non CID link: %v", lnk)
    }
    block, err := i.BlockGet(theCid.String())
    if err != nil {
        return nil, fmt.Errorf("error loading %v: %v", theCid.String(), err)
    }
    return bytes.NewBuffer(block), nil
}
