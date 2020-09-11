# dagjose-go

There are two parts to this codebase:

### `main.go`

This is a simple interop test where I have been experimenting with how to wire up the go-ipld-prime codebase to talk to IPFS and to load JWS's.

### `./dagjose`

This is the code which will shortly be pulled out into it's own repo as the dag-jose implementation for Go

## How does this work? 

We are interested in reading data from IPFS into JOSE objects which then expose an interface to easily interact with using ecosystem JOSE libraries such as go-jose. There are two points where we need to integrate with the `go-ipld-prime` library:

1. We need to register a new decoder which can handle the `0x85` multicodec code
2. When we pull nodes from IPFS (using `ipld.Link.Load`) we need to specify the `ipld.Node` implementation we want to build

The former is reasonably straightforward, there is a module `init` in `./dagjose/multicodec.go` which registers a decoder function with the go-ipld-prime library to be used when encountering the `0x85` code. Importing `dagjose` will automatically register a decoder as a side effect. This appears to be the way existing formats like `dag-cbor` and `dag-json` work in ipld prime. In fact, because we are not actually introducing a new wire format we pretty much just forward everything to the CBOR format handler.

The main feature this library offers then is an implementation of `NodeAssembler` which can be passed to `Link.load` when you want to build a `JOSE` object. Right now this is actually implemented with a hack which we cannot use for production (see the comments in multicodec.go for more information), this is because the `NodeAssembler` interface is quite large and I wanted to verify that the approach worked before implementing it. As a user then, if you want to load a JOSE object you do something like this:

```go
builder := dagjose.NewBuilder()
link := <somehow get a link to a JOSE object>
err := link.Load(
    <instance of context.Context>,
    <instance of ipld.LinkContext>,
    builder,                   
    <instance of ipld.Loader>
)
node = builder.Build()
```

Node is at this point an instance of `ipld.Node`, but because it has been built by `dagjose.DagJOSENodeBuilder` we can cast it to a `dagjose.DagJOSE` struct.

## Questions I have 

### Users have to choose a builder?

In order to get a DagJOSE object out of this approach users have to know that they are trying to load a DagJOSE object and pass in the correct builder. Otherwise the node they will get back will be of some other type and not expose the any kind of integration with existing JOSE tools. For example, if a user were to pass a `basicnode.NewBuilder()` in then the node implementation will be (for the example of a `dag-jose` object) a `node.basic.plainMap` (I think). This seems a little redundant as we already know that the data contains a JOSE object due to the `0x85` code, and the user has already indicated that they want to decode those into `DagJOSE` objects by importing the `dagjose` package.

This is probably due to the fact the the go-ipld-prime codebase is designed for generic deserialization with minimal allocations, so you need to specify what you are trying to build at load time. It's conceivable that someone might want to decode a JOSE object into some kind of alternative representation, in which case they would provide a different implementation of `ipld.NodeBuilder` but the import of `dagjose` would still be required to register the `dagcbor` decoder.

Still, I wonder if I am missing something.
