package port

type GRPCClientReceiver interface {
	RecvType(i interface{})
	RecvPayloadMatchFn(i ...interface{})
	RecvErrMatchFn(i ...interface{})
}

type GRPCServerReceiver interface {
	RecvType(i interface{})
	RecvPayloadMatchFn(i ...interface{})
}

type HTTPReceiver interface {
	RecvType(i interface{})
	RecvPayloadMatchFn(i ...interface{})
}
