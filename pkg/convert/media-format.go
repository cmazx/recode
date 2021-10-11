package convert

type MediaEncoding string

const (
	ImageJpeg MediaEncoding = "jpg"
	ImagePng  MediaEncoding = "png"
)

type MediaFormat struct {
	Name     string
	Width    uint
	Height   uint
	Quality  uint8
	Encoding MediaEncoding
}
