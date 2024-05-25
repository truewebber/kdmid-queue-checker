package image

import (
	"bytes"
	"fmt"

	"image"
	"image/draw"
	"image/png"
)

type (
	PNG          []byte
	CroppingRect struct {
		X0, Y0, X1, Y1 int
	}
	CroppingFunc func(height, width int) CroppingRect
)

func Crop(pngImage PNG, cropFunc CroppingFunc) (PNG, error) {
	srcImage, _, err := image.Decode(bytes.NewReader(pngImage))
	if err != nil {
		return nil, fmt.Errorf("could not decode source image: %w", err)
	}

	rectToCrop := cropFunc(srcImage.Bounds().Dy(), srcImage.Bounds().Dx())
	rect := image.Rect(rectToCrop.X0, rectToCrop.Y0, rectToCrop.X1, rectToCrop.Y1)

	croppedImage := image.NewRGBA(rect)
	draw.Draw(croppedImage, rect, srcImage, rect.Min, draw.Src)

	var buffer bytes.Buffer

	err = png.Encode(&buffer, croppedImage)
	if err != nil {
		return nil, fmt.Errorf("could not encode output image: %w", err)
	}

	return buffer.Bytes(), nil
}
