package dagjose

import (
	"fmt"
	"io"
    "encoding/base64"
    "encoding/json"

	cbor "github.com/fxamacker/cbor/v2"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

)


func init() {
	cidlink.RegisterMulticodecDecoder(0x85, Decoder)
	cidlink.RegisterMulticodecEncoder(0x85, dagcbor.Encoder)
}

type JOSESignature struct {
    protected *string
    header map[string]string
    signature []byte
}

type DagJOSE struct {
    payload *cid.Cid
    signatures []JOSESignature
}

func (d *DagJOSE) GeneralJSONSerialization() string {
    jsonJose := make(map[string]interface{})
    jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.payload.Bytes()) 
    sigs := make([]map[string]string, 0)
    for _, sig := range d.signatures {
        jsonSig := make(map[string]string) 
        if sig.protected != nil {
            jsonSig["protected"] = *sig.protected
        }
        jsonSig["signature"] = base64.RawURLEncoding.EncodeToString(sig.signature)
        sigs = append(sigs, jsonSig)
    }
    jsonJose["signatures"] = sigs
    encoded, err := json.Marshal(jsonJose)
    if err != nil {
        panic("impossible")
    }
    return string(encoded)
}

func Decoder(na ipld.NodeAssembler, r io.Reader) error {
    joseAssembler, isJoseAssembler := na.(*DagJOSENodeBuilder)
    // THIS IS A HACK
    // Rather than implementing the `NodeAssembler` interface, we are just 
    // checking if the user has indicated that they want to construct a 
    // DagJOSE, which they do by passing a DagJOSENodeBuilder to ipld.Link.Load.
    // We then proceed to decode the data to CBOR, and then construct a DagJOSE
    // object from the deserialized CBOR. Allocating an intermediary object
    // is explicitly what the whole `NodeAssembler` machinery is designed to avoid
    // so we absolutely should not do this.
    //
    // The next step here is to implement `NodeAssembler` (in `assembler.go`)
    // in such a way that it throws errors if the incoming data does not match
    // the expected layout of a dag-jose object. The only reason I have not
    // done this yet is that it requires a lot of code to implement NodeAssembler
    // and I wanted to check that the user facing API made sense first.
    //
    // This is also why this code contains very little error checking, we'll be
    // doing that more thoroughly in the NodeAssembler implementation
    if isJoseAssembler {
        rawDecoded := make(map[string]interface{})
        decoder := cbor.NewDecoder(r)
        err := decoder.Decode(&rawDecoded)
        if err != nil {
            return fmt.Errorf("error decoding CBOR for dag-jose: %v", err)
        }
        payload := rawDecoded["payload"].([]byte)
        cidPayload, err := cid.Cast(payload)
        if err != nil {
            return fmt.Errorf("Error casting payload to cid: %v", err)
        }
        joseAssembler.dagJose.payload = &cidPayload
        for _, rawSig := range rawDecoded["signatures"].([]interface{}) {
            sig := rawSig.(map[interface{}]interface{})
            protected := base64.RawURLEncoding.EncodeToString(sig["protected"].([]byte))
            signature := sig["signature"].([]byte)
            joseAssembler.dagJose.signatures = append(
                joseAssembler.dagJose.signatures,
                JOSESignature{
                    protected: &protected,
                    signature: signature,
                    header: nil,
                },
            )
        }
        return nil
    }
    err := dagcbor.Decoder(na, r)
    if err != nil {
        return err
    }
    return nil
}
