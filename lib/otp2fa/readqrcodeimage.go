package otp2fa

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func ReadQrCodeImage(data []byte) (string, error) {
	var imgData image.Image
	var err error
	if imgData, err = png.Decode(bytes.NewReader(data)); err != nil {
		if imgData, err = jpeg.Decode(bytes.NewReader(data)); err != nil {
			return "", err
		}
	}
	// Prepare BinaryBitmap
	bitmap, err := gozxing.NewBinaryBitmapFromImage(imgData)
	if err != nil {
		return "", err
	}
	// Decode the Bitmap as QR Code
	result, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
