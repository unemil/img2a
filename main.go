package main

import (
	"fmt"
	"image"
	"os"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/spf13/cobra"
)

const (
	ansiOverhead = 25

	defaultWidth int     = 150
	defaultRatio float64 = 2.3
)

var charset = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`\\'. "

func brightness(r, g, b uint32) int {
	return int(.2126*float64(uint8(r>>8)) + .7152*float64(uint8(g>>8)) + .0722*float64(uint8(b>>8)))
}

func areaSampling(image image.Image, startX, startY, endX, endY int) (r, g, b uint32) {
	bounds := image.Bounds()
	if startX < bounds.Min.X {
		startX = bounds.Min.X
	}
	if startY < bounds.Min.Y {
		startY = bounds.Min.Y
	}
	if endX > bounds.Max.X {
		endX = bounds.Max.X
	}
	if endY > bounds.Max.Y {
		endY = bounds.Max.Y
	}
	if startX >= endX || startY >= endY {
		return
	}

	var totalR, totalG, totalB uint32
	var count int
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			pr, pg, pb, _ := image.At(x, y).RGBA()
			totalR += pr
			totalG += pg
			totalB += pb
			count++
		}
	}
	if count == 0 {
		return
	}

	return totalR / uint32(count),
		totalG / uint32(count),
		totalB / uint32(count)
}

func imageToAscii(image image.Image, width int, ratio float64) string {
	bounds := image.Bounds()
	if width <= 0 {
		width = defaultWidth
	}
	if ratio <= 0 {
		ratio = defaultRatio
	}
	height := int(float64(width) * (float64(bounds.Dy()) / float64(bounds.Dx())) / ratio)

	var ascii strings.Builder
	ascii.Grow(width * height * ansiOverhead)

	for y := range height {
		for x := range width {
			startX := bounds.Min.X + (x*bounds.Dx())/width
			startY := bounds.Min.Y + (y*bounds.Dy())/height
			endX := bounds.Min.X + ((x+1)*bounds.Dx())/width
			endY := bounds.Min.Y + ((y+1)*bounds.Dy())/height

			r, g, b := areaSampling(image, startX, startY, endX, endY)

			ascii.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m",
				uint8(r>>8),
				uint8(g>>8),
				uint8(b>>8),
				charset[(len(charset)-1)-brightness(r, g, b)*len(charset)/256],
			))
		}
		ascii.WriteByte('\n')
	}

	return ascii.String()
}

func main() {
	root := &cobra.Command{
		Use:     "img2a <path>",
		Short:   "A command-line tool that converts an image to ASCII art",
		Version: "1.0.6",
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			path := args[0]
			width, _ := cmd.Flags().GetInt("width")
			ratio, _ := cmd.Flags().GetFloat64("ratio")

			file, err := os.Open(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			defer file.Close()

			image, _, err := image.Decode(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			ascii := imageToAscii(image, width, ratio)
			fmt.Fprint(os.Stdout, ascii)
		},
	}

	root.Flags().IntP("width", "w", defaultWidth, "width of ASCII art output")
	root.Flags().Float64P("ratio", "r", defaultRatio, "aspect ratio of ASCII art character")
	root.Flags().SortFlags = false

	_ = root.Execute()
}
