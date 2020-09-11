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
        panic("argh")
    }
    return string(encoded)
}

func Decoder(na ipld.NodeAssembler, r io.Reader) error {
    joseAssembler, isJoseAssembler := na.(*DagJOSENodeBuilder)
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
