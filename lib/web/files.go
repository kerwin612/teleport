/*
Copyright 2018 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/ssh"

	"github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/keys"
	"github.com/gravitational/teleport/lib/client"

	// webdefaults "github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/reversetunnel"
	"github.com/gravitational/teleport/lib/sshutils/scp"
)

// fileTransferRequest describes HTTP file transfer request
type fileTransferRequest struct {
	// Server describes a server to connect to (serverId|hostname[:port]).
	server string
	// Login is Linux username to connect as.
	login string
	// Namespace is node namespace.
	namespace string
	// Cluster is the name of the remote cluster to connect to.
	cluster string
	// remoteLocation is file remote location
	remoteLocation string
	// filename is a file name
	filename string
}

func (h *Handler) fileTransferhandler(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
	sctx *SessionContext,
	site reversetunnel.RemoteSite,
) (interface{}, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errMsg := "Error upgrading to websocket"
		http.Error(w, errMsg, http.StatusInternalServerError)
	}
	defer ws.Close()

	query := r.URL.Query()
	req := fileTransferRequest{
		cluster:        site.GetName(),
		login:          p.ByName("login"),
		server:         p.ByName("server"),
		remoteLocation: query.Get("location"),
		filename:       query.Get("filename"),
		namespace:      defaults.Namespace,
	}

	tc, err := h.createClient(ws, r, p, sctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	nodeClient, err := connectToNode(r.Context(), tc, ws, r, p, sctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	h.handler(r.Context(), ws, sctx, req, nodeClient)
	// return nil response here because we upgraded to wss
	return nil, nil
}

type FileTransferStream struct {
	// buffer is a buffer used to store the remaining payload data if it did not
	// fit into the buffer provided by the callee to Read method
	buffer []byte

	// once ensures that resizeC is closed at most one time
	once sync.Once

	// mu protects writes to ws
	mu sync.Mutex
	// ws the connection to the UI
	ws *websocket.Conn
}

func NewFileTransferStream(ws *websocket.Conn, opts ...func(*FileTransferStream)) (*FileTransferStream, error) {
	switch {
	case ws == nil:
		return nil, trace.BadParameter("required parameter ws not provided")
	}

	t := &FileTransferStream{
		ws: ws,
		// encoder: unicode.UTF8.NewEncoder(),
		// decoder: unicode.UTF8.NewDecoder(),
	}

	return t, nil
}

// type FileTransferEnvelope struct {
// 	Type    string `json:"type"`
// 	Payload []byte `json:"payload"`
// }

const (
	WebsocketRawData      = "r"
	WebsocketMfaChallenge = "n"
)

func (f *FileTransferStream) Write(data []byte) (n int, err error) {
	envelope := &Envelope{
		Type:    WebsocketRawData,
		Payload: data,
	}

	fmt.Printf("%+v\n", envelope)

	// envelopeBytes, _ := proto.Marshal(envelope)
	// Send bytes over the websocket to the web client.
	f.mu.Lock()
	err = f.ws.WriteMessage(websocket.BinaryMessage, data)
	f.mu.Unlock()
	if err != nil {
		return 0, trace.Wrap(err)
	}

	return len(data), nil
}

func (h *Handler) handler(ctx context.Context, ws *websocket.Conn, sctx *SessionContext, req fileTransferRequest, nc *client.NodeClient) {
	stream, err := NewFileTransferStream(ws)
	if err != nil {
		return
	}

	cmd, err := scp.CreateSocketDownload(scp.WebsocketFileRequest{
		RemoteLocation: req.remoteLocation,
		Response:       stream,
		FileName:       req.filename,
		User:           sctx.GetUser(),
	})
	if err != nil {
		return
	}

	nc.ExecuteSCP(ctx, cmd)
}

func (h *Handler) createClient(ws *websocket.Conn, r *http.Request, p httprouter.Params, sctx *SessionContext, site reversetunnel.RemoteSite) (*client.TeleportClient, error) {
	ctx := r.Context()
	clientConfig, err := makeTeleportClientConfig(ctx, sctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	clientConfig.Namespace = defaults.Namespace
	if !types.IsValidNamespace(defaults.Namespace) {
		return nil, trace.BadParameter("invalid namespace %q", clientConfig.Namespace)
	}

	clientConfig.HostLogin = p.ByName("login")
	if clientConfig.HostLogin == "" {
		return nil, trace.BadParameter("missing login")
	}

	servers, err := sctx.cfg.RootClient.GetNodes(ctx, clientConfig.Namespace)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	server := p.ByName("server")
	hostName, hostPort, err := resolveServerHostPort(server, servers)
	if err != nil {
		return nil, trace.BadParameter("invalid server name %q: %v", server, err)
	}

	clientConfig.SiteName = site.GetName()
	if err := clientConfig.ParseProxyHost(h.ProxyHostPort()); err != nil {
		return nil, trace.BadParameter("failed to parse proxy address: %v", err)
	}
	clientConfig.Host = hostName
	clientConfig.HostPort = hostPort
	clientConfig.ClientAddr = r.RemoteAddr

	tc, err := client.NewClient(clientConfig)
	if err != nil {
		return nil, trace.BadParameter("failed to create client: %v", err)
	}
	return tc, nil
}

func connectToNode(ctx context.Context, tc *client.TeleportClient, ws *websocket.Conn, r *http.Request, p httprouter.Params, sctx *SessionContext, site reversetunnel.RemoteSite) (*client.NodeClient, error) {
	stream, err := NewTerminalStream(ws)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	pk, err := keys.ParsePrivateKey(sctx.cfg.Session.GetPriv())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	pc, err := tc.ConnectToProxy(ctx)

	details, err := pc.ClusterDetails(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	nodeAddrs, err := tc.GetTargetNodes(ctx, pc)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if len(nodeAddrs) == 0 {
		return nil, trace.BadParameter("no target host specified")
	}

	key, err := pc.IssueUserCertsWithMFA(ctx, client.ReissueParams{
		NodeName:       nodeAddrs[0],
		RouteToCluster: site.GetName(),
		ExistingCreds: &client.Key{
			PrivateKey: pk,
			Cert:       sctx.cfg.Session.GetPub(),
			TLSCert:    sctx.cfg.Session.GetTLSCert(),
		},
	}, promptMFAChallenge(stream, protobufMFACodec{}))

	am, err := key.AsAuthMethod()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	user := tc.Config.HostLogin

	nodeClient, err := pc.ConnectToNode(ctx,
		client.NodeDetails{Addr: nodeAddrs[0], Namespace: tc.Namespace, Cluster: tc.SiteName},
		user,
		details,
		[]ssh.AuthMethod{am})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return nodeClient, nil
}
