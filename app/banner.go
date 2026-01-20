package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ASCIIアートバナー
const bannerArt = `
 _                      _        _   _
| |    __ _ _____   _  / \   ___| |_(_) ___  _ __  ___
| |   / _' |_  / | | |/ _ \ / __| __| |/ _ \| '_ \/ __|
| |__| (_| |/ /| |_| / ___ \ (__| |_| | (_) | | | \__ \
|_____\__,_/___|\__, /_/   \_\___|\__|_|\___/|_| |_|___/
               |___/
`

// BannerStyle defines the cyan color style for the banner
var BannerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

// PrintBanner prints the ASCII art banner in cyan with a brief delay
func PrintBanner() {
	fmt.Println(BannerStyle.Render(bannerArt))
	time.Sleep(500 * time.Millisecond)
}
