"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const ipfs_1 = __importDefault(require("ipfs"));
// @ts-ignore
const basics_1 = __importDefault(require("multiformats/basics"));
// @ts-ignore
const legacy_1 = __importDefault(require("multiformats/legacy"));
const dagJose = __importStar(require("dag-jose"));
const did_jwt_1 = __importDefault(require("did-jwt"));
const cids_1 = __importDefault(require("cids"));
const ed25519 = __importStar(require("@stablelib/ed25519"));
basics_1.default.multicodec.add(dagJose.default);
const format = legacy_1.default(basics_1.default, dagJose.default.name);
const hash = "bafyrkmfzcmeff5yo5m2pfhisessei3mjrtbbe6oe6femzfqh4le4wwvpbwlk3ad5mv6tk4q2lmoercivxwxq";
const seed = base64ToBytes("Akiqzqln85ct29L9KXmMD2YHplqpvH8xSekpTTGu34A=");
async function main() {
    const keypair = ed25519.generateKeyPairFromSeed(seed);
    const signer = did_jwt_1.default.NaclSigner(bytesToBase64(keypair.secretKey));
    const api = await ipfs_1.default.create({ ipld: { formats: [format] } });
    const payload = new cids_1.default(hash);
    const jws = await dagJose.createDagJWS(payload, signer, { "alg": "EdDSA" });
    //dagJose.verifyDagJWS(jws, [{
    //id: "",
    //type: "",
    //controller: "",
    //publicKeyBase64: bytesToBase64(keypair.publicKey)
    //}])
    console.log(jws);
    const cid = await api.dag.put(jws, { format: format.codec, hashAlg: "sha2-256" });
    console.log(cid);
}
main().then(() => {
    console.log("done");
}).catch(e => console.log(e));
function base64ToBytes(s) {
    return new Uint8Array(Array.prototype.slice.call(Buffer.from(s, 'base64'), 0));
}
function bytesToBase64(b) {
    return Buffer.from(b).toString('base64');
}
