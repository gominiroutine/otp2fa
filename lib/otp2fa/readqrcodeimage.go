package otp2fa

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func ReadQrCodeImage(data []byte) (string, error) {
	var imgData image.Image
	var err error
	if imgData, err = png.Decode(bytes.NewReader(data)); err != nil {
		if imgData, err = jpeg.Decode(bytes.NewReader(data)); err != nil {
			w, h := 256, 256
			var icon *oksvg.SvgIcon
			icon, err = oksvg.ReadIconStream(bytes.NewReader(data))
			if err != nil {
				return "", err
			}
			icon.SetTarget(0, 0, float64(w), float64(h))
			rgba := image.NewRGBA(image.Rect(0, 0, w, h))
			icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)
			imgData = rgba
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
