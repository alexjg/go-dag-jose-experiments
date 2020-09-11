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
    if (mod != null) for (var k in mod) if (k !== "default" && Object.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
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
basics_1.default.multicodec.add(dagJose.default);
const format = legacy_1.default(basics_1.default, dagJose.default.name);
const hash = "bafyrkmfzcmeff5yo5m2pfhisessei3mjrtbbe6oe6femzfqh4le4wwvpbwlk3ad5mv6tk4q2lmoercivxwxq";
const key = "0248aacea967f3972ddbd2fd29798c0f6607a65aa9bc7f3149e9294d31aedf80";
const signer = did_jwt_1.default.EllipticSigner(key);
async function main() {
    const api = await ipfs_1.default.create({ ipld: { formats: [format] } });
    const payload = new cids_1.default(hash);
    const jws = await dagJose.createDagJWS(payload, signer, {});
    console.log(jws);
    const cid = await api.dag.put(jws, { format: format.codec, hashAlg: "sha2-256" });
    console.log(cid);
}
main().then(() => {
    console.log("done");
}).catch(e => console.log(e));
