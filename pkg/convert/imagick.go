package convert

import (
	"github.com/google/uuid"
	"gopkg.in/gographics/imagick.v2/imagick"
	"log"
)

func Process(format MediaFormat, sourcePath string, targetDir string) (string, error) {
	imagick.Initialize()
	defer imagick.Terminate()
	outPath := targetDir + "/" + uuid.New().String() + "." + string(format.Encoding)
	log.Printf("Convert %s to %s with format: %v\n", sourcePath, outPath, format)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err := mw.ReadImage(sourcePath)
	if err != nil {
		return "", err
	}
	err = mw.ResizeImage(format.Height, format.Width, imagick.FILTER_TRIANGLE, 0)
	if err != nil {
		return "", err
	}

	err = mw.WriteImage(outPath)
	if err != nil {
		return "", err
	}

	return outPath, err
}
