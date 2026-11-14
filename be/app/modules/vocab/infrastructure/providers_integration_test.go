package infrastructure

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// Calls the real Pixabay and Unsplash APIs. Run with PROVIDERS_INTEGRATION=1 (keys from be/.env).
func TestImageProvidersLive(t *testing.T) {
	if os.Getenv("PROVIDERS_INTEGRATION") != "1" {
		t.Skip("set PROVIDERS_INTEGRATION=1 to call the real image APIs")
	}
	_ = godotenv.Load("../../../../.env") // be/.env
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if key := os.Getenv("PIXABAY_API_KEY"); key != "" {
		p := NewPixabayProvider(key)
		imgs, err := p.Search(ctx, "person choosing between two paths", 6)
		if err != nil || len(imgs) == 0 {
			t.Fatalf("pixabay search: %v (%d results)", err, len(imgs))
		}
		if !strings.HasPrefix(imgs[0].URL, "https://") || imgs[0].LinkExpiresAt.IsZero() {
			t.Fatalf("pixabay image: %+v", imgs[0])
		}
		fresh, err := p.Lookup(ctx, imgs[0].ProviderID)
		if err != nil || fresh == nil || fresh.ProviderID != imgs[0].ProviderID {
			t.Fatalf("pixabay lookup: %v %+v", err, fresh)
		}
		if gone, err := p.Lookup(ctx, "99999999999"); gone != nil || err != nil {
			t.Fatalf("a removed photo should return (nil, nil): %+v %v", gone, err)
		}
		t.Logf("pixabay: %d results, lookup and removed-photo detection ok", len(imgs))
	}
	if key := os.Getenv("UNSPLASH_ACCESS_KEY"); key != "" {
		p := NewUnsplashProvider(key)
		imgs, err := p.Search(ctx, "person choosing between two paths", 6)
		if err != nil || len(imgs) == 0 {
			t.Fatalf("unsplash search: %v (%d results)", err, len(imgs))
		}
		if imgs[0].AuthorName == "" || !strings.Contains(imgs[0].AuthorURL, "utm_source=") || imgs[0].DownloadLocation == "" {
			t.Fatalf("unsplash attribution missing: %+v", imgs[0])
		}
		t.Logf("unsplash: %d results", len(imgs))
	}
}
