package main

import (
	"fmt"
	"image"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/spf13/cobra"
)

var charset = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`\\'. "

func brightness(r, g, b uint32) int {
	return int(.2126*float64(uint8(r>>8)) + .7152*float64(uint8(g>>8)) + .0722*float64(uint8(b>>8)))
}

func imageToAscii(image image.Image, width int) (ascii string) {
	bounds := image.Bounds()
	if width == 0 {
		width = 150
	}
	height := int(float64(width) * (float64(bounds.Dy()) / float64(bounds.Dx())) * .5)

	for y := 0; y < bounds.Dy(); y += int(bounds.Dy()/height) + 1 {
		for x := 0; x < bounds.Dx(); x += int(bounds.Dx()/width) + 1 {
			r, g, b, _ := image.At(x, y).RGBA()
			ascii += fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m",
				uint8(r>>8),
				uint8(g>>8),
				uint8(b>>8),
				charset[(len(charset)-1)-brightness(r, g, b)*len(charset)/256],
			)
		}
		ascii += "\n"
	}

	return
}

func main() {
	root := &cobra.Command{
		Use:     "img2a <path>",
		Short:   "A tool that converts images to ASCII art",
		Version: "1.0.0",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			path := args[0]
			width, _ := cmd.Flags().GetInt("width")

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			image, _, err := image.Decode(file)
			if err != nil {
				return err
			}

			ascii := imageToAscii(image, width)
			fmt.Print(ascii)

			return nil
		},
	}

	root.Flags().IntP("width", "w", 150, "width of ASCII art output")
	root.Flags().SortFlags = false

	root.Execute()
}
