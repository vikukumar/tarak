package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// Bilinear / Nearest Neighbor interpolation
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := srcBounds.Min.X + (x*srcW)/width
			srcY := srcBounds.Min.Y + (y*srcH)/height
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func savePNG(img image.Image, path string) error {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func createICO(images []image.Image, path string) error {
	var pngBuffers [][]byte
	for _, img := range images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		pngBuffers = append(pngBuffers, buf.Bytes())
	}

	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// ICONDIR Header
	binary.Write(f, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(f, binary.LittleEndian, uint16(1)) // Type (1 = ICO)
	binary.Write(f, binary.LittleEndian, uint16(len(images)))

	offset := uint32(6 + len(images)*16)
	for i, img := range images {
		b := img.Bounds()
		w := byte(b.Dx())
		if b.Dx() >= 256 {
			w = 0
		}
		h := byte(b.Dy())
		if b.Dy() >= 256 {
			h = 0
		}

		f.Write([]byte{w, h, 0, 0})                     // Width, Height, ColorCount, Reserved
		binary.Write(f, binary.LittleEndian, uint16(1))  // Color planes
		binary.Write(f, binary.LittleEndian, uint16(32)) // Bits per pixel
		binary.Write(f, binary.LittleEndian, uint32(len(pngBuffers[i])))
		binary.Write(f, binary.LittleEndian, offset)
		offset += uint32(len(pngBuffers[i]))
	}

	for _, buf := range pngBuffers {
		f.Write(buf)
	}
	return nil
}

func copyFile(src, dst string) error {
	_ = os.MkdirAll(filepath.Dir(dst), 0755)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func main() {
	srcPath := "docs/assets/icon.png"
	file, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", srcPath, err)
		return
	}
	defer file.Close()

	srcImg, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return
	}

	sizes := map[string]int{
		"favicon-16x16.png":           16,
		"favicon-32x32.png":           32,
		"favicon-48x48.png":           48,
		"apple-touch-icon.png":        180,
		"apple-touch-icon-180x180.png": 180,
		"android-chrome-192x192.png":  192,
		"android-chrome-512x512.png":  512,
		"icon-128x128.png":            128,
		"icon-256x256.png":            256,
		"icon.png":                    512,
	}

	targets := []string{
		"docs/assets",
		"dashboard/public/assets",
		"dashboard/public",
		"internal/ui/dist/assets",
		"internal/ui/dist",
	}

	var icoImages []image.Image
	for _, sz := range []int{16, 32, 48, 64, 128, 256} {
		icoImages = append(icoImages, resizeImage(srcImg, sz, sz))
	}

	for _, targetDir := range targets {
		for filename, sz := range sizes {
			resized := resizeImage(srcImg, sz, sz)
			outPath := filepath.Join(targetDir, filename)
			if err := savePNG(resized, outPath); err != nil {
				fmt.Printf("Error saving %s: %v\n", outPath, err)
			}
		}

		icoPath := filepath.Join(targetDir, "favicon.ico")
		if err := createICO(icoImages, icoPath); err != nil {
			fmt.Printf("Error creating %s: %v\n", icoPath, err)
		}

		// Copy logos
		for _, logo := range []string{
			"tarak_logo_horizontal.png",
			"tarak_logo_vertical.png",
			"tarak_logo_poster.png",
			"tarak_poster_logo.png",
			"og-image.jpg",
			"tarak_github.jpg",
			"tarak_icon.jpg",
			"tarak_logo.jpg",
		} {
			srcLogo := filepath.Join("docs/assets", logo)
			dstLogo := filepath.Join(targetDir, logo)
			if _, err := os.Stat(srcLogo); err == nil {
				_ = copyFile(srcLogo, dstLogo)
			}
		}
	}

	fmt.Println("Successfully generated all icons and logos across all directories!")
}
