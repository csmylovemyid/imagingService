package parser

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Options struct {
	Width      int
	Height     int
	SmartCrop  bool
	Flip       bool
	CropRegion [4]int
	Filters    map[string]float64
	Format     string
	Quality    int
	Watermark  string
}

func ParseOptions(path string) (Options, string, error) {
	if path == "" {
		return Options{}, "", errors.New("empty path")
	}

	wmRe := regexp.MustCompile(`watermark\("([^"]+)"\)`)
	wmMatch := wmRe.FindStringSubmatch(path)
	var wmURL string
	if len(wmMatch) == 2 {
		wmURL = wmMatch[1]
		path = strings.Replace(path, wmMatch[0], "", 1)
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return Options{}, "", errors.New("invalid URL format, must be /OPTIONS/ENCODED_URL")
	}

	optStr := parts[0]
	imageURL := parts[1]

	if decoded, err := url.PathUnescape(imageURL); err == nil {
		imageURL = decoded
	}

	imageURL = strings.TrimSpace(imageURL)
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		imageURL = "https://" + imageURL
	}

	opts := Options{
		Quality: 75,
		Filters: make(map[string]float64),
	}

	sizeRe := regexp.MustCompile(`(-?\d+)x(-?\d+)`)
	if matches := sizeRe.FindStringSubmatch(optStr); len(matches) == 3 {
		opts.Width, _ = strconv.Atoi(matches[1])
		opts.Height, _ = strconv.Atoi(matches[2])
		if opts.Width < 0 {
			opts.Flip = true
			opts.Width = -opts.Width
		}
	}

	if strings.Contains(optStr, "smart") {
		opts.SmartCrop = true
	}

	cropRe := regexp.MustCompile(`(\d+):(\d+):(\d+):(\d+)`)
	if matches := cropRe.FindStringSubmatch(optStr); len(matches) == 5 {
		for i := 1; i <= 4; i++ {
			opts.CropRegion[i-1], _ = strconv.Atoi(matches[i])
		}
	}

	if strings.Contains(optStr, "filters:") {
		filterStr := strings.Split(optStr, "filters:")[1]
		filterParts := strings.Split(filterStr, ":")
		for _, f := range filterParts {
			if f == "" {
				continue
			}

			name := strings.Split(f, "(")[0]
			value := 1.0
			param := ""

			if name != "watermark" {
				valStr := regexp.MustCompile(`\((.*?)\)`).FindStringSubmatch(f)
				if len(valStr) == 2 {
					param = strings.TrimSpace(valStr[1])
					if v, err := strconv.ParseFloat(param, 64); err == nil {
						value = v
					}
				}
			}

			if name != "crop" {
			opts.Filters[name] = value
		}

		if name == "crop" {
			coords := strings.Split(param, ",")
			if len(coords) == 4 {
				for i := 0; i < 4; i++ {
					if v, err := strconv.Atoi(strings.TrimSpace(coords[i])); err == nil {
						opts.CropRegion[i] = v
					}
				}
			}
		}

			switch name {
			case "format":
				if param != "" {
					opts.Format = strings.ToLower(param)
				}
			case "quality":
				opts.Quality = int(value)
			}
		}
	}

	if wmURL != "" {
		opts.Watermark = wmURL
	}

	fmt.Println("DEBUG PARSER:")
	fmt.Println("ImageURL:", imageURL)
	fmt.Println("Options:", opts)

	return opts, imageURL, nil
}
