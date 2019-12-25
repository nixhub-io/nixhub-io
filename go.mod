module gitlab.com/nixhub/nixhub.io

go 1.13

replace github.com/bwmarrin/discordgo => github.com/diamondburned/discordgo v0.12.2-0.20191225065313-b45ce3585027

require (
	git.sr.ht/~diamondburned/geotz v0.0.0-20191224053446-e16e6d4aaf84
	git.sr.ht/~diamondburned/gocad v0.0.0-20191225012957-705675bb6b23
	github.com/alecthomas/chroma v0.7.0
	github.com/bwmarrin/discordgo v0.20.1
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/markbates/pkger v0.13.0
	github.com/pkg/errors v0.8.1
	github.com/russross/blackfriday v1.5.2 // indirect
	gitlab.com/shihoya-inc/errchi v0.0.0-20191218161822-dd2516e8def9
)
