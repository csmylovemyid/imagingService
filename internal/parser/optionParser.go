package parser

import (
	"errors"
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
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return Options{}, "", errors.New("invalid URL format, must be /OPTIONS/ENCODED_URL")
	}

	optStr := parts[0]
	imageURL := parts[1]

	if decoded, err := url.PathUnescape(imageURL); err == nil {
		imageURL = decoded
	}

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
			valStr := regexp.MustCompile(`\((.*?)\)`).FindStringSubmatch(f)
			value := 1.0
			if len(valStr) == 2 {
				value, _ = strconv.ParseFloat(valStr[1], 64)
			}
			opts.Filters[name] = value

			switch name {
			case "format":
				opts.Format = strings.ToLower(valStr[1])
			case "watermark":
				opts.Watermark = valStr[1]
			case "quality":
				opts.Quality = int(value)
			}
		}
	}

	return opts, imageURL, nil
}
