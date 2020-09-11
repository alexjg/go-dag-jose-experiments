import Ipfs from "ipfs"
// @ts-ignore
import multiformats from "multiformats/basics"
// @ts-ignore
import legacy from "multiformats/legacy"
import * as dagJose from "dag-jose"
import  didJWT from "did-jwt"
import CID from "cids"
import * as ed25519 from '@stablelib/ed25519'

multiformats.multicodec.add(dagJose.default)
const format = legacy(multiformats, dagJose.default.name)

const hash = "bafyrkmfzcmeff5yo5m2pfhisessei3mjrtbbe6oe6femzfqh4le4wwvpbwlk3ad5mv6tk4q2lmoercivxwxq"
const seed = base64ToBytes("Akiqzqln85ct29L9KXmMD2YHplqpvH8xSekpTTGu34A=")

async function main() {
    const keypair = ed25519.generateKeyPairFromSeed(seed)
    const signer = didJWT.NaclSigner(bytesToBase64(keypair.secretKey))
    const api = await Ipfs.create({ipld: {formats: [format]}})
    const payload = new CID(hash)
    const jws = await dagJose.createDagJWS(payload, signer, {"alg": "EdDSA"})
    dagJose.verifyDagJWS(jws, [{
        id: "",
        type: "",
        controller: "",
        publicKeyBase64: bytesToBase64(keypair.publicKey)
    }])
    console.log(jws)
    const cid = await api.dag.put(jws, { format: format.codec, hashAlg: "sha2-256"})
    console.log(cid)
}

main().then(() => {
    console.log("done")
}).catch(
    e => console.log(e)
)

function base64ToBytes(s: string): Uint8Array {
  return new Uint8Array(Array.prototype.slice.call(Buffer.from(s, 'base64'), 0))
}

function bytesToBase64(b: Uint8Array): string {
  return Buffer.from(b).toString('base64')
}
