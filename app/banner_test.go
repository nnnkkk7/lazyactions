package app

import (
	"testing"
)

func TestPrintBanner_NoPanic(t *testing.T) {
	// PrintBanner がパニックしないことを確認
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintBanner panicked: %v", r)
		}
	}()
	PrintBanner()
}

func TestBannerStyle_NotNil(t *testing.T) {
	// BannerStyle が正しく初期化されていることを確認
	if BannerStyle.GetForeground() == nil {
		t.Error("BannerStyle foreground color should not be nil")
	}
}

func TestBannerArt_NotEmpty(t *testing.T) {
	// bannerArt が空でないことを確認
	if bannerArt == "" {
		t.Error("bannerArt should not be empty")
	}
}
