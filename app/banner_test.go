package app

import (
	"testing"
)

func TestPrintBanner_NoPanic(t *testing.T) {
	// Verify that PrintBanner does not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintBanner panicked: %v", r)
		}
	}()
	PrintBanner()
}

func TestBannerStyle_NotNil(t *testing.T) {
	// Verify that BannerStyle is properly initialized
	if BannerStyle.GetForeground() == nil {
		t.Error("BannerStyle foreground color should not be nil")
	}
}

func TestBannerArt_NotEmpty(t *testing.T) {
	// Verify that bannerArt is not empty
	if bannerArt == "" {
		t.Error("bannerArt should not be empty")
	}
}
