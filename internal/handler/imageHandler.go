package handler

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"

	"github.com/chai2010/webp"

	"imaging-service/internal/parser"
	"imaging-service/internal/processor"
	"imaging-service/pkg/utils"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "missing image path", http.StatusBadRequest)
		return
	}

	opts, imageURL, err := parser.ParseOptions(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse options: %v", err), http.StatusBadRequest)
		return
	}

	img, err := utils.FetchImage(imageURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch image: %v", err), http.StatusBadGateway)
		return
	}

	processed, err := processor.ProcessImage(img, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process image: %v", err), http.StatusInternalServerError)
		return
	}

	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "jpeg"
	}

	switch format {
	case "png":
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, processed); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode png: %v", err), http.StatusInternalServerError)
			return
		}

	case "webp":
		w.Header().Set("Content-Type", "image/webp")
		q := opts.Quality
		if q <= 0 {
			q = 75
		}
		if err := webp.Encode(w, processed, &webp.Options{Lossless: false, Quality: float32(q)}); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode webp: %v", err), http.StatusInternalServerError)
			return
		}

	case "jpg", "jpeg":
		fallthrough
	default:
		w.Header().Set("Content-Type", "image/jpeg")
		q := opts.Quality
		if q <= 0 {
			q = 75
		}
		if err := jpeg.Encode(w, processed, &jpeg.Options{Quality: q}); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode jpeg: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
