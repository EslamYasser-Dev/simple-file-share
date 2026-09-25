package grpcapi

import (
	"context"
	"encoding/base64"
	"io"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}
func (noopLogger) Fatal(string, ...any) {}

const (
	adminUser = "admin"
	adminPass = "admin-secret"
)

type grpcFixture struct {
	files  filesharev1.FileServiceClient
	auth   filesharev1.AuthServiceClient
	shares filesharev1.ShareServiceClient
	events filesharev1.EventsServiceClient
	bus    *events.Bus
}

func newGRPCFixture(t *testing.T, enableAuth bool) *grpcFixture {
	t.Helper()
	dir := t.TempDir()

	index := memory.NewFileIndexRepository()
	localRepo := fs.NewLocalFileRepository(dir)
	fileRepo := fs.NewIndexedFileRepository(localRepo, index, nil)
	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(dir)
	shareRepo := fs.NewShareFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()

	if _, err := services.NewSeedAdminService(userRepo, hasher, fileRepo, scoper).Execute(adminUser, adminPass); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	authProvider := auth.NewUserAuthProvider(userRepo, hasher)
	authenticateService := services.NewAuthenticateService(authProvider)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)
	tokenService := services.NewTokenService(authenticateService, jwtManager, userRepo)
	roleCatalog := services.NewRoleCatalog(nil)

	listService := services.NewListFilesService(fileRepo, scoper)
	fileDownloadService := services.NewDownloadFileService(fileRepo, scoper)
	zipService := services.NewDownloadZipService(fileRepo, scoper)
	downloadService := services.NewDownloadService(fileDownloadService, zipService)
	uploadService := services.NewUploadService(fileRepo, scoper, index, userRepo, 0)
	updateService := services.NewUpdateFileContentService(fileRepo, scoper)
	createDirService := services.NewCreateDirectoryService(fileRepo, scoper)
	deleteService := services.NewDeletePathService(fileRepo, scoper)
	infoService := services.NewGetFileInfoService(fileRepo, scoper)
	searchService := services.NewSearchFilesService(index, scoper, nil)
	registerService := services.NewRegisterUserService(userRepo, hasher, fileRepo, scoper, true, 0)
	usersService := services.NewListUsersService(userRepo, index, scoper, roleCatalog)
	createShareService := services.NewCreateShareService(fileRepo, shareRepo, scoper)
	listSharesService := services.NewListSharesService(shareRepo, scoper, roleCatalog)
	revokeShareService := services.NewRevokeShareService(shareRepo, scoper, roleCatalog)
	eventBus := events.NewBus()

	userInfoService := services.NewUserInfoService(userRepo, index, scoper)
	authService := NewAuthService(registerService, usersService, authenticateService, userInfoService, true, tokenService)
	shareService := NewShareService(createShareService, listSharesService, revokeShareService)
	eventsService := NewEventsService(eventBus)
	fileService := NewFileService(
		listService, infoService, searchService, createDirService,
		deleteService, updateService, uploadService, downloadService,
	)

	srv, err := NewServer("0", noopLogger{}, &tls.InMemoryTLSCertGenerator{}, false, authenticateService, tokenService, enableAuth, authService, shareService, eventsService, fileService)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	lis := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return &grpcFixture{
		files:  filesharev1.NewFileServiceClient(conn),
		auth:   filesharev1.NewAuthServiceClient(conn),
		shares: filesharev1.NewShareServiceClient(conn),
		events: filesharev1.NewEventsServiceClient(conn),
		bus:    eventBus,
	}
}

func withBasic(ctx context.Context, user, pass string) context.Context {
	cred := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+cred)
}

func withBearer(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func TestGRPCLoginIssuesBearerToken(t *testing.T) {
	f := newGRPCFixture(t, true)

	login, err := f.auth.Login(context.Background(), &filesharev1.LoginRequest{Username: adminUser, Password: adminPass})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.GetAccessToken() == "" || login.GetTokenType() != "Bearer" {
		t.Fatalf("login response = %+v", login)
	}
	if login.GetExpiresIn() <= 0 {
		t.Fatalf("expires_in = %d, want > 0", login.GetExpiresIn())
	}

	if _, err := f.auth.Login(context.Background(), &filesharev1.LoginRequest{Username: adminUser, Password: "wrong"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("bad credentials code = %v, want Unauthenticated", status.Code(err))
	}

	ctx := withBearer(context.Background(), login.GetAccessToken())
	if _, err := f.files.ListFiles(ctx, &filesharev1.ListFilesRequest{}); err != nil {
		t.Fatalf("bearer list: %v", err)
	}
	me, err := f.auth.Me(ctx, &filesharev1.MeRequest{})
	if err != nil || me.GetUsername() != adminUser {
		t.Fatalf("me = %+v, err = %v", me, err)
	}

	if _, err := f.files.ListFiles(withBearer(context.Background(), "not-a-token"), &filesharev1.ListFilesRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("bad token code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestGRPCLogoutRevokesToken(t *testing.T) {
	f := newGRPCFixture(t, true)

	login, err := f.auth.Login(context.Background(), &filesharev1.LoginRequest{Username: adminUser, Password: adminPass})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	ctx := withBearer(context.Background(), login.GetAccessToken())
	if _, err := f.files.ListFiles(ctx, &filesharev1.ListFilesRequest{}); err != nil {
		t.Fatalf("bearer list before logout: %v", err)
	}

	logout, err := f.auth.Logout(ctx, &filesharev1.LogoutRequest{})
	if err != nil || logout.GetStatus() != "revoked" {
		t.Fatalf("logout = %+v, err = %v", logout, err)
	}
	if _, err := f.files.ListFiles(ctx, &filesharev1.ListFilesRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("revoked token code = %v, want Unauthenticated", status.Code(err))
	}

	if _, err := f.auth.Logout(context.Background(), &filesharev1.LogoutRequest{}); err != nil {
		t.Fatalf("logout without token should be a no-op: %v", err)
	}
}

func TestGRPCSharesRoundTrip(t *testing.T) {
	f := newGRPCFixture(t, true)
	ctx := withBasic(context.Background(), adminUser, adminPass)

	created, err := f.shares.CreateShare(ctx, &filesharev1.CreateShareRequest{Path: "note.txt", ExpiresInSeconds: 3600})
	if err == nil {
		// note.txt may not exist yet; a missing path is a domain NotFound.
		t.Fatalf("expected create share to fail for missing path, got %+v", created)
	}

	up, err := f.files.UploadFile(ctx)
	if err != nil {
		t.Fatalf("open upload: %v", err)
	}
	_ = up.Send(&filesharev1.UploadRequest{
		Payload: &filesharev1.UploadRequest_Metadata{Metadata: &filesharev1.UploadMetadata{Filename: "shared.txt"}},
	})
	_ = up.Send(&filesharev1.UploadRequest{Payload: &filesharev1.UploadRequest_Chunk{Chunk: []byte("hi")}})
	if _, err := up.CloseAndRecv(); err != nil {
		t.Fatalf("upload: %v", err)
	}

	created, err = f.shares.CreateShare(ctx, &filesharev1.CreateShareRequest{Path: "shared.txt", ExpiresInSeconds: 3600})
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	if created.GetToken() == "" || created.GetName() != "shared.txt" {
		t.Fatalf("share = %+v", created)
	}

	list, err := f.shares.ListShares(ctx, &filesharev1.ListSharesRequest{})
	if err != nil || len(list.GetShares()) != 1 {
		t.Fatalf("list shares = %+v, err = %v", list, err)
	}

	if _, err := f.shares.RevokeShare(ctx, &filesharev1.RevokeShareRequest{Token: created.GetToken()}); err != nil {
		t.Fatalf("revoke share: %v", err)
	}
	list, err = f.shares.ListShares(ctx, &filesharev1.ListSharesRequest{})
	if err != nil || len(list.GetShares()) != 0 {
		t.Fatalf("list after revoke = %+v, err = %v", list, err)
	}
}

func TestGRPCEventsSubscribeReceivesBroadcast(t *testing.T) {
	f := newGRPCFixture(t, true)
	base := withBasic(context.Background(), adminUser, adminPass)
	ctx, cancel := context.WithTimeout(base, 5*time.Second)
	defer cancel()

	stream, err := f.events.Subscribe(ctx, &filesharev1.SubscribeRequest{})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	// The server subscribes asynchronously after the call opens, so publish on
	// a short interval until the deadline instead of racing a single publish.
	go func() {
		for {
			f.bus.Publish(events.Event{Type: events.TypeQuota, User: "", Path: "shared.txt"})
			select {
			case <-ctx.Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()

	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv event: %v", err)
	}
	if msg.GetType() != events.TypeQuota || msg.GetPath() != "shared.txt" {
		t.Fatalf("event = %+v", msg)
	}
}

func TestGRPCAuthRequired(t *testing.T) {
	f := newGRPCFixture(t, true)

	if _, err := f.files.ListFiles(context.Background(), &filesharev1.ListFilesRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated list code = %v, want Unauthenticated", status.Code(err))
	}

	ctx := withBasic(context.Background(), adminUser, adminPass)
	resp, err := f.files.ListFiles(ctx, &filesharev1.ListFilesRequest{})
	if err != nil {
		t.Fatalf("authenticated list: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a list response")
	}
}

func TestGRPCAuthenticateAndPublicAuthInfo(t *testing.T) {
	f := newGRPCFixture(t, true)

	user, err := f.auth.Authenticate(context.Background(), &filesharev1.AuthenticateRequest{Username: adminUser, Password: adminPass})
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if user.GetUsername() != adminUser || !user.GetIsAdmin() {
		t.Fatalf("unexpected user: %+v", user)
	}

	if _, err := f.auth.Authenticate(context.Background(), &filesharev1.AuthenticateRequest{Username: adminUser, Password: "wrong"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("bad password code = %v, want Unauthenticated", status.Code(err))
	}

	info, err := f.auth.GetAuthInfo(context.Background(), &filesharev1.GetAuthInfoRequest{})
	if err != nil || !info.GetSignupEnabled() {
		t.Fatalf("auth info = %+v, err = %v", info, err)
	}
}

func TestGRPCUploadDownloadRoundTrip(t *testing.T) {
	f := newGRPCFixture(t, true)
	ctx := withBasic(context.Background(), adminUser, adminPass)

	const content = "hello grpc streaming\n"
	up, err := f.files.UploadFile(ctx)
	if err != nil {
		t.Fatalf("open upload: %v", err)
	}
	if err := up.Send(&filesharev1.UploadRequest{
		Payload: &filesharev1.UploadRequest_Metadata{Metadata: &filesharev1.UploadMetadata{Filename: "note.txt"}},
	}); err != nil {
		t.Fatalf("send metadata: %v", err)
	}
	half := len(content) / 2
	for _, part := range []string{content[:half], content[half:]} {
		if err := up.Send(&filesharev1.UploadRequest{
			Payload: &filesharev1.UploadRequest_Chunk{Chunk: []byte(part)},
		}); err != nil {
			t.Fatalf("send chunk: %v", err)
		}
	}
	res, err := up.CloseAndRecv()
	if err != nil {
		t.Fatalf("close upload: %v", err)
	}
	if res.GetPath() != "note.txt" || res.GetSize() != int64(len(content)) {
		t.Fatalf("upload result = %+v", res)
	}

	stream, err := f.files.DownloadFile(ctx, &filesharev1.DownloadFileRequest{Path: "note.txt"})
	if err != nil {
		t.Fatalf("open download: %v", err)
	}
	var got []byte
	var filename string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("recv chunk: %v", err)
		}
		if filename == "" {
			filename = chunk.GetFilename()
		}
		got = append(got, chunk.GetData()...)
	}
	if filename != "note.txt" {
		t.Fatalf("filename = %q, want note.txt", filename)
	}
	if string(got) != content {
		t.Fatalf("downloaded %q, want %q", got, content)
	}
}

func TestGRPCValidationErrorMapping(t *testing.T) {
	f := newGRPCFixture(t, true)
	ctx := withBasic(context.Background(), adminUser, adminPass)

	_, err := f.files.CreateDirectory(ctx, &filesharev1.CreateDirectoryRequest{Path: "../escape"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("traversal code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestGRPCAuthDisabledIsSystemAdmin(t *testing.T) {
	f := newGRPCFixture(t, false)

	user, err := f.auth.Me(context.Background(), &filesharev1.MeRequest{})
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if !user.GetIsAdmin() {
		t.Fatalf("expected system admin when auth disabled, got %+v", user)
	}
}

var _ ports.Logger = noopLogger{}
