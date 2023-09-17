package banner

import (
	"embed"
	"github.com/dstgo/filebox"
)

func PrintLnBanner(banner string) {
	println(banner)
}

func PrintlnBannerFs(filename string) {
	bannerStr, err := filebox.ReadFileString(filename)
	if err != nil {
		PrintLnBanner(bannerStr)
	}
}

func PrintlnBannerEmbed(fs embed.FS, filename string) {
	bytes, err := fs.ReadFile(filename)
	if err != nil {
		PrintLnBanner(string(bytes))
	}
}
