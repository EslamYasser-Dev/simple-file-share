package services

import (
	"fmt"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func seedSocialBench(b *testing.B, events, follows int) (*TimelineService, *models.User) {
	b.Helper()
	dir := b.TempDir()
	fileRepo := fs.NewLocalFileRepository(dir)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()

	users := make([]*models.User, follows+2)
	for i := range users {
		name := fmt.Sprintf("user%04d", i)
		if err := userRepo.CreateUser(&models.User{Username: name}); err != nil {
			b.Fatal(err)
		}
		u, _ := userRepo.FindByUsername(name)
		users[i] = u
	}
	viewer := users[0]
	followRepo := fs.NewFollowFileRepository(dir)
	for _, u := range users[1 : follows+1] {
		if err := followRepo.Follow(viewer.Username, u.Username); err != nil {
			b.Fatal(err)
		}
	}
	timelineRepo := fs.NewTimelineFileRepository(dir)
	visRepo := fs.NewVisibilityFileRepository(dir)
	_ = fileRepo
	_ = scoper
	_ = visRepo
	svc := NewTimelineService(timelineRepo, followRepo)
	for i := 0; i < events; i++ {
		owner := users[1+(i%(follows+1))].Username
		vis := models.VisibilityPrivate
		if i%2 == 0 {
			vis = models.VisibilityLink
		}
		if _, err := svc.Record(models.TimelineUpload, &models.User{Username: owner}, "f.txt", 1, vis); err != nil {
			b.Fatal(err)
		}
	}
	return svc, viewer
}

func BenchmarkFeed2000Events500Follows(b *testing.B) {
	svc, viewer := seedSocialBench(b, 2000, 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Feed(viewer, "", 20); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTimelineList(b *testing.B) {
	svc, _ := seedSocialBench(b, 2000, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Feed(&models.User{Username: "user0000"}, "", 20); err != nil {
			b.Fatal(err)
		}
	}
}
