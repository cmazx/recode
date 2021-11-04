package convert

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/gographics/imagick.v2/imagick"
	"log"
)

/*
func Process(format MediaFormat, sourcePath string, targetDir string) (string, error) {
	outPath := targetDir + "/" + uuid.New().String() + "." + string(format.Encoding)
	file, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	_, err = file.Write([]byte("test"))
	if err != nil {
		return "", err
	}
	return outPath, nil
}
*/

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

	fmt.Println("Format resize method " + format.Method)
	switch format.Method {
	case MethodCrop:
		err = Crop(mw, format)
	case MethodBox:
		err = Box(mw, format)
	}
	if err != nil {
		return "", err
	}

	err = mw.SetCompressionQuality(format.Quality)
	if err != nil {
		return "", err
	}

	cols := mw.GetImageWidth()
	rows := mw.GetImageHeight()
	fmt.Printf("Result image  %dx%d\n", cols, rows)

	err = mw.WriteImage(outPath)
	if err != nil {
		return "", err
	}

	return outPath, err
}

func Crop(mw *imagick.MagickWand, format MediaFormat) error {
	cols := mw.GetImageWidth()
	rows := mw.GetImageHeight()
	fmt.Printf("[CROP] image size %dx%d\n", cols, rows)
	ratio := float32(rows) / float32(cols)
	height := uint(float32(format.Width) * ratio)
	fmt.Printf("[CROP]Resize image  %dx%d to %dx%d\n", cols, rows, format.Width, height)
	err := mw.ScaleImage(format.Width, height)
	if err != nil {
		return err
	}
	if format.Height > 0 && height > format.Height {
		diff := int(height/2 - format.Height/2)
		fmt.Printf("[CROP]Crop image %dx%d to %dx%d - 0x%d\n", format.Width, height, format.Width, format.Height, diff)
		err = mw.CropImage(format.Width, format.Height, 0, diff)
		if err != nil {
			return err
		}
	}

	return nil
}
func Box(sourceImage *imagick.MagickWand, format MediaFormat) error {
	cols := sourceImage.GetImageWidth()
	rows := sourceImage.GetImageHeight()
	fmt.Printf("[BOX] image size %dx%d\n", cols, rows)
	ratio := float32(rows) / float32(cols)

	width := format.Width
	height := format.Height
	if ratio > 1 {
		width = uint(float32(format.Height) / ratio)
		if width > format.Width {
			height = format.Height
			width = uint(float32(format.Height) / ratio)
		}
	} else {
		height = uint(float32(format.Width) * ratio)
		if height > format.Height {
			height = format.Height
			width = uint(float32(format.Height) / ratio)
		}
	}

	fmt.Printf("[BOX]Resize image %dx%d to %dx%d\n", cols, rows, width, height)
	err := sourceImage.ScaleImage(width, height)
	if err != nil {
		return err
	}

	pw := imagick.NewPixelWand()
	pw.SetAlpha(0)

	fmt.Printf("[BOX]Box image %dx%d to %dx%d\n", width, height, format.Width, format.Height)
	rowB := (format.Width - width) / 2
	if rowB < 0 {
		rowB = 0
	}
	rowH := (format.Height - height) / 2
	if rowH < 0 {
		rowH = 0
	}
	fmt.Printf("[BOX]Add border %dx%d\n", rowB, rowH)
	err = sourceImage.BorderImage(pw, rowB, rowH)
	if err != nil {
		return err
	}

	return nil
}
