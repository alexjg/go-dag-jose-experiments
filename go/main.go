package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"encoding/hex"

	"golang.org/x/crypto/ed25519"

	"github.com/ipfs/go-cid"
	ipfsApi "github.com/ipfs/go-ipfs-api"
	ipld "github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basic"
	gojose "gopkg.in/square/go-jose.v2"
    "github.com/alexjg/go-dag-jose/dagjose"
)

func main() {
	shell := ipfsApi.NewShell("localhost:5001")
	bridge := IPFSBridge{shell}

	// This is the private key we used to sign the JWS in typescript
	key, err := hex.DecodeString("0248aacea967f3972ddbd2fd29798c0f6607a65aa9bc7f3149e9294d31aedf80")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding key: %v", err)
		os.Exit(-1)
	}
	privateKey := ed25519.NewKeyFromSeed(key)

	// This is the CID of the JWS the typescript application is advertising
	jwsCid, err := cid.Decode("bagcqcerafzecmo67npzj56gfyukh3tbwxywgfsmxirr4oq57s6sddou4p5dq")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JWS link: %v", err)
		os.Exit(-1)
	}
	jwsLnk := cidlink.Link{Cid: jwsCid}

	// This is where we actually attempt to load the JWS
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

	helloNode := fluent.MustBuildMap(
		basicnode.Prototype.Map,
		1,
		func(ma fluent.MapAssembler) {
			ma.AssembleEntry("hello").AssignString("mars")
		},
	)
	helloLink, err := bridge.BuildJWS(helloNode)
	helloCid := helloLink.(cidlink.Link).Cid
	fmt.Printf("Hellocid: %v\n", helloCid)
	signer, err := gojose.NewSigner(gojose.SigningKey{
		Algorithm: gojose.EdDSA,
		Key:       privateKey,
	}, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating signer: %v", err)
		os.Exit(-1)
	}
	gojoseJws, err := signer.Sign(helloCid.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing payload: %v", err)
		os.Exit(-1)
	}
	dagJws, err := dagjose.NewDagJWS(gojoseJws.FullSerialize())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dagjose.NewDagJWS: %v", err)
		os.Exit(-1)
	}
    newJwsLink, err := bridge.BuildJWS(dagJws)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error creating JWS node: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("New link is: %v\n", newJwsLink)

    retrieved, err := bridge.LoadJWS(newJwsLink)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error retrieving newly created JWS: %v", err)
        os.Exit(-1)
    }
    fmt.Printf("Retrieved jws: %v\n", retrieved.GeneralJSONSerialization())

	sig2, err := gojose.ParseSigned(jwsNode.GeneralJSONSerialization())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JOSE serialization of second JWS: %v", err)
		os.Exit(-1)
	}
	fmt.Printf("Parsed JWS: %v\n", sig2)
    _, err = sig2.Verify(privateKey.Public())
    if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying second JWS: %v", err)
		os.Exit(-1)
    }
    println("Verified second payload")
}

// The IPFSBridge is an adapter which teaches the ipfs.Shell HTTP client how
// to load data for go-ipld-prime
type IPFSBridge struct {
	*ipfsApi.Shell
}

// Pushe a node to IPFS and return the corresponding ipld.Link
func (i *IPFSBridge) BuildJWS(node ipld.Node) (ipld.Link, error) {
	lb := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x85, // 0x71 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
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
				_, err := i.BlockPut(buf.Bytes(), "dag-jose", "sha3-384", lb.MhLength)
				return err
			}, nil
		},
	)
}

// Given a link attempt to load a dagjose.DagJOSE object from it
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
