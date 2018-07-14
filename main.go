package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func reduce(u uint32) int {
	return int(u >> 8)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getRowCol(i, width, height int) (row int, col int) {
	i = i % (width * height)
	return i % width, i / width
}

func getNextInOrder(r *rand.Rand, max int, visited map[int]struct{}) int {
	for {
		poss := r.Intn(max)
		if _, has := visited[poss]; !has {
			visited[poss] = struct{}{}
			return poss
		}
	}
}

func encodeImgAt(img *image.RGBA, x, y int, c byte, uniq uint8) {
	r, g, b, _ := img.RGBAAt(x, y).RGBA()
	img.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(c) ^ uniq})
}

func decodeImgAt(img image.Image, x, y int, uniq uint8) byte {
	_, _, _, a := img.At(x, y).RGBA()
	return byte(uint8(a>>8) ^ uniq)
}

func Clone(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			result.Set(x, y, img.At(x, y))
		}
	}
	return result
}

func Init(filepath string) image.Image {
	file, err := os.Open(filepath)
	defer file.Close()
	check(err)

	img, _, err := image.Decode(file)
	check(err)

	return img
}

func GenKey(img image.Image, msg string) (int, int, int) {
	bounds := img.Bounds()

	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	width := bounds.Dx()
	height := bounds.Dy()
	x := ra.Intn(width)
	y := ra.Intn(height)

	fmt.Printf("%d %d %d", x, y, len(msg))

	return x, y, len(msg)
}

func GenUniqOrder(img image.Image, x, y, msgLen int) (uint8, *rand.Rand) {
	r, g, b, _ := img.At(x, y).RGBA()
	uniq := uint8(x ^ y ^ reduce(r) ^ reduce(g) ^ reduce(b) ^ msgLen)
	order := rand.New(rand.NewSource(int64(x + y + reduce(r) + reduce(g) + reduce(b) + msgLen)))
	return uniq, order
}

func Measurements(i image.Image) (int, int, int) {
	bounds := i.Bounds()
	return bounds.Dx(), bounds.Dy(), bounds.Dx() * bounds.Dy()
}

func Encode(filepath, msg string) {
	img := Init(filepath)
	x, y, msgLen := GenKey(img, msg)
	uniq, order := GenUniqOrder(img, x, y, msgLen)
	width, height, pixelCount := Measurements(img)

  if msgLen > pixelCount {
    log.Fatal("There is not enough space to store the message in the image")
  }

	clone := Clone(img)
	visited := make(map[int]struct{})
	for i := 0; i < msgLen; i++ {
		next := getNextInOrder(order, pixelCount, visited)
		row, col := getRowCol(next, width, height)
		encodeImgAt(clone, row, col, msg[i], uniq + uint8(i))
	}

	out, err := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	check(err)

	defer out.Close()
	err = png.Encode(out, clone)
	check(err)
}

func Decode(filepath, key string) {
	img := Init(filepath)

	parts := strings.Split(key, " ")
	x, err := strconv.Atoi(parts[0])
	check(err)
	y, err := strconv.Atoi(parts[1])
	check(err)
	msgLen, err := strconv.Atoi(parts[2])
	check(err)

	width, height, pixelCount := Measurements(img)
	uniq, order := GenUniqOrder(img, x, y, msgLen)

	result := make([]byte, msgLen)
	visited := make(map[int]struct{})
	for i := 0; i < msgLen; i++ {
		next := getNextInOrder(order, pixelCount, visited)
		row, col := getRowCol(next, width, height)
		result = append(result, decodeImgAt(img, row, col, uniq + uint8(i)))
	}
	fmt.Println(string(result))
}

func main() {
	var filepath = flag.String("f", "", "The path to the image to run on")
	var message = flag.String("msg", "", "The message to be encoded")
	var key = flag.String("key", "", "The key to be used when decoding the message")

	flag.Parse()

	if len(*filepath) == 0 {
		log.Fatal("Must pass filepath")
	}

	if len(*key) != 0 && len(*message) != 0 {
		log.Fatal("Can only pass one of key or msg")
	} else if len(*key) != 0 {
		Decode(*filepath, *key)
		return
	} else if len(*message) != 0 {
		Encode(*filepath, *message)
		return
	}
	log.Fatal("Must pass key or msg")
}
