package convert

type MediaEncoding string

const (
	ImageJpeg MediaEncoding = "jpg"
	ImagePng  MediaEncoding = "png"
)

type ResizeMethod string

const (
	MethodCrop ResizeMethod = "crop"
	MethodBox  ResizeMethod = "box"
)

type MediaFormat struct {
	Name     string
	Width    uint
	Height   uint
	Quality  uint
	Encoding MediaEncoding
	Method   ResizeMethod
}
