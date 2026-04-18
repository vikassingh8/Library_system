package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/library_system/internal/blob"
	"github.com/library_system/internal/utils/response"
)

// UploadImage handles multipart image uploads, stores the file in Azure Blob
// Storage, and returns the public URL.
//
//	POST /upload-image
//	Content-Type: multipart/form-data
//	Field name: "image"
func UploadImage(blobClient *blob.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const maxSize = 10 << 20 // 10 MB

		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		if err := r.ParseMultipartForm(maxSize); err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, fmt.Errorf("file too large (max 10 MB)"))
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, fmt.Errorf("field 'image' is required"))
			return
		}
		defer file.Close()

		// Validate extension
		filename := header.Filename
		lower := strings.ToLower(filename)
		if !strings.HasSuffix(lower, ".jpg") &&
			!strings.HasSuffix(lower, ".jpeg") &&
			!strings.HasSuffix(lower, ".png") &&
			!strings.HasSuffix(lower, ".gif") &&
			!strings.HasSuffix(lower, ".webp") {
			response.ApiErrorResponse(w, http.StatusBadRequest, fmt.Errorf("only jpg, png, gif, webp images are allowed"))
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to read file"))
			return
		}

		contentType := blob.ContentTypeFromFilename(filename)

		// Build a unique blob name: covers/<timestamp>_<original-name>
		ext := lower[strings.LastIndex(lower, "."):]
		blobName := fmt.Sprintf("covers/%d%s", time.Now().UnixNano(), ext)

		publicURL, err := blobClient.Upload(blobName, data, contentType)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("upload failed: %w", err))
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]string{
			"url": publicURL,
		})
	})
}
