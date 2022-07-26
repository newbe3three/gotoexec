package util

import (
	"bytes"
	"github.com/kbinani/screenshot"
	"image"
	"image/png"

)

func Screenshot() []*image.RGBA {
	var images []*image.RGBA
	//获取当前活动屏幕数量
	i := screenshot.NumActiveDisplays()
	if i == 0 {

	}
	for j :=0; j <= i-1; j++ {
		image,_ := screenshot.CaptureDisplay(j)
		images = append(images, image)
	}
	return images
}


func ImageToByte(image *image.RGBA) []byte{
	buf := new(bytes.Buffer)
	png.Encode(buf,image)
	b := buf.Bytes()
	return b
}