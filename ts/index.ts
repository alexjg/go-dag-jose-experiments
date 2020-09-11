import Ipfs from "ipfs"
// @ts-ignore
import multiformats from "multiformats/basics"
// @ts-ignore
import legacy from "multiformats/legacy"
import * as dagJose from "dag-jose"
import  didJWT from "did-jwt"
import CID from "cids"

multiformats.multicodec.add(dagJose.default)
const format = legacy(multiformats, dagJose.default.name)

const hash = "bafyrkmfzcmeff5yo5m2pfhisessei3mjrtbbe6oe6femzfqh4le4wwvpbwlk3ad5mv6tk4q2lmoercivxwxq"
const key = "0248aacea967f3972ddbd2fd29798c0f6607a65aa9bc7f3149e9294d31aedf80"
const signer = didJWT.EllipticSigner(key)

async function main() {
    const api = await Ipfs.create({ipld: {formats: [format]}})
    const payload = new CID(hash)
    const jws = await dagJose.createDagJWS(payload, signer, {})
    console.log(jws)
    const cid = await api.dag.put(jws, { format: format.codec, hashAlg: "sha2-256"})
    console.log(cid)
}

main().then(() => {
    console.log("done")
}).catch(
    e => console.log(e)
)
