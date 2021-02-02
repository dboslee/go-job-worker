package api

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// authContext checks for a clientID in the TLS Cert CommonName and passes it to the context
func authContext(ctx context.Context) (context.Context, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, PermissionDenied
	}
	mtls, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, PermissionDenied
	}

	// TODO: Clients should use x509 extensions since CommonName is depricated
	var clientID string
	for _, cert := range mtls.State.PeerCertificates {
		clientID = cert.Subject.CommonName
		break
	}
	if clientID == "" {
		return nil, PermissionDenied
	}
	log.Printf("%v connected", clientID)

	return context.WithValue(ctx, KeyClientID, clientID), nil
}

// AuthUnary checks for a clientID and passes it to the context
func AuthUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctx, err = authContext(ctx)
	if err != nil {
		log.Printf("method: %v err: %v", info.FullMethod, err)
		return nil, err
	}
	cID := ctx.Value(KeyClientID)

	resp, err = handler(ctx, req)
	log.Printf("%v method: %v err: %v", cID, info.FullMethod, err)
	return resp, err

}

// AuthStream checks a context for auth details
func AuthStream(serv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx, err := authContext(stream.Context())
	if err != nil {
		log.Printf("method: %v err: %v", info.FullMethod, err)
		return err
	}
	cID := ctx.Value(KeyClientID)

	// Wrap stream with new context
	stream = newAuthStream(ctx, stream)
	err = handler(serv, stream)
	log.Printf("%v method: %v err: %v", cID, info.FullMethod, err)
	return err
}

// authStream wraps a ServerStream so we can inject a context
type authStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the context
func (s *authStream) Context() context.Context {
	return s.ctx
}

// newAuthStream creates a ServerStream wrapped with a given context
func newAuthStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &authStream{stream, ctx}
}
