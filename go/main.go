package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"dagjoseroundtrip/dagjose"
    "encoding/hex"

	"golang.org/x/crypto/ed25519"

	"github.com/ipfs/go-cid"
	ipfsApi "github.com/ipfs/go-ipfs-api"
	ipld "github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	gojose "gopkg.in/square/go-jose.v2"
)

func main() {
	shell := ipfsApi.NewShell("localhost:5001")
    bridge := IPFSBridge{shell}

    key, err := hex.DecodeString("0248aacea967f3972ddbd2fd29798c0f6607a65aa9bc7f3149e9294d31aedf80")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error decoding key: %v", err)
        os.Exit(-1)
    }
    privateKey := ed25519.NewKeyFromSeed(key)
    fmt.Printf("Public: %v", privateKey.Public())

    jwsCid, err := cid.Decode("bagcqcerafzecmo67npzj56gfyukh3tbwxywgfsmxirr4oq57s6sddou4p5dq")
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
    fmt.Printf("Loaded jws: %v\n", jwsNode.GeneralJSONSerialization())

    sig, err := gojose.ParseSigned(jwsNode.GeneralJSONSerialization())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing JOSE serialization: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("Parsed JWS: %v\n", sig)
    c, e := sig.CompactSerialize()
    if e != nil { panic("compact!") }
    fmt.Printf("Compact serialization: %v\n", c)
    verifiedPayload, err := sig.Verify(privateKey.Public())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error verifying jws: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("Verified payload: %v\n", verifiedPayload)
    _, payloadCid, err := cid.CidFromBytes(verifiedPayload)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error decoding verified payload to CID: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("Decoded payload: %v\n", payloadCid)
}

type IPFSBridge struct {
    *ipfsApi.Shell
}

func (i *IPFSBridge) Build(node ipld.Node) (ipld.Link, error) {
    lb := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 48,   // sha3-224 hash has a 48-byte sum.
	}}
	return lb.Build(
		context.Background(),
		ipld.LinkContext{},
		node,
		func(ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
			buf := bytes.Buffer{}
			return &buf, func(lnk ipld.Link) error {
                _, err := i.BlockPut(buf.Bytes(), "cbor", "sha3-384", lb.MhLength)
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
