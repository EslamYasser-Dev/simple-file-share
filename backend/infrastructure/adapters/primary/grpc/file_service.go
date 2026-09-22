package grpcapi

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// FileService adapts the file use cases to the gRPC transport.
type FileService struct {
	filesharev1.UnimplementedFileServiceServer

	list      *services.ListFilesService
	info      *services.GetFileInfoService
	search    *services.SearchFilesService
	createDir *services.CreateDirectoryService
	delete    *services.DeletePathService
	update    *services.UpdateFileContentService
	upload    *services.UploadService
	download  *services.DownloadService
}

func NewFileService(
	list *services.ListFilesService,
	info *services.GetFileInfoService,
	search *services.SearchFilesService,
	createDir *services.CreateDirectoryService,
	delete *services.DeletePathService,
	update *services.UpdateFileContentService,
	upload *services.UploadService,
	download *services.DownloadService,
) *FileService {
	return &FileService{
		list:      list,
		info:      info,
		search:    search,
		createDir: createDir,
		delete:    delete,
		update:    update,
		upload:    upload,
		download:  download,
	}
}

func (s *FileService) ListFiles(ctx context.Context, req *filesharev1.ListFilesRequest) (*filesharev1.ListFilesResponse, error) {
	page, err := s.list.Execute(authctx.UserFromContext(ctx), req.GetPath())
	if err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.ListFilesResponse{Files: toProtoFileInfos(page.Files)}, nil
}

func (s *FileService) GetFileInfo(ctx context.Context, req *filesharev1.GetFileInfoRequest) (*filesharev1.FileInfo, error) {
	info, err := s.info.Execute(authctx.UserFromContext(ctx), req.GetPath())
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoFileInfo(info), nil
}

func (s *FileService) SearchFiles(ctx context.Context, req *filesharev1.SearchFilesRequest) (*filesharev1.SearchFilesResponse, error) {
	files, err := s.search.Execute(authctx.UserFromContext(ctx), req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.SearchFilesResponse{Files: toProtoFileInfos(files)}, nil
}

func (s *FileService) CreateDirectory(ctx context.Context, req *filesharev1.CreateDirectoryRequest) (*filesharev1.CreateDirectoryResponse, error) {
	if err := s.createDir.Execute(authctx.UserFromContext(ctx), req.GetPath()); err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.CreateDirectoryResponse{Message: "directory created"}, nil
}

func (s *FileService) DeletePath(ctx context.Context, req *filesharev1.DeletePathRequest) (*filesharev1.DeletePathResponse, error) {
	if err := s.delete.Execute(authctx.UserFromContext(ctx), req.GetPath()); err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.DeletePathResponse{Message: "deleted"}, nil
}

func (s *FileService) UpdateFileContent(ctx context.Context, req *filesharev1.UpdateFileContentRequest) (*filesharev1.UpdateFileContentResponse, error) {
	size, err := s.update.Execute(authctx.UserFromContext(ctx), req.GetPath(), req.GetContent())
	if err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.UpdateFileContentResponse{Message: "file updated", Size: size}, nil
}

// UploadFile receives a metadata message followed by chunks and streams them
// straight into the upload use case without buffering the whole file.
func (s *FileService) UploadFile(stream filesharev1.FileService_UploadFileServer) error {
	first, err := stream.Recv()
	if err == io.EOF {
		return status.Error(codes.InvalidArgument, "empty upload stream")
	}
	if err != nil {
		return err
	}
	meta := first.GetMetadata()
	if meta == nil {
		return status.Error(codes.InvalidArgument, "first upload message must carry metadata")
	}

	pr, pw := io.Pipe()
	go pumpUpload(stream, pw)

	uploads, execErr := s.upload.Execute(authctx.UserFromContext(stream.Context()), []models.UploadPart{{
		Name:        meta.GetFilename(),
		Destination: meta.GetPath(),
		Content:     pr,
	}})
	_ = pr.Close()
	if execErr != nil {
		return toStatus(execErr)
	}

	uploaded := uploads[0]
	return stream.SendAndClose(&filesharev1.UploadResponse{
		Path: uploaded.Filename,
		Size: uploaded.Size,
	})
}

// pumpUpload forwards chunk messages into the pipe until the client closes the
// stream. Closing the pipe signals EOF to the upload use case.
func pumpUpload(stream filesharev1.FileService_UploadFileServer, pw *io.PipeWriter) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			_ = pw.Close()
			return
		}
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if chunk := msg.GetChunk(); len(chunk) > 0 {
			if _, err := pw.Write(chunk); err != nil {
				return
			}
		}
	}
}

// DownloadFile streams a file (or zipped directory) in chunks. The first chunk
// carries the resolved filename and content type.
func (s *FileService) DownloadFile(req *filesharev1.DownloadFileRequest, stream filesharev1.FileService_DownloadFileServer) error {
	download, err := s.download.Execute(authctx.UserFromContext(stream.Context()), req.GetPath())
	if err != nil {
		return toStatus(err)
	}
	defer download.Stream.Close()

	buf := make([]byte, 64*1024)
	first := true
	for {
		n, readErr := download.Stream.Read(buf)
		if n > 0 {
			chunk := &filesharev1.DownloadChunk{Data: buf[:n]}
			if first {
				chunk.Filename = download.Filename
				chunk.ContentType = download.ContentType
				first = false
			}
			if sendErr := stream.Send(chunk); sendErr != nil {
				return sendErr
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return status.Error(codes.Internal, "stream read failed")
		}
	}
}

func toProtoFileInfo(f *models.FileInfo) *filesharev1.FileInfo {
	if f == nil {
		return nil
	}
	info := &filesharev1.FileInfo{
		Name:  f.Name,
		Path:  f.Path,
		Size:  f.Size,
		IsDir: f.IsDir,
	}
	if !f.Modified.IsZero() {
		info.Modified = timestamppb.New(f.Modified)
	}
	return info
}

func toProtoFileInfos(files []*models.FileInfo) []*filesharev1.FileInfo {
	out := make([]*filesharev1.FileInfo, 0, len(files))
	for _, f := range files {
		out = append(out, toProtoFileInfo(f))
	}
	return out
}
