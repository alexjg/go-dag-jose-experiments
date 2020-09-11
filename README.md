# Go/Javascript Dag JOSE Interop Example

This repo is three things:

- Some typescript which uses the js-dag-jose library to push some signed data to IPFS (`./ts`)
- A Go package called `dagjose` which implements various pieces of machinery to pull JOSE data out of IPFS (`./go/dagjose`)
- A Go program which uses the `dagjose` package to pull the signed data pushed _to_ IPFS by the Typescript in `./ts` out of IPFS and validate the signature on it

In general this code is mostly a set of experiments which I have used to understand how all the different pieces fit together.

My intention has been to get a proof of concept implementation of the dag-jose codec interoperating with the javascript implementation, this is why the `dagjose` package is not in it's own repo. We're currently able to read JWS's from IPFS using Go and I'm working on implementing writing JWS's _to_ IPFS. Once writing is complete it will be time to pull the `dagjose` Go package out into it's own repo.

To run this example you will need to do a few things:

- Get an IPFS daemon running on localhost:5001
- Run the typescript application:
    -  `cd ./ts`
    - `npm install`
    - `tsc`
    - `node index.js` - this will create a dag-jose JWS and advertise it on the network.
- Build and run the go appliction
    - `cd ./go`
    - `go build`
    - `./dagjoseroundtrip` - This will attempt to find the JWS advertised by the typescript application and verify it

    
Refer to `./go/README.md` and the comments in the code for more information

